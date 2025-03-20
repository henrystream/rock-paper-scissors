package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"rock-paper-scissors/internal/game"
	"rock-paper-scissors/internal/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins for simplicity
}

type Server struct {
	players    map[string]*models.Player
	queue      chan *models.Player
	matches    map[string]*models.Match
	scoreboard map[string]models.Score // Global scoreboard
	mu         sync.Mutex
}

func NewServer() *Server {
	return &Server{
		players:    make(map[string]*models.Player),
		queue:      make(chan *models.Player, 10),
		matches:    make(map[string]*models.Match),
		scoreboard: make(map[string]models.Score),
	}
}

func (s *Server) Run() {
	go s.matchPlayers()
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/scoreboard", s.handleScoreboard) // New REST endpoint
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	player := models.NewPlayer(conn)
	s.mu.Lock()
	s.players[player.ID] = player
	s.mu.Unlock()

	// Send connected message
	msg := models.Message{Event: "connected", PlayerID: player.ID}
	s.sendMessage(player, msg)

	// Handle outgoing messages
	go s.handlePlayerWrites(player)

	// Add to matchmaking queue
	s.queue <- player

	// Handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error for player %s: %v", player.ID, err)
			s.removePlayer(player)
			return
		}
		s.handleMessage(player, message)
	}
}

func (s *Server) handlePlayerWrites(player *models.Player) {
	defer s.removePlayer(player) // Ensure cleanup on exit
	for msg := range player.SendChan {
		if err := player.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("Write error for player %s: %v", player.ID, err)
			return
		}
	}
}

func (s *Server) sendMessage(player *models.Player, msg models.Message) {
	data, _ := json.Marshal(msg)
	select {
	case player.SendChan <- data:
	default:
		log.Printf("Failed to send to player %s, channel full", player.ID)
	}
}

func (s *Server) handleMessage(player *models.Player, data []byte) {
	var msg models.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Invalid message from %s: %v", player.ID, err)
		s.sendMessage(player, models.Message{
			Event: "error",
			// Use PlayerID as a placeholder for a message or detail field
			PlayerID: "Expected JSON, e.g., {\"event\": \"player_choice\", \"choice\": \"rock\"}",
		})
		return
	}

	if msg.Event == "player_choice" && player.InMatch {
		s.mu.Lock()
		player.Choice = msg.Choice
		match := s.matches[player.ID]
		s.mu.Unlock()

		if match == nil {
			return
		}

		s.processMatch(match)
	}
}

func (s *Server) matchPlayers() {
	for {
		player1 := <-s.queue
		player2 := <-s.queue

		s.mu.Lock()
		player1.InMatch = true
		player2.InMatch = true
		match := &models.Match{Player1: player1, Player2: player2}
		s.matches[player1.ID] = match
		s.matches[player2.ID] = match
		s.mu.Unlock()

		startMsg := models.Message{
			Event:      "start_game",
			OpponentID: player2.ID,
		}
		s.sendMessage(player1, startMsg)
		startMsg.OpponentID = player1.ID
		s.sendMessage(player2, startMsg)
	}
}

func (s *Server) processMatch(match *models.Match) {
	if match.Player1.Choice == "" || match.Player2.Choice == "" {
		return // Wait for both choices
	}

	winner, err := game.DetermineWinner(match.Player1.Choice, match.Player2.Choice)
	if err != nil {
		log.Printf("Game error: %v", err)
		return
	}

	// Update scoreboard
	s.mu.Lock()
	s.updateScoreboard(match, winner)
	s.mu.Unlock()

	result := models.Message{
		Event:          "game_result",
		Winner:         winner,
		PlayerChoice:   match.Player1.Choice,
		OpponentChoice: match.Player2.Choice,
	}
	s.sendMessage(match.Player1, result)

	result.PlayerChoice = match.Player2.Choice
	result.OpponentChoice = match.Player1.Choice
	s.sendMessage(match.Player2, result)

	s.mu.Lock()
	delete(s.matches, match.Player1.ID)
	delete(s.matches, match.Player2.ID)
	match.Player1.InMatch = false
	match.Player2.InMatch = false
	match.Player1.Choice = ""
	match.Player2.Choice = ""
	s.mu.Unlock()

	// Broadcast scoreboard
	s.broadcastScoreboard()

	s.queue <- match.Player1
	s.queue <- match.Player2
}

func (s *Server) updateScoreboard(match *models.Match, winner string) {
	p1Score := s.scoreboard[match.Player1.ID]
	p2Score := s.scoreboard[match.Player2.ID]

	switch winner {
	case "player1":
		p1Score.Wins++
		p2Score.Losses++
	case "player2":
		p2Score.Wins++
		p1Score.Losses++
	case "draw":
		p1Score.Draws++
		p2Score.Draws++
	}

	s.scoreboard[match.Player1.ID] = p1Score
	s.scoreboard[match.Player2.ID] = p2Score
}

func (s *Server) broadcastScoreboard() {
	s.mu.Lock()
	msg := models.Message{
		Event:      "scoreboard_update",
		Scoreboard: s.scoreboard,
	}
	data, _ := json.Marshal(msg)
	s.mu.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, player := range s.players {
		if !player.Closed {
			select {
			case player.SendChan <- data:
			default:
				log.Printf("Failed to send scoreboard to player %s, channel full", player.ID)
			}
		}
	}
}

func (s *Server) removePlayer(player *models.Player) {
	s.mu.Lock()
	defer s.mu.Unlock()

	//Already removed skip
	if _, exists := s.players[player.ID]; !exists {
		return
	}

	delete(s.players, player.ID)
	if match, exists := s.matches[player.ID]; exists {
		if match.Player1 == player {
			s.sendMessage(match.Player2, models.Message{Event: "opponent_disconnected"})
		} else {
			s.sendMessage(match.Player1, models.Message{Event: "opponent_disconnected"})
		}
		delete(s.matches, match.Player1.ID)
		delete(s.matches, match.Player2.ID)
	}
	if !player.Closed {
		close(player.SendChan)
		player.Closed = true
	}
}

func (s *Server) handleScoreboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.scoreboard); err != nil {
		log.Printf("Failed to encode scoreboard: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
