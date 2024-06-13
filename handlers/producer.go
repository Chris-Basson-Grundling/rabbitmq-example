package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

func RunProducer(ctx *fiber.Ctx) error {
	go runProducer()
	ctx.JSON("producer running")

	return nil
}

func runProducer() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	handleError(err, "Can't connect to AMQP")
	defer conn.Close()

	amqpChannel, err := conn.Channel()
	handleError(err, "Can't create a amqpChannel")

	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare("add", true, false, false, false, nil)
	handleError(err, "Could not declare `add` queue")

	rand.Seed(time.Now().UnixNano())

	limit := 10
	var wg sync.WaitGroup
	wg.Add(limit)

	for i := 0; i < limit; i++ {
		number := i
		go func() {
			defer wg.Done()
			err = amqpChannel.Publish("", queue.Name, false, false, amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         []byte(fmt.Sprintf("Data %v", number)),
			})

			if err != nil {
				log.Fatalf("Error publishing message: %s", err)
			}
		}()
	}

	wg.Wait()
}
