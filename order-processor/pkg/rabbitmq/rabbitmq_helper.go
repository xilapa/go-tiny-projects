package rabbitmq

import strongrabbit "github.com/xilapa/go-tiny-projects/strong-rabbit"

func SetupConsumeChannel(url, connGroup, queue string) *strongrabbit.StrongChannel {
	conn, err := strongrabbit.Connect(url, connGroup)
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel(strongrabbit.Consumer, "consumer")
	if err != nil {
		panic(err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		panic(err)
	}

	_, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		panic(err)
	}

	return ch
}

func SetupProducerChannel(url, group, exchange, queue, binding string) *strongrabbit.StrongChannel {
	conn, err := strongrabbit.Connect(url, group)
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel(strongrabbit.Publisher, "producer")
	if err != nil {
		panic(err)
	}

	err = ch.Confirm(false)
	if err != nil {
		panic(err)
	}

	err = ch.ExchangeDeclare(
		exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(
		queue,
		binding,
		exchange,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	return ch
}
