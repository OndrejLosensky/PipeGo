package events

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// EventBroker manages SSE connections and broadcasts events
type EventBroker struct {
	clients map[chan string]bool
	mu      sync.RWMutex
}

// Global event broker instance
var broker = &EventBroker{
	clients: make(map[chan string]bool),
}

// GetBroker returns the global event broker
func GetBroker() *EventBroker {
	return broker
}

// Register adds a new SSE client
func (b *EventBroker) Register(client chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[client] = true
	log.Printf("📡 SSE client connected (total: %d)", len(b.clients))
}

// Unregister removes an SSE client
func (b *EventBroker) Unregister(client chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, client)
	close(client)
	log.Printf("📡 SSE client disconnected (total: %d)", len(b.clients))
}

// Broadcast sends an event to all connected clients
func (b *EventBroker) Broadcast(eventType string, data interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return
	}

	// Format SSE message
	message := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(jsonData))

	// Send to all clients
	for client := range b.clients {
		select {
		case client <- message:
			// Message sent
		default:
			// Client buffer full, skip
		}
	}

	log.Printf("📢 Broadcast event: %s to %d client(s)", eventType, len(b.clients))
}

