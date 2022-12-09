package strongrabbit

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StrongChannel encapsulates an amqp channel pointer and the
// additional data to make auto-reconnect possible.
// For the caller, it can be used as an *amqp.Channel.
type StrongChannel struct {
	*amqp.Channel
	conn        *StrongConnection
	notifyClose chan *amqp.Error
	done        chan struct{}
	stopped     chan struct{}
	opts        *ConsumeOpts
}

// Channel returns a *StrongChannel and an error.
func (conn *StrongConnection) Channel() (*StrongChannel, error) {
	ch, err := conn.Connection.Channel()
	if err != nil {
		return nil, err
	}
	notifyClose := ch.NotifyClose(make(chan *amqp.Error))
	return &StrongChannel{
		Channel:     ch,
		conn:        conn,
		notifyClose: notifyClose,
		done:        make(chan struct{}),
		stopped:     make(chan struct{}),
	}, nil
}

// ConsumeOpts is a struct that encapsulates the options to
// start consuming from a queue.
// These options are the same that  *amqp.Consume receive.
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
// The receiveds messages are sent to the out chan.
// If the connection is closed gracefully, eg.: calling ch.Close()
// or conn.Close(), it'll stopp consuming and do not reconnect.
// Any other connection error, will make the channel to recconect.
func (ch *StrongChannel) Consume(opts *ConsumeOpts, out chan amqp.Delivery) {
	// set the opts on the channel
	ch.opts = opts

	for consumeLoop := true; consumeLoop; {
		select {
		case <-ch.done:
			consumeLoop = false
			continue
		default:
			// only try to consume if the channel is open
			if ch.Channel != nil && !ch.IsClosed() {
				consumeLoop = ch.internalConsume(out)
				continue // continue to check if the chan is done
			}

			<-time.After(time.Second * 5)
			ch.reconnect()
		}
	}
	close(ch.stopped)
	log.Printf("%s: channel gracefully closed", ch.opts.Consumer)
}

// internalConsume start consuming the messages of the channel and send them
// to the out chan.
// If the connection is closed gracefully, it returns false to finalize the
// consumeLoop.
func (ch *StrongChannel) internalConsume(out chan amqp.Delivery) bool {
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
		log.Printf("%s: %s", ch.opts.Consumer, err)
		return true
	}

	for {
		select {
		case <-ch.done:
			return false
		case err := <-ch.notifyClose:
			if err == nil {
				return false
			}
			log.Printf("%s: error %s\n\treconnection will start", ch.opts.Consumer, err)
			return true
		case msg := <-msgs:
			out <- msg
		}
	}
}

// reconnect will try to reconnect and open a new channel.
// If the reconnection or channel opening fails, it'll log
// the error and return.
func (ch *StrongChannel) reconnect() {
	conn, err := Connect(ch.conn.url, ch.conn.group)
	if err != nil {
		log.Printf("%s: cannot reconnect: %s", ch.opts.Consumer, err)
		return
	}
	// replace the connection on the channel
	ch.conn = conn

	newChan, err := conn.Connection.Channel()
	if err != nil {
		log.Printf("%s: cannot open a channel: %s", ch.opts.Consumer, err)
		return
	}
	// replace the notify close chan and the underlying channel
	newNotifyClose := newChan.NotifyClose(make(chan *amqp.Error))
	ch.Channel = newChan
	ch.notifyClose = newNotifyClose
	log.Printf("%s: reconnected", ch.opts.Consumer)
}

// Close stop consuming messages and then close the channel.
// It is safe to call this method multiple times.
func (ch *StrongChannel) Close() error {
	// capture error on successive close calls, to maintain
	// its idempotent as the official one
	defer recover()

	var err error
	// only close the underlying channel, if it's not null
	if ch.Channel != nil {
		err = ch.Channel.Close()
	}

	// signalize the consume loop to stop
	close(ch.done)
	// await the consume loop stop, to return
	<-ch.stopped
	log.Printf("%s: consumer loop stopped", ch.opts.Consumer)
	return err
}
