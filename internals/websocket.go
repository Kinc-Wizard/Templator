package internals

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// WSCommand represents a WebSocket command
type WSCommand struct {
	Type string `json:"type"`
}

// Global channel for debug messages
var DebugChannel = make(chan string, 100)

// Terminal history for WebSocket clients
var TerminalHistory []string

// WebSocket upgrader
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SendDebugMessage sends a debug message to the terminal
func SendDebugMessage(message string) {
	TerminalHistory = append(TerminalHistory, message)
	select {
	case DebugChannel <- message:
	default:
		// Channel full, ignore the message
	}
}

// TerminalWSHandler handles WebSocket connections for the terminal
func TerminalWSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	defer conn.Close()

	// Send the message history
	for _, message := range TerminalHistory {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			break
		}
	}

	// Listen for incoming commands
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var cmd WSCommand
			if err := json.Unmarshal(message, &cmd); err != nil {
				continue
			}

			if cmd.Type == "clear_terminal" {
				// Clear the history
				TerminalHistory = nil
				SendDebugMessage("Terminal cleared")
			}
		}
	}()

	// Continue listening for new messages
	// Drain any queued messages to avoid duplicating startup logs (history + buffered channel)
	for {
		select {
		case <-DebugChannel:
			// discard
		default:
			goto startStream
		}
	}

startStream:
	for message := range DebugChannel {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			break
		}
	}
}
