package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xilapa/go-tiny-projects/order-processor/internal/order/infra/database"
	"github.com/xilapa/go-tiny-projects/order-processor/internal/order/usecases"
	rabbithelper "github.com/xilapa/go-tiny-projects/order-processor/pkg/rabbitmq"
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

	workerCount := 5
	for i := 1; i <= workerCount; i++ {
		go worker(out, useCase, i)
	}

	getTotalUC := usecases.NewGetTotalUseCase(repo)

	http.HandleFunc("/status", statusHandler(getTotalUC))
	http.ListenAndServe(":8080", nil)
}

func worker(deliveries <-chan amqp.Delivery, uc *usecases.CalculateFinalPriceUseCase, workerID int) {
	for msg := range deliveries {
		var cmmd usecases.OrderCommand
		if err := json.Unmarshal(msg.Body, &cmmd); err != nil {
			log.Printf("bad coded message: %s", string(msg.Body))
			msg.Ack(false)
			continue
		}
		res, err := uc.Handle(&cmmd)
		if err != nil {
			log.Printf("error processing the message: %s", err)
		}
		msg.Ack(false)
		fmt.Printf("Worker %d processed order %s\n", workerID, res.ID)
		<-time.After(time.Second * 10)
	}
}

func statusHandler(uc *usecases.GetTotalUseCase) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			w.Write([]byte("method not allowed"))
		}

		res, err := uc.Handle()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		json.NewEncoder(w).Encode(res)
	}
}
