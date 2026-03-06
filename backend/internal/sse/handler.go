package sse

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	broker *Broker
}

func NewHandler(broker *Broker) *Handler {
	return &Handler{broker: broker}
}

func (h *Handler) HandleSSE(c *fiber.Ctx) error {
	topic := c.Params("topic")
	if topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "topic is required",
		})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("Access-Control-Allow-Origin", "*")

	clientID := uuid.New().String()
	client := h.broker.Subscribe(clientID, []string{topic})
	defer h.broker.Unsubscribe(client)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"clientId\":\"%s\"}\n\n", clientID)
		_ = w.Flush()

		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case data, ok := <-client.Channel:
				if !ok {
					return
				}
				_, _ = fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
				_ = w.Flush()

			case <-heartbeat.C:
				_, _ = fmt.Fprintf(w, "event: heartbeat\ndata: {\"time\":\"%s\"}\n\n", time.Now().UTC().Format(time.RFC3339))
				_ = w.Flush()
			}
		}
	})

	return nil
}

func (h *Handler) HandleTournamentSSE(c *fiber.Ctx) error {
	tournamentID := c.Params("id")
	if tournamentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament ID is required",
		})
	}

	topic := fmt.Sprintf("tournament:%s", tournamentID)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("Access-Control-Allow-Origin", "*")

	clientID := uuid.New().String()

	client := h.broker.Subscribe(clientID, []string{topic})
	defer h.broker.Unsubscribe(client)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"clientId\":\"%s\",\"tournamentId\":\"%s\"}\n\n", clientID, tournamentID)
		_ = w.Flush()

		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case data, ok := <-client.Channel:
				if !ok {
					return
				}
				_, _ = fmt.Fprintf(w, "event: leaderboard\ndata: %s\n\n", data)
				_ = w.Flush()

			case <-heartbeat.C:
				_, _ = fmt.Fprintf(w, "event: heartbeat\ndata: {\"time\":\"%s\"}\n\n", time.Now().UTC().Format(time.RFC3339))
				_ = w.Flush()
			}
		}
	})

	return nil
}

func (h *Handler) Broker() *Broker {
	return h.broker
}
