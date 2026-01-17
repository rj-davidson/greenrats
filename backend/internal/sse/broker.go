package sse

import (
	"encoding/json"
	"sync"
)

// Event represents an SSE event to broadcast.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Client represents a connected SSE client.
type Client struct {
	ID       string
	Channel  chan []byte
	Topics   map[string]bool
	mu       sync.RWMutex
}

// Broker manages SSE client connections and message broadcasting.
type Broker struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *topicMessage
	mu         sync.RWMutex
}

type topicMessage struct {
	topic string
	data  []byte
}

// NewBroker creates a new SSE broker.
func NewBroker() *Broker {
	b := &Broker{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *topicMessage, 256),
	}

	go b.run()
	return b
}

// run processes broker events in a goroutine.
func (b *Broker) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client.ID] = client
			b.mu.Unlock()

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client.ID]; ok {
				close(client.Channel)
				delete(b.clients, client.ID)
			}
			b.mu.Unlock()

		case msg := <-b.broadcast:
			b.mu.RLock()
			for _, client := range b.clients {
				client.mu.RLock()
				subscribed := client.Topics[msg.topic]
				client.mu.RUnlock()

				if subscribed {
					select {
					case client.Channel <- msg.data:
					default:
						// Client buffer full, skip this message
					}
				}
			}
			b.mu.RUnlock()
		}
	}
}

// Subscribe adds a client to the broker.
func (b *Broker) Subscribe(clientID string, topics []string) *Client {
	topicMap := make(map[string]bool)
	for _, t := range topics {
		topicMap[t] = true
	}

	client := &Client{
		ID:      clientID,
		Channel: make(chan []byte, 64),
		Topics:  topicMap,
	}

	b.register <- client
	return client
}

// Unsubscribe removes a client from the broker.
func (b *Broker) Unsubscribe(client *Client) {
	b.unregister <- client
}

// AddTopic subscribes a client to an additional topic.
func (b *Broker) AddTopic(client *Client, topic string) {
	client.mu.Lock()
	client.Topics[topic] = true
	client.mu.Unlock()
}

// RemoveTopic unsubscribes a client from a topic.
func (b *Broker) RemoveTopic(client *Client, topic string) {
	client.mu.Lock()
	delete(client.Topics, topic)
	client.mu.Unlock()
}

// Broadcast sends an event to all clients subscribed to a topic.
func (b *Broker) Broadcast(topic string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	b.broadcast <- &topicMessage{
		topic: topic,
		data:  data,
	}
	return nil
}

// ClientCount returns the number of connected clients.
func (b *Broker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
