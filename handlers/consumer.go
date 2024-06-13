package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/http"
	"os"

	"github.com/Chris-Basson-Grundling/rabbitmq-example/models"
	"github.com/streadway/amqp"
)

func AddConsumer(ctx *fiber.Ctx) error {
	var request struct {
		Name string `json:"name" xml:"name" form:"name"`
	}
	if err := ctx.BodyParser(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, "unable to process your request")
		return err
	}

	// run a new consumer
	data := models.WebsocketDataPayload{
		Type: "consumer",
		Name: request.Name,
	}

	models.SocketDataPayload <- data

	go launchConsumer(request.Name, models.SocketDataPayload)

	ctx.JSON(http.StatusOK, "added")

	return nil
}

var goroutineMap = make(map[string]consumer)

func CloseAllConsumers() {
	for _, consumer := range goroutineMap {
		consumer.Stop()
	}
}

func GetConsumerList() map[string]consumer {
	return goroutineMap
}

// Create a map to store goroutine handlers and their corresponding done channels
type consumer struct {
	name string
	done chan bool
}

// New consumer
func NewConsumer(name string) *consumer {
	c := consumer{
		name: name,
		done: make(chan bool, 1),
	}
	goroutineMap[name] = c
	return &c
}

// Run consumer
func (c *consumer) Run(consumeChannel <-chan amqp.Delivery, responseChannel chan models.WebsocketDataPayload) {

	log.Printf("Consumer %v ready, PID: %d", c.name, os.Getpid())

	for {
		select {
		case <-c.done:
			log.Println("Consumer stopped.")
			return
		case d, ok := <-consumeChannel:
			if !ok {
				log.Println("Consumer channel closed. Exiting.")
				return
			}

			log.Printf("[%v] Received a message: %s", c.name, d.Body)

			responseChannel <- models.WebsocketDataPayload{
				Type: "consumed-data",
				Name: c.name,
				Data: string(d.Body),
			}
			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message: %s", err)
			} else {
				log.Printf("[%v] Acknowledged message", c.name)
			}
		}
	}
}

// Stop the consumer
func (c *consumer) Stop() {
	fmt.Printf("Closing consumer [%v] \n", c.name)
	c.done <- true
	delete(goroutineMap, c.name)
}

// LaunchConsumer
func launchConsumer(name string, socketDataPaylod chan models.WebsocketDataPayload) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	handleError(err, "Can't connect to AMQP")
	defer conn.Close()

	amqpChannel, err := conn.Channel()
	handleError(err, "Can't create a amqpChannel")

	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare("add", true, false, false, false, nil)
	handleError(err, "Could not declare `add` queue")

	err = amqpChannel.Qos(1, 0, false)
	handleError(err, "Could not configure QoS")

	autoAck, exclusive, noLocal, noWait := false, false, false, false

	messageChannel, err := amqpChannel.Consume(
		queue.Name,
		"",
		autoAck,
		exclusive,
		noLocal,
		noWait,
		nil,
	)
	handleError(err, "Could not register consumer")

	fmt.Println("connection success")

	consumer := NewConsumer(name)
	consumer.Run(messageChannel, socketDataPaylod)
}
