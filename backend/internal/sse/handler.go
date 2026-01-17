package sse

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles SSE connections.
type Handler struct {
	broker *Broker
}

// NewHandler creates a new SSE handler.
func NewHandler(broker *Broker) *Handler {
	return &Handler{broker: broker}
}

// HandleSSE handles SSE connections for a given topic.
func (h *Handler) HandleSSE(c *fiber.Ctx) error {
	topic := c.Params("topic")
	if topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "topic is required",
		})
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("Access-Control-Allow-Origin", "*")

	// Generate client ID
	clientID := uuid.New().String()

	// Subscribe to topic
	client := h.broker.Subscribe(clientID, []string{topic})
	defer h.broker.Unsubscribe(client)

	// Stream events
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Send initial connection event
		fmt.Fprintf(w, "event: connected\ndata: {\"clientId\":\"%s\"}\n\n", clientID)
		w.Flush()

		// Heartbeat ticker to keep connection alive
		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case data, ok := <-client.Channel:
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
				w.Flush()

			case <-heartbeat.C:
				fmt.Fprintf(w, "event: heartbeat\ndata: {\"time\":\"%s\"}\n\n", time.Now().UTC().Format(time.RFC3339))
				w.Flush()
			}
		}
	})

	return nil
}

// HandleTournamentSSE handles SSE connections for tournament updates.
func (h *Handler) HandleTournamentSSE(c *fiber.Ctx) error {
	tournamentID := c.Params("id")
	if tournamentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament ID is required",
		})
	}

	topic := fmt.Sprintf("tournament:%s", tournamentID)

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("Access-Control-Allow-Origin", "*")

	// Generate client ID
	clientID := uuid.New().String()

	// Subscribe to tournament topic
	client := h.broker.Subscribe(clientID, []string{topic})
	defer h.broker.Unsubscribe(client)

	// Stream events
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Send initial connection event
		fmt.Fprintf(w, "event: connected\ndata: {\"clientId\":\"%s\",\"tournamentId\":\"%s\"}\n\n", clientID, tournamentID)
		w.Flush()

		// Heartbeat ticker
		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case data, ok := <-client.Channel:
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: leaderboard\ndata: %s\n\n", data)
				w.Flush()

			case <-heartbeat.C:
				fmt.Fprintf(w, "event: heartbeat\ndata: {\"time\":\"%s\"}\n\n", time.Now().UTC().Format(time.RFC3339))
				w.Flush()
			}
		}
	})

	return nil
}

// Broker returns the underlying broker for external broadcasting.
func (h *Handler) Broker() *Broker {
	return h.broker
}
