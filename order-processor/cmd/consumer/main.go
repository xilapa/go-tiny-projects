package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	strongrabbit "github.com/xilapa/go-tiny-projects/strong-rabbit"
)

func main() {

	ch := SetupChannel()

	out := make(chan amqp.Delivery, 2)
	go ch.Consume(&strongrabbit.ConsumeOpts{
		Queue:     "orders",
		Consumer:  "order-consumer-go",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	}, out)

	for msg := range out {
		fmt.Println(string(msg.Body))
		msg.Ack(false)
	}
}

func SetupChannel() *strongrabbit.StrongChannel {
	conn, err := strongrabbit.Connect("amqp://guest:guest@localhost:5672/", "consume")
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		panic(err)
	}
	return ch
}
