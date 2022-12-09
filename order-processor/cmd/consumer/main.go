package main

import (
	"encoding/json"
	"log"

	_ "github.com/mattn/go-sqlite3"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xilapa/go-tiny-projects/order-processor/internal/order/infra/database"
	rabbithelper "github.com/xilapa/go-tiny-projects/order-processor/internal/order/infra/rabbitmq"
	"github.com/xilapa/go-tiny-projects/order-processor/internal/order/usecases"
	strongrabbit "github.com/xilapa/go-tiny-projects/strong-rabbit"
)

func main() {
	db, err := database.InitialiazeDb("./orders.db?_fk=on")
	if err != nil {
		panic(err)
	}
	repo := database.NewOrderRepository(db)
	useCase := usecases.NewCalculateFinalPriceUseCase(repo)

	ch := rabbithelper.SetupConsumeChannel(
		"amqp://guest:guest@localhost:5672/",
		"consume",
		"orders",
	)

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
		var cmmd usecases.OrderCommand
		if err := json.Unmarshal(msg.Body, &cmmd); err != nil {
			log.Printf("bad coded message: %s", string(msg.Body))
			msg.Ack(false)
			continue
		}
		res, err := useCase.Handle(&cmmd)
		if err != nil {
			log.Printf("error processing the message: %s", err)
		} else {
			resjson, _ := json.Marshal(res)
			log.Printf("message processed: %s", string(resjson))
		}
		msg.Ack(false)
	}
}
