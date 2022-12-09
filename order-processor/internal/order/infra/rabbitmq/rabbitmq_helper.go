package rabbitmq

import strongrabbit "github.com/xilapa/go-tiny-projects/strong-rabbit"

func SetupConsumeChannel(url, connGroup, queue string) *strongrabbit.StrongChannel {
	conn, err := strongrabbit.Connect(url, connGroup)
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

	_, err = ch.QueueDeclare(
		"orders",
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
