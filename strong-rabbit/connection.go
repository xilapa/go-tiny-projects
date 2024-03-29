package strongrabbit

import (
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	connPool = make(map[string]*StrongConnection)
	connLock sync.Mutex
)

// StrongConnection encapsulates an amqp connection pointer and the
// additional data to make auto-reconnect possible.
// For the caller, it can be used as an *amqp.Connection.
type StrongConnection struct {
	*amqp.Connection
	group string
	url   string
}

// Connect receives the rabbitmq endpoint and the connection group,
// returning a *StrongConnection and an error.
// If a connection for the given group is already made, it's returned
// without creating a new one.
//
// The connection group is used to store the connection on a pool and
// separate the connections by their usage. Making it easy to follow the
// official recommendation of having different connections to publish
// and consume, while making multiplexing channels on connections possible.
// https://pkg.go.dev/github.com/rabbitmq/amqp091-go#Channel.Consume
func Connect(url, group string) (*StrongConnection, error) {
	if conn := getConnection(group); conn != nil {
		log.Println("got a connection from pool")
		return conn, nil
	}
	connLock.Lock()
	defer connLock.Unlock()
	// check again if there is a connection after getting the lock
	if conn := getConnection(group); conn != nil {
		log.Println("got a connection from pool")
		return conn, nil
	}
	log.Println("connecting")
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	// Give the connection sometime to setup (as the official library does)
	<-time.After(time.Second)

	// Creating the strong connection
	strongConn := &StrongConnection{
		Connection: conn,
		group:      group,
		url:        url,
	}

	connPool[group] = strongConn
	log.Println("connected")
	return strongConn, nil
}

// getConnection returns the connection of a given group, if there is
// no connection or the connection is closed, nil is returned
func getConnection(group string) *StrongConnection {
	conn, ok := connPool[group]
	if !ok || conn == nil || conn.IsClosed() {
		return nil
	}
	return conn
}

// Close the connection and remove it from the internal pool.
// If the connection is already closed or nil, no error is returned.
// It is safe to call this method multiple times.
func (cn *StrongConnection) Close() error {
	if cn.Connection == nil || cn.Connection.IsClosed() {
		delete(connPool, cn.group)
		return nil
	}
	err := cn.Connection.Close()
	delete(connPool, cn.group)
	return err
}
