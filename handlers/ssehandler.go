package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Chris-Basson-Grundling/rabbitmq-example/models"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"strings"
)

func SseHandler(ctx *fiber.Ctx) error {
	CloseAllConsumers()

	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")

	fmt.Println("client connected")

	ctx.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case data := <-models.SocketDataPayload:
				fmt.Println("got a msg on socketDataPayload")
				formattedMessage, _ := formatSSEMessage("event", data)
				fmt.Fprintf(w, formattedMessage)

				err := w.Flush()
				if err != nil {
					// Refreshing page in web browser will establish a new
					// SSE connection, but only (the last) one is alive, so
					// dead connections must be closed here.
					fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)

					break
				}
			}
		}
	}))

	return nil
}

func formatSSEMessage(eventType string, data any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	err := enc.Encode(data)
	if err != nil {
		return "", nil
	}
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("event:%s\n", eventType))
	sb.WriteString(fmt.Sprintf("data:%v\n\n", buf.String()))

	return sb.String(), nil
}
