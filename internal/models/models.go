// internal/models/models.go
package models

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Player represents a connected player
type Player struct {
	ID       string
	Conn     *websocket.Conn
	Choice   string // "rock", "paper", "scissors", or empty if not chosen
	InMatch  bool
	SendChan chan []byte // Channel to send messages to the player
	Closed   bool        //Track if SendChan is closed
}

// Match represents a game between two players
type Match struct {
	Player1 *Player
	Player2 *Player
}

// Message represents WebSocket communication
type Message struct {
	Event          string `json:"event"`
	PlayerID       string `json:"playerId,omitempty"`
	OpponentID     string `json:"opponentId,omitempty"`
	Choice         string `json:"choice,omitempty"`
	Winner         string `json:"winner,omitempty"` // PlayerID or "draw"
	PlayerChoice   string `json:"playerChoice,omitempty"`
	OpponentChoice string `json:"opponentChoice,omitempty"`
}

// NewPlayer creates a player with a unique ID
func NewPlayer(conn *websocket.Conn) *Player {
	return &Player{
		ID:       uuid.New().String(),
		Conn:     conn,
		SendChan: make(chan []byte, 10),
	}
}
