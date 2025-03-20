package models

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Player represents a connected player
type Player struct {
	ID       string
	Conn     *websocket.Conn
	Choice   string
	InMatch  bool
	SendChan chan []byte
	Closed   bool
}

// Match represents a game between two players
type Match struct {
	Player1 *Player
	Player2 *Player
}

// Score represents a player's scoreboard stats
type Score struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

// Message represents WebSocket communication
type Message struct {
	Event          string           `json:"event"`
	PlayerID       string           `json:"playerId,omitempty"`
	OpponentID     string           `json:"opponentId,omitempty"`
	Choice         string           `json:"choice,omitempty"`
	Winner         string           `json:"winner,omitempty"`
	PlayerChoice   string           `json:"playerChoice,omitempty"`
	OpponentChoice string           `json:"opponentChoice,omitempty"`
	Scoreboard     map[string]Score `json:"scoreboard,omitempty"` // For scoreboard updates
}

// NewPlayer creates a player with a unique ID
func NewPlayer(conn *websocket.Conn) *Player {
	return &Player{
		ID:       uuid.New().String(),
		Conn:     conn,
		SendChan: make(chan []byte, 10),
	}
}
