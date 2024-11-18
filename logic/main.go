package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
)

func LoadTestGameState() *GameState {
	gameState := NewGameState()

	gameState.Board = Board{
		{1, 8, 8, 0, 8, 8},
		{0, 1, 2, 0, 1, 0},
		{2, 0, 1, 0, 0, 0},
		{2, 0, 8, 0, 2, 0},
		{0, 0, 8, 0, 8, 8},
		{8, 0, 0, 0, 0, 0},
	}

	gameState.P1.Cats = 8
	gameState.P1.Kittens = 8

	return gameState
}

type NewMove struct {
	Position Position    `json:"position"`
	Piece    json.Number `json:"piece"`
}

type Server struct {
	serverMutex sync.Mutex
	games       map[string]*Game
	waitingGame *Game
}

type Game struct {
	ID        string                     `json:"id"`
	Players   map[string]*websocket.Conn `json:"players"`
	GameState *GameState                 `json:"gameState"`
	mutex     sync.Mutex
}

type Message struct {
	Type     string      `json:"type"`
	GameID   string      `json:"gameId"`
	PlayerID string      `json:"playerId"`
	Payload  interface{} `json:"payload"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origins := map[string]bool{
			"http://localhost:5173": true,
			"http://127.0.0.1:5173": true,
		}
		return origins[r.Header.Get("Origin")]
	},
}

func NewServer() *Server {
	return &Server{
		games: make(map[string]*Game),
	}
}

func NewGame() *Game {
	return &Game{
		ID:        generateGameID(),
		GameState: NewGameState(),
		Players:   make(map[string]*websocket.Conn),
	}
}

func (gs *Server) assignPlayerToGame(conn *websocket.Conn, requestedGameID string) (*Game, string) {
	gs.serverMutex.Lock()
	defer gs.serverMutex.Unlock()

	// If specific game requested
	if requestedGameID != "" {
		if game, exists := gs.games[requestedGameID]; exists {
			if len(game.Players) < 2 {
				playerID := fmt.Sprintf("player%d", len(game.Players)+1)
				game.Players[playerID] = conn
				return game, playerID
			}
			return nil, "" // Game full
		}
		return nil, "" // Game not found
	}

	// Join or create waiting game
	if gs.waitingGame == nil {
		gs.waitingGame = NewGame()
		gs.games[gs.waitingGame.ID] = gs.waitingGame
	}

	game := gs.waitingGame
	playerID := fmt.Sprintf("player%d", len(game.Players)+1)
	game.Players[playerID] = conn

	// If game is now full, clear waiting game
	if len(game.Players) == 2 {
		gs.waitingGame = nil
	}

	return game, playerID
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	game, playerID := s.assignPlayerToGame(conn, r.URL.Query().Get("gameId"))
	if game == nil {
		conn.WriteJSON(Message{Type: "error", Payload: "Could not join game"})
		conn.Close()
		return
	}

	// Send initial game state
	if err := conn.WriteJSON(Message{
		Type:     "joined",
		GameID:   game.ID,
		PlayerID: playerID,
		Payload:  game.GameState,
	}); err != nil {
		s.handlePlayerDisconnect(game.ID, playerID)
		return
	}

	// Start game loop
	s.handleGameLoop(conn, game, playerID)
}

func (s *Server) handleGameLoop(conn *websocket.Conn, game *Game, playerID string) {
	defer func() {
		s.handlePlayerDisconnect(game.ID, playerID)
		conn.Close()
	}()

	for {
		var newMove NewMove
		if err := conn.ReadJSON(&newMove); err != nil {
			log.Printf("Read error from %s: %v", playerID, err)
			return
		}

		if !s.isValidTurn(game, playerID) {
			conn.WriteJSON(Message{
				Type:    "error",
				GameID:  game.ID,
				Payload: "Not your turn",
			})
			continue // Continue waiting for valid turn instead of disconnecting
		}

		if err := s.processTurn(conn, game, &newMove); err != nil {
			conn.WriteJSON(Message{
				Type:    "error",
				GameID:  game.ID,
				Payload: err.Error(),
			})
			continue
		}

		// Broadcast updated state to all players
		s.broadcastGameState(game, false)
	}
}

func (s *Server) isValidTurn(game *Game, playerID string) bool {
	return (game.GameState.isPlayer1() && playerID == "player1") ||
		(!game.GameState.isPlayer1() && playerID == "player2")
}

func (s *Server) processTurn(conn *websocket.Conn, game *Game, newMove *NewMove) error {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	kittenOrCat, _ := newMove.Piece.Int64()

	var piece uint8
	if game.GameState.isPlayer1() {
		if kittenOrCat == 0 {
			piece = 1
		} else {
			piece = 2
		}
	} else {
		if kittenOrCat == 0 {
			piece = 8
		} else {
			piece = 9
		}
	}

	if err := game.GameState.Board.move(newMove.Position, piece, game.GameState); err != nil {
		return fmt.Errorf("invalid move: %w", err)
	}
	s.broadcastGameState(game, true)

	// Handle graduation logic
	if len(game.GameState.Lines) > 1 {
		game.GameState.Waiting = true
		if err := s.handleMultipleGraduations(conn, game); err != nil {
			return err
		}
	} else if len(game.GameState.Lines) == 1 {
		game.GameState.Board.graduatePieces(game.GameState.Lines[0], game.GameState)
		game.GameState.TurnNumber++
	} else {
		if s.shouldCheckMaxedOut(game) {
			if game.GameState.Board.winCheckMaxCats(game.GameState) {
				return nil
			}
			game.GameState.Waiting = true
			if err := s.handleMaxedOutGraduation(conn, game); err != nil {
				return err
			}
		}
		game.GameState.TurnNumber++
	}

	return nil
}

func (s *Server) shouldCheckMaxedOut(game *Game) bool {
	return (game.GameState.isPlayer1() && game.GameState.P1.Placed == 8) ||
		(!game.GameState.isPlayer1() && game.GameState.P2.Placed == 8)
}

func (s *Server) handleMultipleGraduations(conn *websocket.Conn, game *Game) error {
	for game.GameState.Waiting {
		var selection NewMove
		if err := conn.ReadJSON(&selection); err != nil {
			return fmt.Errorf("failed to read graduation selection: %w", err)
		}

		if !slices.Contains(game.GameState.ThreeChoices, selection.Position) {
			continue
		}

		line := game.GameState.getLineContainingPosition(selection.Position)
		if line == nil {
			continue
		}

		game.GameState.Board.graduatePieces(line, game.GameState)
		game.GameState.TurnNumber++
		game.GameState.Waiting = false
	}
	return nil
}

func (s *Server) handleMaxedOutGraduation(conn *websocket.Conn, game *Game) error {
	for game.GameState.Waiting {
		var selection NewMove
		if err := conn.ReadJSON(&selection); err != nil {
			return fmt.Errorf("failed to read maxed out graduation selection: %w", err)
		}

		playerPieces := game.GameState.Board.getPlayerPiecePositions(game.GameState)
		if !slices.Contains(playerPieces, selection.Position) {
			continue
		}

		game.GameState.Board.graduatePiece(selection.Position, game.GameState)
		game.GameState.Waiting = false
	}
	return nil
}

func (s *Server) broadcastGameState(game *Game, alreadyLocked bool) {

	if !alreadyLocked {
		game.mutex.Lock()
		defer game.mutex.Unlock()
	}

	stateMsg := Message{
		Type:    "gameState",
		GameID:  game.ID,
		Payload: game.GameState,
	}

	for playerID, conn := range game.Players {
		stateMsg.PlayerID = playerID
		if err := conn.WriteJSON(stateMsg); err != nil {
			log.Printf("Failed to broadcast to %s: %v", playerID, err)
			s.handlePlayerDisconnect(game.ID, playerID)
		}
	}
}

func (s *Server) handlePlayerDisconnect(gameID string, playerID string) {
	s.serverMutex.Lock()
	defer s.serverMutex.Unlock()

	game, exists := s.games[gameID]
	if !exists {
		return
	}

	game.mutex.Lock()
	defer game.mutex.Unlock()

	delete(game.Players, playerID)

	if len(game.Players) == 0 {
		delete(s.games, gameID)
		if s.waitingGame == game {
			s.waitingGame = nil
		}
		return
	}

	game.GameState.Winner = 0
	s.broadcastGameState(game, true)
}

// Helper functions remain the same
func generateGameID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	server := NewServer()
	http.HandleFunc("/ws", server.handleConnection)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
