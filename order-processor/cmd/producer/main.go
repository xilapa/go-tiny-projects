package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xilapa/go-tiny-projects/order-processor/config"
	orders "github.com/xilapa/go-tiny-projects/order-processor/internal/order/entity"
	rabbithelper "github.com/xilapa/go-tiny-projects/order-processor/pkg/rabbitmq"
	strongrabbit "github.com/xilapa/go-tiny-projects/strong-rabbit"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	ch := rabbithelper.SetupProducerChannel(
		cfg.RabbitMq.Url,
		"producer",
		"order-processor",
		"orders",
		"orders",
	)

	for producerLoop := true; producerLoop; {
		if ch.Done {
			producerLoop = false
			continue
		}
		err := Publish(ch, GenerateOrder())
		if err != nil {
			fmt.Printf("error while publishing order: %s", err)
			<-time.After(time.Second * 5)
			continue
		}
		fmt.Println("message published")
		<-time.After(time.Second * 5)
	}
	fmt.Println("producer program finalized")
}

func Publish(ch *strongrabbit.StrongChannel, order *orders.Order) error {
	body, err := json.Marshal(*order)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	confirm, err := ch.PublishWithDeferredConfirmWithContext(
		ctx,
		"order-processor",
		"orders",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return err
	}

	published := confirm.Wait()
	if !published {
		return errors.New("not confirmed")
	}

	return nil
}

func GenerateOrder() *orders.Order {
	order, err := orders.NewOrder(
		uuid.New().String(),
		rand.Float64()*10,
		rand.Float64()*2)
	if err != nil {
		panic(err)
	}
	return order
}
