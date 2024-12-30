package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"
)

func LoadTestGameState() *GameState {
	gameState := NewGameState()

	gameState.Board = Board{
		{1, 0, 0, 0, 0, 8},
		{0, 1, 0, 1, 0, 0},
		{0, 0, 1, 0, 0, 0},
		{0, 0, 8, 0, 0, 0},
		{0, 0, 0, 0, 0, 8},
		{0, 8, 8, 0, 9, 1},
	}

	gameState.P1.Cats = 8
	gameState.P1.Kittens = 8
	gameState.P1.Placed = 0
	gameState.P2.Placed = 0

	return gameState
}

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
	State     string
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
		GameState: LoadTestGameState(),
		Players:   make(map[string]*websocket.Conn),
		send:      make(chan Message),
		// State:     "WAITING",
	}
}

func (server *Server) createGame(conn *websocket.Conn) *Game {
	server.serverMutex.Lock()
	defer server.serverMutex.Unlock()

	game := NewGame()
	game.Players["player1"] = conn
	server.games[game.ID] = game
	server.waitingGames[game.ID] = game
	log.Println(server.waitingGames)
	return game
}

func (server *Server) joinGame(conn *websocket.Conn, requestedGameID string) *Game {
	server.serverMutex.Lock()
	defer server.serverMutex.Unlock()

	// If specific game requested
	if requestedGameID != "" {
		if game, exists := server.waitingGames[requestedGameID]; exists {
			if len(game.Players) < 2 {
				playerID := fmt.Sprintf("player%d", len(game.Players)+1)
				game.Players[playerID] = conn
				delete(server.waitingGames, game.ID)
				return game
			}
			return nil // Game full
		}
		return nil // Game not found
	}
	return nil // No game ID provided
}

// Any WriteJSON call will only be called from here
func (game *Game) writePump(conn *websocket.Conn, playerID string, s *Server, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case msg, ok := <-game.send:
			// if channel is closed, return
			if !ok {
				return
			}
			if msg.Type == "error" {
				game.mutex.Lock()
				if err := conn.WriteJSON(msg); err != nil {
					log.Printf("Failed to write error message to %s: %v", playerID, err)
					s.handlePlayerDisconnect(game.ID, playerID)
					return
				}
				game.mutex.Unlock()
			}
			if msg.Type == "gameState" {
				game.mutex.Lock()

				for playerID, conn := range game.Players {
					msg.PlayerID = playerID
					if err := conn.WriteJSON(msg); err != nil {
						log.Printf("Failed to broadcast to %s: %v", playerID, err)
						s.handlePlayerDisconnect(game.ID, playerID)
						return
					}
				}

				game.mutex.Unlock()
			}
		case <-ticker.C:
			// Send ping to client every 30 seconds
			// conn.SetWriteDeadline(time.Now().Add(writeWait))
			log.Printf("Sending ping to %s", playerID)
			if err := conn.WriteJSON(Message{Type: "ping"}); err != nil {
				log.Printf("Failed to write ping to %s: %v", playerID, err)
				s.handlePlayerDisconnect(game.ID, playerID)
				return
			}
		}
	}
}

func (game *Game) readMove(conn *websocket.Conn, playerID string, s *Server) (error, Message, *NewMove) {
	var newMove NewMove
	errMsg := Message{
		Type:    "error",
		GameID:  game.ID,
		Payload: "Failed to read move",
		State:   game.GameState.State,
	}

	if err := conn.ReadJSON(&newMove); err != nil {
		// If the client closes without sending a close message
		if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
			log.Printf("Client closed without sending a close message")
			s.handlePlayerDisconnect(game.ID, playerID)
			errMsg.Payload = "Disconnected"
			return err, errMsg, nil
		}

		// Handles all other unexpected close errors
		if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
			log.Printf("unexpected websocket close, error: %v", err)
			s.handlePlayerDisconnect(game.ID, playerID)
			errMsg.Payload = "Disconnected"
			return err, errMsg, nil
		}

		s.handlePlayerDisconnect(game.ID, playerID)
		errMsg.Payload = "Disconnected"
		game.send <- errMsg
		log.Printf("Read error from %s: %v", playerID, err)
		return err, errMsg, nil
	}

	//If the received piece is 99, that is a client sending a pong to the server, reset the read deadline
	decodedPiece, _ := newMove.Piece.Int64()
	if decodedPiece == 99 {
		log.Printf("Pong received from %s", playerID)
		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Printf("Client failed to send pong in time, closing connection")
			s.handlePlayerDisconnect(game.ID, playerID)
			game.send <- errMsg
			errMsg.Payload = "Disconnected"
			return fmt.Errorf("Client failed to send pong in time"), errMsg, nil
		}
		return nil, Message{Type: "Pong", Payload: "Pong received"}, &newMove
	}

	if !game.isValidTurn(playerID) {
		errMsg.Payload = "Not your turn"
		game.send <- errMsg
		log.Printf("Not your turn %s", playerID)
		return fmt.Errorf("Not your turn"), errMsg, &newMove
	}

	return nil, Message{}, &newMove
}

const (
	// Time allowed to write a message to the peer
	pingPeriod = 30 * time.Second

	// Time allowd to read the next pong message from the peer
	pongWait = 60 * time.Second

	// If the server can't write to the peer within this time, the connection is closed
	writeWait = 10 * time.Second
)

// Any ReadJSON call will only be called from here
func (game *Game) readPump(conn *websocket.Conn, playerID string, s *Server, wg *sync.WaitGroup) {
	defer wg.Done()

	//The readpump will need to receive a ping from the client every 30 seconds to keep the connection alive
	conn.SetReadDeadline(time.Now().Add(pongWait))
	//When the pong is received, the read deadline is reset
	// conn.SetPongHandler(func(string) error {
	// 	conn.SetReadDeadline(time.Now().Add(pongWait))
	// 	return nil
	// })

	log.Printf("ReadPump: Player %s connected to game %s", playerID, game.ID)
	for {
		switch game.GameState.State {
		case "WAITING":
			// log.Println("ReadPump: WAITING")
			err, errMsg, newMove := game.readMove(conn, playerID, s)
			if err != nil || errMsg.Type == "Pong" {
				if errMsg.Payload == "Disconnected" {
					return
				}
				continue
			}
			if err := game.processTurn(newMove); err != nil {
				errMsg.Payload = err.Error()
				errMsg.State = game.GameState.State
				game.send <- errMsg
				continue
			}

		case "MULTIPLE_WAITING":
			// log.Println("ReadPump: MULTIPLE_WAITING")
			err, errMsg, newMove := game.readMove(conn, playerID, s)
			if err != nil || errMsg.Type == "Pong" {
				if errMsg.Payload == "Disconnected" {
					return
				}
				continue
			}
			if err := game.handleMultipleGraduations(newMove); err != nil {
				errMsg.Payload = err.Error()
				errMsg.State = game.GameState.State
				game.send <- errMsg
				continue
			}

		case "MAX_WAITING":
			// log.Println("ReadPump: MAX_WAITING")
			err, errMsg, newMove := game.readMove(conn, playerID, s)
			if err != nil || errMsg.Type == "Pong" {
				if errMsg.Payload == "Disconnected" {
					return
				}
				continue
			}
			if err := game.handleMaxedOutGraduation(newMove); err != nil {
				errMsg.Payload = err.Error()
				errMsg.State = game.GameState.State
				game.send <- errMsg
				continue
			}

		}
		if game.GameState.State == "WAITING" {
			game.GameState.TurnNumber++
		}
		// Broadcast updated state to all players
		game.broadcastGameState()
	}
}

// I want the game to handle the game loop and not the server
func (s *Server) handleGameLoop(conn *websocket.Conn, game *Game, playerID string) {
	defer func() {
		log.Printf("Defer Func: closing connection %s", game.ID)
		s.handlePlayerDisconnect(game.ID, playerID)
		conn.Close()
		return
	}()

	var wg sync.WaitGroup

	wg.Add(2)

	go game.readPump(conn, playerID, s, &wg)
	go game.writePump(conn, playerID, s, &wg)

	wg.Wait()

	log.Printf("End of handle Game loop function %s", game.ID)
}

func (game *Game) isValidTurn(playerID string) bool {
	return (game.GameState.isPlayer1() && playerID == "player1") ||
		(!game.GameState.isPlayer1() && playerID == "player2")
}

func (game *Game) processTurn(newMove *NewMove) error {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	game.GameState.PreviousBoard = game.GameState.Board

	//Not sure why I can't use Piece directly, probably due to the json.Number type
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
	// game.broadcastGameState(game, true)

	// Handle graduation logic
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

func getPlayerIDFromMap(m map[string]*websocket.Conn, conn *websocket.Conn) string {
	for playerID, c := range m {
		if c == conn {
			return playerID
		}
	}
	return ""
}

func (game *Game) handleMultipleGraduations(selection *NewMove) error {

	log.Printf("handleMultipleGrad: Called, selection: %+v", selection)
	log.Println(game.GameState.ThreeChoices)
	if !slices.Contains(game.GameState.ThreeChoices, selection.Position) {
		return fmt.Errorf("Error: HandleMultipleGraduations: The position selected is not a valid graduation piece")
	}

	line := game.GameState.getLineContainingPosition(selection.Position)
	if line == nil {
		return fmt.Errorf("Error: HandleMultipleGraduations: No complete line found on position. The position selected is not a valid graduation piece")
	}

	game.GameState.Board.graduatePieces(line, game.GameState)
	log.Println("handleMultipleGrad is changing the State to WAITING")
	game.GameState.State = "WAITING"
	return nil
}

func (game *Game) handleMaxedOutGraduation(selection *NewMove) error {
	log.Println("handleMaxedGrad: Called")
	playerPieces := game.GameState.Board.getPlayerPiecePositions(game.GameState)
	if !slices.Contains(playerPieces, selection.Position) {
		log.Println("handleMaxedGrad: Position not found in player pieces, returning error")
		return fmt.Errorf("Error: HandleMaxedOutGraduation: The position selected is not a valid graduation piece")
	}

	game.GameState.Board.graduatePiece(selection.Position, game.GameState)
	log.Println("handleMaxedGrad is changing the State to WAITING")
	game.GameState.State = "WAITING"

	return nil
}

func (game *Game) broadcastGameState() {
	log.Printf("Broadcasting game state, game state is: %+v", game.GameState.State)
	stateMsg := Message{
		Type:    "gameState",
		GameID:  game.ID,
		Payload: game.GameState,
		State:   game.GameState.State,
	}

	game.send <- stateMsg
}

// When a player disconnects, the game should be deleted from the server
func (s *Server) handlePlayerDisconnect(gameID string, playerID string) {
	s.serverMutex.Lock()
	defer s.serverMutex.Unlock()

	game, exists := s.games[gameID]
	if !exists {
		return
	}

	game.mutex.Lock()
	defer game.mutex.Unlock()

	game.broadcastGameState()

	if waitingGame, exists := s.waitingGames[gameID]; exists {
		delete(waitingGame.Players, playerID)
		delete(s.waitingGames, gameID)
	}

	delete(game.Players, playerID)

	// if len(game.Players) == 0 {
	delete(s.games, gameID)
	return
	// }

	// game.GameState.Winner = 0
	// Won't be able to broadcast if j
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
	(*w).Header().Set("Access-Control-Allow-Origin", os.Getenv("ORIGIN_URL"))
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	defer log.Println("Server shutting down")
	// addr := flag.String("addr", ":8080", "http service address")
	// flag.Parse()
	//
	if os.Getenv("ENV") == "PROD" {
		os.Setenv("ORIGIN_URL", "https://boop.oatmocha.com")
	} else {
		os.Setenv("ORIGIN_URL", "http://localhost:5173")
	}

	server := NewServer()
	// Using local mux instead of default as defaultservemux is a global var which can be accessed by any 3rd party package and
	// have routes registered to it
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.handleConnection)
	mux.HandleFunc("/getWaitingGame", server.handleGetWaitingGameID)

	log.Println("Server starting on, ", ":8080")
	// log.Println("Origin URL: ", os.Getenv("ORIGIN_URL"))
	// log.Println("ENV: ", os.Getenv("ENV"))
	// if os.Getenv("ENV") == "PROD" {
	// 	log.Println("Running in production mode")
	// }
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
