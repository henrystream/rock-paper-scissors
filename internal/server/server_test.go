package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rock-paper-scissors/internal/models"

	"github.com/gorilla/websocket"
)

func TestMatchmaking(t *testing.T) {
	s := NewServer()
	go s.matchPlayers()

	// Start the server
	server := httptest.NewServer(http.HandlerFunc(s.handleWebSocket))
	defer server.Close()

	url := "ws" + server.URL[4:] + "/ws"

	// Player 1
	conn1, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial error for player 1: %v", err)
	}
	defer conn1.Close()

	// Read "connected" message for Player 1
	_, msg1, err := conn1.ReadMessage()
	if err != nil {
		t.Fatalf("Read error for player 1: %v", err)
	}
	var connMsg1 models.Message
	if err := json.Unmarshal(msg1, &connMsg1); err != nil {
		t.Fatalf("Unmarshal error for player 1: %v", err)
	}
	player1ID := connMsg1.PlayerID

	// Player 2
	conn2, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial error for player 2: %v", err)
	}
	defer conn2.Close()

	// Read "connected" message for Player 2
	_, msg2, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("Read error for player 2: %v", err)
	}
	var connMsg2 models.Message
	if err := json.Unmarshal(msg2, &connMsg2); err != nil {
		t.Fatalf("Unmarshal error for player 2: %v", err)
	}
	player2ID := connMsg2.PlayerID

	// Read "start_game" messages
	_, startMsg1, err := conn1.ReadMessage()
	if err != nil {
		t.Fatalf("Read error for start_game player 1: %v", err)
	}
	_, startMsg2, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("Read error for start_game player 2: %v", err)
	}

	var gameMsg1, gameMsg2 models.Message
	if err := json.Unmarshal(startMsg1, &gameMsg1); err != nil {
		t.Fatalf("Unmarshal error for player 1 start_game: %v", err)
	}
	if err := json.Unmarshal(startMsg2, &gameMsg2); err != nil {
		t.Fatalf("Unmarshal error for player 2 start_game: %v", err)
	}

	// Verify opponent IDs
	if gameMsg1.OpponentID != player2ID || gameMsg2.OpponentID != player1ID {
		t.Errorf("Mismatch in opponent IDs: got %s and %s, expected %s and %s",
			gameMsg1.OpponentID, gameMsg2.OpponentID, player2ID, player1ID)
	}
	if gameMsg1.Event != "start_game" || gameMsg2.Event != "start_game" {
		t.Errorf("Expected 'start_game' event, got %s and %s", gameMsg1.Event, gameMsg2.Event)
	}

	// Give server time to process cleanup
	time.Sleep(100 * time.Millisecond)
}
