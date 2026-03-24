package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

type NewMove struct {
	Position Position    `json:"position"`
	Piece    json.Number `json:"piece"`
}

type Move struct {
	Position Position `json:"position"`
	Piece    uint8    `json:"piece"`
}

type Server struct {
	serverMutex  sync.Mutex
	games        map[string]*Game
	waitingGames map[string]*Game
}

type Game struct {
	ID        string                     `json:"id"`
	Players   map[string]*websocket.Conn `json:"players"`
	GameState *GameState                 `json:"gameState"`
	mutex     sync.Mutex
	send      chan Message
	done      chan struct{} // signals all goroutines to stop
	closeOnce sync.Once    // ensures done is closed exactly once
}

type Message struct {
	Type     string      `json:"type"`
	GameID   string      `json:"gameID"`
	PlayerID string      `json:"playerID"`
	State    string      `json:"state"`
	Payload  interface{} `json:"payload"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origins := map[string]bool{
			os.Getenv("ORIGIN_URL"): true,
		}
		return origins[r.Header.Get("Origin")]
	},
}

func NewServer() *Server {
	return &Server{
		games:        make(map[string]*Game),
		waitingGames: make(map[string]*Game),
	}
}

func NewGame() *Game {
	return &Game{
		ID:        generateGameID(),
		GameState: NewGameState(),
		Players:   make(map[string]*websocket.Conn),
		send:      make(chan Message, 16), // buffered to prevent blocking
		done:      make(chan struct{}),
	}
}

func (game *Game) shutdown() {
	game.closeOnce.Do(func() {
		close(game.done)
	})
}

func (server *Server) createGame(conn *websocket.Conn) *Game {
	server.serverMutex.Lock()
	defer server.serverMutex.Unlock()

	game := NewGame()
	// Check for ID collision
	for _, exists := server.games[game.ID]; exists; _, exists = server.games[game.ID] {
		game.ID = generateGameID()
	}
	game.Players["player1"] = conn
	server.games[game.ID] = game
	server.waitingGames[game.ID] = game
	log.Printf("Game created: %s", game.ID)
	return game
}

func (server *Server) joinGame(conn *websocket.Conn, requestedGameID string) *Game {
	server.serverMutex.Lock()
	defer server.serverMutex.Unlock()

	if requestedGameID == "" {
		return nil
	}

	game, exists := server.waitingGames[requestedGameID]
	if !exists {
		return nil
	}
	if len(game.Players) >= 2 {
		return nil
	}

	playerID := fmt.Sprintf("player%d", len(game.Players)+1)
	game.Players[playerID] = conn
	delete(server.waitingGames, game.ID)
	return game
}

// writePump handles all writes to all players for a game.
// One writePump per game (not per player).
func (game *Game) writePump(s *Server, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in writePump for game %s: %v", game.ID, r)
		}
	}()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-game.done:
			return

		case msg, ok := <-game.send:
			if !ok {
				return
			}

			// Snapshot players under lock, then write without lock
			game.mutex.Lock()
			players := make(map[string]*websocket.Conn, len(game.Players))
			for id, conn := range game.Players {
				players[id] = conn
			}
			game.mutex.Unlock()

			if msg.Type == "error" {
				// Error messages go to the specific player (if set) or all
				for _, conn := range players {
					conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := conn.WriteJSON(msg); err != nil {
						log.Printf("Failed to write error message: %v", err)
					}
				}
			}

			if msg.Type == "gameState" {
				for playerID, conn := range players {
					outMsg := msg
					outMsg.PlayerID = playerID
					conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := conn.WriteJSON(outMsg); err != nil {
						log.Printf("Failed to broadcast to %s: %v", playerID, err)
					}
				}
			}

		case <-ticker.C:
			game.mutex.Lock()
			players := make(map[string]*websocket.Conn, len(game.Players))
			for id, conn := range game.Players {
				players[id] = conn
			}
			game.mutex.Unlock()

			for playerID, conn := range players {
				log.Printf("Sending ping to %s", playerID)
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteJSON(Message{Type: "ping"}); err != nil {
					log.Printf("Failed to write ping to %s: %v", playerID, err)
				}
			}
		}
	}
}

func (game *Game) readMove(conn *websocket.Conn, playerID string) (error, Message, *NewMove) {
	var newMove NewMove
	errMsg := Message{
		Type:    "error",
		GameID:  game.ID,
		Payload: "Failed to read move",
		State:   game.GameState.State,
	}

	if err := conn.ReadJSON(&newMove); err != nil {
		if websocket.IsCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			log.Printf("WebSocket closed for %s: %v", playerID, err)
		} else if websocket.IsUnexpectedCloseError(err) {
			log.Printf("Unexpected WebSocket close for %s: %v", playerID, err)
		} else {
			log.Printf("Read error from %s: %v", playerID, err)
		}
		errMsg.Payload = "Disconnected"
		return err, errMsg, nil
	}

	// Piece 99 = client pong
	decodedPiece, _ := newMove.Piece.Int64()
	if decodedPiece == 99 {
		log.Printf("Pong received from %s", playerID)
		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Printf("Failed to set read deadline for %s: %v", playerID, err)
			errMsg.Payload = "Disconnected"
			return fmt.Errorf("failed to set read deadline"), errMsg, nil
		}
		return nil, Message{Type: "Pong", Payload: "Pong received"}, &newMove
	}

	if !game.isValidTurn(playerID) {
		errMsg.Payload = "Not your turn"
		// Non-blocking send
		select {
		case game.send <- errMsg:
		default:
			log.Printf("Send channel full, dropping 'not your turn' error for %s", playerID)
		}
		log.Printf("Not your turn %s", playerID)
		return fmt.Errorf("not your turn"), errMsg, &newMove
	}

	return nil, Message{}, &newMove
}

const (
	pingPeriod = 30 * time.Second
	pongWait   = 60 * time.Second
	writeWait  = 10 * time.Second
)

func (game *Game) readPump(conn *websocket.Conn, playerID string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in readPump for game %s player %s: %v", game.ID, playerID, r)
		}
	}()

	conn.SetReadDeadline(time.Now().Add(pongWait))

	log.Printf("ReadPump: Player %s connected to game %s", playerID, game.ID)
	for {
		// Check if game is shutting down
		select {
		case <-game.done:
			return
		default:
		}

		switch game.GameState.State {
		case "WAITING":
			err, errMsg, newMove := game.readMove(conn, playerID)
			if err != nil || errMsg.Type == "Pong" {
				if errMsg.Payload == "Disconnected" {
					return
				}
				continue
			}
			if err := game.processTurn(newMove); err != nil {
				errMsg.Payload = err.Error()
				errMsg.State = game.GameState.State
				select {
				case game.send <- errMsg:
				default:
				}
				continue
			}

		case "MULTIPLE_WAITING":
			err, errMsg, newMove := game.readMove(conn, playerID)
			if err != nil || errMsg.Type == "Pong" {
				if errMsg.Payload == "Disconnected" {
					return
				}
				continue
			}
			if err := game.handleMultipleGraduations(newMove); err != nil {
				errMsg.Payload = err.Error()
				errMsg.State = game.GameState.State
				select {
				case game.send <- errMsg:
				default:
				}
				continue
			}

		case "MAX_WAITING":
			err, errMsg, newMove := game.readMove(conn, playerID)
			if err != nil || errMsg.Type == "Pong" {
				if errMsg.Payload == "Disconnected" {
					return
				}
				continue
			}
			if err := game.handleMaxedOutGraduation(newMove); err != nil {
				errMsg.Payload = err.Error()
				errMsg.State = game.GameState.State
				select {
				case game.send <- errMsg:
				default:
				}
				continue
			}
		}

		if game.GameState.State == "WAITING" {
			game.GameState.TurnNumber++
		}
		game.broadcastGameState()
	}
}

func (s *Server) handleGameLoop(conn *websocket.Conn, game *Game, playerID string) {
	defer func() {
		log.Printf("handleGameLoop ending for %s player %s", game.ID, playerID)
		conn.Close()
		s.handlePlayerDisconnect(game.ID, playerID)
	}()

	var wg sync.WaitGroup

	// readPump per player, writePump per game (started once by first player)
	wg.Add(1)
	go game.readPump(conn, playerID, &wg)

	wg.Wait()
}

func (game *Game) isValidTurn(playerID string) bool {
	return (game.GameState.isPlayer1() && playerID == "player1") ||
		(!game.GameState.isPlayer1() && playerID == "player2")
}

func (game *Game) processTurn(newMove *NewMove) error {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	game.GameState.PreviousBoard = game.GameState.Board
	game.GameState.GraduatedLine = nil

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
	game.GameState.Placed = Move{Position: newMove.Position, Piece: piece}
	game.GameState.calculateOriginal()

	if len(game.GameState.Lines) > 1 {
		game.GameState.State = "MULTIPLE_WAITING"
	} else if len(game.GameState.Lines) == 1 {
		game.GameState.Board.graduatePieces(game.GameState.Lines[0], game.GameState)
	} else {
		if game.shouldCheckMaxedOut() {
			if game.GameState.Board.winCheckMaxCats(game.GameState) {
				return nil
			}
			game.GameState.State = "MAX_WAITING"
		}
	}

	return nil
}

func (game *Game) shouldCheckMaxedOut() bool {
	return (game.GameState.isPlayer1() && game.GameState.P1.Placed == 8) ||
		(!game.GameState.isPlayer1() && game.GameState.P2.Placed == 8)
}

func (game *Game) handleMultipleGraduations(selection *NewMove) error {
	log.Printf("handleMultipleGrad: Called, selection: %+v", selection)
	log.Println(game.GameState.ThreeChoices)
	if !slices.Contains(game.GameState.ThreeChoices, selection.Position) {
		return fmt.Errorf("invalid graduation selection: position is not a valid choice")
	}

	line := game.GameState.getLineContainingPosition(selection.Position)
	if line == nil {
		return fmt.Errorf("invalid graduation selection: no complete line found at position")
	}

	game.GameState.Board.graduatePieces(line, game.GameState)
	log.Println("handleMultipleGrad: changing State to WAITING")
	game.GameState.State = "WAITING"
	game.GameState.GraduatedLine = line
	game.GameState.Lines = nil
	game.GameState.BoopMovement = nil
	game.GameState.Booped = nil
	game.GameState.Placed = Move{}
	return nil
}

func (game *Game) handleMaxedOutGraduation(selection *NewMove) error {
	log.Println("handleMaxedGrad: Called")
	playerPieces := game.GameState.Board.getPlayerPiecePositions(game.GameState)
	if !slices.Contains(playerPieces, selection.Position) {
		log.Println("handleMaxedGrad: Position not found in player pieces")
		return fmt.Errorf("invalid graduation selection: position is not a valid piece")
	}

	game.GameState.Board.graduatePiece(selection.Position, game.GameState)
	log.Println("handleMaxedGrad: changing State to WAITING")
	game.GameState.State = "WAITING"
	game.GameState.GraduatedLine = []Position{selection.Position}
	game.GameState.Lines = nil
	game.GameState.BoopMovement = nil
	game.GameState.Booped = nil
	game.GameState.Placed = Move{}

	return nil
}

func (game *Game) broadcastGameState() {
	game.GameState.BroadcastSeq++
	log.Printf("Broadcasting game state: %s", game.GameState.State)
	stateMsg := Message{
		Type:    "gameState",
		GameID:  game.ID,
		Payload: game.GameState,
		State:   game.GameState.State,
	}

	// Non-blocking send — if channel is full, log and skip
	select {
	case game.send <- stateMsg:
	default:
		log.Printf("WARNING: send channel full for game %s, dropping broadcast", game.ID)
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
	delete(game.Players, playerID)
	remaining := len(game.Players)
	game.mutex.Unlock()

	if waitingGame, exists := s.waitingGames[gameID]; exists {
		delete(waitingGame.Players, playerID)
		delete(s.waitingGames, gameID)
	}

	// Clean up game when no players remain
	if remaining == 0 {
		game.shutdown()
		delete(s.games, gameID)
		log.Printf("Game %s cleaned up (no players remaining)", gameID)
	}
}

func generateGameID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", os.Getenv("ORIGIN_URL"))
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	// Log to both stderr and a persistent file so logs survive container restarts
	dbPath := os.Getenv("DB_PATH")
	logDir := "/data"
	if dbPath != "" {
		logDir = dbPath[:len(dbPath)-len("/games.db")]
	}
	logFile, err := os.OpenFile(logDir+"/backend.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(io.MultiWriter(os.Stderr, logFile))
		defer logFile.Close()
	}

	server := NewServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.handleConnection)
	mux.HandleFunc("/getWaitingGame", server.handleGetWaitingGameID)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Graceful shutdown on SIGTERM/SIGINT
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Println("Server starting on :8080")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Println("Server stopped")
}
