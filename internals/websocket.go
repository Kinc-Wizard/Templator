package internals

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WSCommand represents a WebSocket command
type WSCommand struct {
	Type string `json:"type"`
}

const maxTerminalHistory = 500

// Global channel for debug messages
var DebugChannel = make(chan string, 100)

// Terminal history for WebSocket clients
var TerminalHistory []string
var terminalMu sync.Mutex

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
	terminalMu.Lock()
	TerminalHistory = append(TerminalHistory, message)
	if len(TerminalHistory) > maxTerminalHistory {
		TerminalHistory = TerminalHistory[len(TerminalHistory)-maxTerminalHistory:]
	}
	terminalMu.Unlock()
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
	terminalMu.Lock()
	history := make([]string, len(TerminalHistory))
	copy(history, TerminalHistory)
	terminalMu.Unlock()

	for _, message := range history {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
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
				terminalMu.Lock()
				TerminalHistory = nil
				terminalMu.Unlock()
				SendDebugMessage("Terminal cleared")
			}
		}
	}()

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
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			break
		}
	}
}
