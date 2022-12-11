package strongrabbit

import (
	"errors"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StrongChannel encapsulates an amqp channel pointer and the
// additional data to make auto-reconnect possible.
// For the caller, it can be used as an *amqp.Channel.
type StrongChannel struct {
	*amqp.Channel                      // The underlying *amqp.Channel
	Done             bool              // Indicates if the channel has been stopped
	Name             string            // Name used for logs
	conn             *StrongConnection // the underlying connection
	notifyClose      chan *amqp.Error  // used to listen to close notifications
	consLoopStop     chan struct{}     // closes to signal the consumer loop to stop, after that it'll be nil
	consLoopStopped  chan struct{}     // used by the consumer loop to signal it has stopped
	reconnectStop    chan struct{}     // used to signal the reconnection go routine to stop
	reconnectStopped chan struct{}     // used by the reconnection go routine to signal it has stopped
	opts             *ConsumeOpts      // the opts used on consume
	err              *amqp.Error       // the last error happened on this channels, it resets when reconnected
	confirm          bool              // save the confirm mode set to the channel
	confirmNoWait    bool              // save the noWait used when setting the confirm mode
	chType           ChannelType       // the underlying channel type
	lock             sync.Mutex        // mutex used when closing the channel
}

type ChannelType int

const (
	Consumer ChannelType = iota + 1
	Publisher
)

var (
	errInvalidChannelType = errors.New("invalid channel type")
	errAlreadyConsuming   = errors.New("channel is already consuming messages")
	errNilConsumeOpts     = errors.New("ConsumeOpts is nil")
)

// Channel receives the channel type and returns a *StrongChannel
// and an error.
// Channel can be of Consumer or Publisher type. Each one has an
// optimized reconnection strategy.
// To stop consuming or publishing, call the Close() method.
func (conn *StrongConnection) Channel(t ChannelType, name string) (*StrongChannel, error) {
	// validate if the channel type is valid
	if t != Consumer && t != Publisher {
		return nil, errInvalidChannelType
	}

	ch, err := conn.Connection.Channel()
	if err != nil {
		return nil, err
	}

	notifyClose := ch.NotifyClose(make(chan *amqp.Error))
	strongCh := &StrongChannel{
		Channel:     ch,
		conn:        conn,
		notifyClose: notifyClose,
		chType:      t,
		Name:        name,
	}

	if strongCh.chType == Consumer {
		// the consumer doesn't need a go routine to reconnect
		// the internalConsume listens to the messages,
		// the amqp.Close method and the stop signal
		strongCh.consLoopStop = make(chan struct{})
		strongCh.consLoopStopped = make(chan struct{})
	}

	if strongCh.chType == Publisher {
		strongCh.reconnectStop = make(chan struct{})
		strongCh.reconnectStopped = make(chan struct{})
		// if the channel is a producer, start a go routine
		// to reconnect it, if needed
		go strongCh.reconnectionLoop()
	}

	return strongCh, nil
}

// ConsumeOpts is a struct that encapsulates the options to
// start consuming from a queue.
// These options are the same that amqp.Consume method receive.
type ConsumeOpts struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

// Consume starts consuming messages from a channel with the given options.
// The received messages are sent to the out chan.
// If there is an error on the opts passed, it's returned.
// If the connection is closed gracefully, eg.: by calling Close() on the
// connection or channel, it'll stop consuming and not reconnect.
// Any other connection error will make the channel reconnect.
//
// To stop consuming Channel.Close() should be called on another go-routine.
// When the channel stops this method returns nil.
//
// This method can only be called once per consumer channel.
func (ch *StrongChannel) Consume(opts *ConsumeOpts, out chan amqp.Delivery) error {
	if ch.opts != nil {
		return errAlreadyConsuming
	}

	if opts == nil {
		return errNilConsumeOpts
	}

	log.Printf("[%s] listening for messages", ch.Name)
	// set the opts on the channel
	ch.opts = opts

	for consLopp := true; consLopp; {
		select {
		case <-ch.consLoopStop:
			consLopp = false
			continue
		default:
			// only try to consume if the channel is open
			if ch.Channel != nil && !ch.IsClosed() {
				consLopp = ch.internalConsume(out)
				continue // continue to check if the chan is done
			}

			// reconnection delay
			<-time.After(time.Second * 5)
			if ok := ch.reconnect(); ok {
				log.Printf("[%s] reconnected", ch.Name)
			}
		}
	}

	close(ch.consLoopStopped)
	log.Printf("[%s] consumer loop stopped, channel gracefully closed", ch.Name)
	return nil
}

// internalConsume starts consuming the messages from the channel
// and sends them to the out chan.
// If notifyClose returns a nil error (graceful channel close),
// keepConsuming returns false, to avoid reconnections
func (ch *StrongChannel) internalConsume(out chan amqp.Delivery) (keepConsuming bool) {
	keepConsuming = true
	msgs, err := ch.Channel.Consume(
		ch.opts.Queue,
		ch.opts.Consumer,
		ch.opts.AutoAck,
		ch.opts.Exclusive,
		ch.opts.NoLocal,
		ch.opts.NoWait,
		ch.opts.Args,
	)

	if err != nil {
		log.Printf("[%s] error: %s", ch.Name, err)
		return
	}

	for {
		select {
		case err := <-ch.notifyClose:
			// when the connection is closed gracefully no error is returned
			// https://pkg.go.dev/github.com/rabbitmq/amqp091-go#hdr-Use_Case
			if err == nil {
				keepConsuming = false
				return
			}
			// when an error happens the amqp library sends it and closes the
			// notifyClose chan. Closed chan returns nil immediately, due to
			// this, at Consume method it's verified if the consLoopStop chan
			// still open to do a reconnect
			log.Printf("[%s] reconnection will start\nerror: %s", ch.Name, err)
			return
		case msg := <-msgs:
			// if the channel closes due to an error
			// the msg.Body and other fields came empty/nil
			if msg.Body != nil {
				out <- msg
			}
		}
	}
}

// reconnectionLoop listens to channel close notifications and
// reconnect the channel until it's gracefully closed
func (ch *StrongChannel) reconnectionLoop() {
	log.Printf("[%s] listening for channel close", ch.Name)
	for {
		select {
		case <-ch.reconnectStop:
			close(ch.reconnectStopped)
			log.Printf("[%s] reconnection loop stopped, connection closed gracefully", ch.Name)
			return

		case err := <-ch.notifyClose:
			// when the connection/channel closes, this chan
			// will also be closed and will always return nil
			// after that. So it's necessary to check if the
			// ch.err has a value
			if err == nil && ch.err == nil {
				continue
			}

			// save the error on the channel to keep trying the reconnection
			// when reconnected, the ch.err will be set to nil
			if err != nil {
				ch.err = err
				log.Printf("[%s] channel closed: %s", ch.Name, ch.err)
			}

			// reconnection delay
			<-time.After(time.Second * 5)

			// stop the reconnection loop, if reconnection succeeds
			if res := ch.reconnect(); res {
				log.Printf("[%s] reconnected", ch.Name)
				return
			}
		}
	}
}

// reconnect will try to reconnect and open a new channel.
// If the reconnection or channel opening fails, it'll log
// the error and returns false.
// If the reconection succeeds, it returns true.
func (ch *StrongChannel) reconnect() (success bool) {
	conn, err := Connect(ch.conn.url, ch.conn.group)
	if err != nil {
		log.Printf("[%s] cannot reconnect: %s", ch.Name, err)
		return
	}
	// replace the connection on the channel
	ch.conn = conn

	newChan, err := conn.Connection.Channel()
	if err != nil {
		log.Printf("[%s] cannot open a channel: %s", ch.Name, err)
		return
	}

	// if channel was in confirm mode, re-set it
	if ch.confirm {
		err = newChan.Confirm(ch.confirmNoWait)
		if err != nil {
			log.Printf("[%s] cannot put channel in confirm mode: %s", ch.Name, err)
			return
		}
	}

	// replace the notify close chan and the underlying channel
	ch.Channel = newChan
	ch.notifyClose = newChan.NotifyClose(make(chan *amqp.Error))

	// clear the error on the channel
	ch.err = nil

	// start another reconnection loop if it's a producer channel
	if ch.chType == Publisher {
		go ch.reconnectionLoop()
	}

	return true
}

// Close stop consuming messages and then close the channel.
// It is safe to call this method multiple times.
func (ch *StrongChannel) Close() error {
	// capture error on successive close calls, to maintain
	// its idempotent as the official one
	// defer recover()
	ch.lock.Lock()
	defer ch.lock.Unlock()

	var err error
	// only close the underlying channel, if it's not null
	if ch.Channel != nil {
		err = ch.Channel.Close()
	}

	if ch.consLoopStop != nil {
		// signalize the consume loop to stop
		close(ch.consLoopStop)

		// await the consume loop stop
		<-ch.consLoopStopped

		// set the chan to nil, to avoid closing it again
		ch.consLoopStop = nil
		ch.consLoopStopped = nil
	}

	if ch.reconnectStop != nil {
		// signalize the reconnection go routine to stop
		close(ch.reconnectStop)

		// await it stop
		<-ch.reconnectStopped

		// set the chan to nil, to avoid closing it again
		ch.reconnectStop = nil
		ch.reconnectStopped = nil
	}

	// indicates for the caller that the channel stopped
	ch.Done = true

	// release resources
	ch.notifyClose = nil
	ch.err = nil
	ch.opts = nil

	log.Printf("[%s] channel stopped", ch.Name)
	return err
}

// Confirm puts the channel on Confirm mode. Only Publisher
// channels can have confirm mode set.
// On reconnection, confirm mode is restored.
// For more information on Channel.Confirm check the amqp docs:
// https://pkg.go.dev/github.com/rabbitmq/amqp091-go#Channel.Confirm
func (ch *StrongChannel) Confirm(noWait bool) error {
	if ch.chType != Publisher {
		return errInvalidChannelType
	}
	// save the confirm mode to use on reconnection
	ch.confirm = true
	ch.confirmNoWait = noWait
	return ch.Channel.Confirm(noWait)
}
