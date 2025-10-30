package api

import (
	"fmt"
	"net/http"

	"pipego/events"
)

// SSEHandler handles Server-Sent Events connections
func SSEHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Create client channel
		client := make(chan string, 10) // Buffer to prevent blocking
		broker := events.GetBroker()
		broker.Register(client)
		defer broker.Unregister(client)

		// Send initial connection message
		fmt.Fprintf(w, "event: connected\ndata: {\"message\": \"Connected to PipeGo events\"}\n\n")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		// Keep connection open and send events
		for {
			select {
			case message := <-client:
				// Send event to client
				fmt.Fprint(w, message)
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			case <-r.Context().Done():
				// Client disconnected
				return
			}
		}
	}
}
