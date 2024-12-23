package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"slices"
	"sync"
)

//	func LoadTestGameState() *GameState {
//		gameState := NewGameState()
//
//		gameState.Board = Board{
//			{1, 0, 0, 0, 0, 8},
//			{0, 1, 0, 1, 0, 0},
//			{0, 0, 1, 0, 0, 0},
//			{0, 0, 8, 0, 0, 0},
//			{0, 0, 0, 0, 0, 8},
//			{0, 0, 0, 0, 0, 0},
//		}
//
//		gameState.P1.Cats = 8
//		gameState.P1.Kittens = 8
//		gameState.P1.Placed = 0
//		gameState.P2.Placed = 0
//
//		return gameState
//	}
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
	GameID   string      `json:"gameId"`
	PlayerID string      `json:"playerId"`
	State    string      `json:"state"`
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
		games:        make(map[string]*Game),
		waitingGames: make(map[string]*Game),
	}
}

func NewGame() *Game {
	return &Game{
		ID:        generateGameID(),
		GameState: NewGameState(),
		Players:   make(map[string]*websocket.Conn),
		send:      make(chan Message),
		State:     "WAITING",
	}
}

func (server *Server) createGame(conn *websocket.Conn) *Game {
	server.serverMutex.Lock()
	defer server.serverMutex.Unlock()

	game := NewGame()
	game.Players["player1"] = conn
	server.games[game.ID] = game
	server.waitingGames[game.ID] = game
	fmt.Println(server.waitingGames)
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
					}
				}

				game.mutex.Unlock()
			}
		}
	}
}

func readMove(conn *websocket.Conn, playerID string, game *Game, s *Server) (error, Message, *NewMove) {
	var newMove NewMove
	errMsg := Message{
		Type:    "error",
		GameID:  game.ID,
		Payload: "Failed to read move",
		State:   game.State,
	}

	if err := conn.ReadJSON(&newMove); err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("unexpected websocket close, error: %v", err)
			s.handlePlayerDisconnect(game.ID, playerID)
		}

		game.send <- errMsg
		log.Printf("Read error from %s: %v", playerID, err)
		return err, errMsg, nil
	}

	if !s.isValidTurn(game, playerID) {
		errMsg.Payload = "Not your turn"
		game.send <- errMsg
		log.Printf("Not your turn %s", playerID)
		return fmt.Errorf("Not your turn"), errMsg, &newMove
	}

	return nil, Message{}, &newMove
}

// Any ReadJSON call will only be called from here
func (game *Game) readPump(conn *websocket.Conn, playerID string, s *Server, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("ReadPump: Player %s connected to game %s", playerID, game.ID)
	// errMsg := Message{
	// 	Type:    "error",
	// 	GameID:  game.ID,
	// 	Payload: "",
	// 	State:   "",
	// }
	for {

		// if err := conn.ReadJSON(&newMove); err != nil {
		// 	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		// 		log.Printf("unexpected websocket close, error: %v", err)
		// 		s.handlePlayerDisconnect(game.ID, playerID)
		// 	}
		// 	errMsg.Payload = "Failed to read move"
		// 	errMsg.State = game.State
		// 	game.send <- errMsg
		// 	log.Printf("Read error from %s: %v", playerID, err)
		// 	return
		// }

		// if !s.isValidTurn(game, playerID) {
		// 	errMsg.Payload = "Not your turn"
		// 	errMsg.State = game.State
		// 	game.send <- errMsg
		// 	continue // Continue waiting for valid turn instead of disconnecting
		// }

		switch game.State {
		case "WAITING":
			err, errMsg, newMove := readMove(conn, playerID, game, s)
			if err != nil {
				continue
			}
			if err := s.processTurn(conn, game, newMove); err != nil {
				errMsg.Payload = err.Error()
				errMsg.State = game.State
				game.send <- errMsg
				continue
			}

		case "MULTIPLE_WAITING":
			log.Println("ReadPump: MULTIPLE_WAITING")
			err, errMsg, newMove := readMove(conn, playerID, game, s)
			if err != nil {
				continue
			}
			if err := s.handleMultipleGraduations(conn, game, newMove); err != nil {
				errMsg.Payload = err.Error()
				errMsg.State = game.State
				game.send <- errMsg
				continue
			}

		case "MAX_WAITING":
			log.Println("ReadPump: MAX_WAITING")
			err, errMsg, newMove := readMove(conn, playerID, game, s)
			if err != nil {
				continue
			}
			if err := s.handleMaxedOutGraduation(conn, game, newMove); err != nil {
				errMsg.Payload = err.Error()
				errMsg.State = game.State
				game.send <- errMsg
				continue
			}

		}
		if game.State == "WAITING" {
			game.GameState.TurnNumber++
		}
		// Broadcast updated state to all players
		s.broadcastGameState(game, false)
	}
}
func (s *Server) handleGameLoop(conn *websocket.Conn, game *Game, playerID string) {
	defer func() {
		log.Printf("Defer Func: closing connection %s", game.ID)
		s.handlePlayerDisconnect(game.ID, playerID)
		conn.Close()
	}()

	var wg sync.WaitGroup

	wg.Add(2)

	go game.readPump(conn, playerID, s, &wg)
	go game.writePump(conn, playerID, s, &wg)

	wg.Wait()

	log.Printf("End of handle Game loop function %s", game.ID)
}

func (s *Server) isValidTurn(game *Game, playerID string) bool {
	return (game.GameState.isPlayer1() && playerID == "player1") ||
		(!game.GameState.isPlayer1() && playerID == "player2")
}

func (s *Server) processTurn(conn *websocket.Conn, game *Game, newMove *NewMove) error {
	game.mutex.Lock()
	defer game.mutex.Unlock()

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
	s.broadcastGameState(game, true)

	// Handle graduation logic
	if len(game.GameState.Lines) > 1 {
		// game.GameState.Waiting = true
		game.State = "MULTIPLE_WAITING"
		// if err := s.handleMultipleGraduations(conn, game); err != nil {
		//	return err
		// }
	} else if len(game.GameState.Lines) == 1 {
		game.GameState.Board.graduatePieces(game.GameState.Lines[0], game.GameState)
		// game.GameState.TurnNumber++
	} else {
		if s.shouldCheckMaxedOut(game) {
			if game.GameState.Board.winCheckMaxCats(game.GameState) {
				return nil
			}
			game.State = "MAX_WAITING"
			// if err := s.handleMaxedOutGraduation(conn, game); err != nil {
			// 	return err
			// }
		}
	}

	return nil
}

func (s *Server) shouldCheckMaxedOut(game *Game) bool {
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

func (s *Server) handleMultipleGraduations(conn *websocket.Conn, game *Game, selection *NewMove) error {

	if !slices.Contains(game.GameState.ThreeChoices, selection.Position) {
		return fmt.Errorf("Error: HandleMultipleGraduations: The position selected is not a valid graduation piece")
	}

	line := game.GameState.getLineContainingPosition(selection.Position)
	if line == nil {
		return fmt.Errorf("Error: HandleMultipleGraduations: No complete line found on position. The position selected is not a valid graduation piece")
	}

	game.GameState.Board.graduatePieces(line, game.GameState)
	game.State = "WAITING"
	return nil
}

func (s *Server) handleMaxedOutGraduation(conn *websocket.Conn, game *Game, selection *NewMove) error {
	log.Println("handleMaxedGrad: Called")
	playerPieces := game.GameState.Board.getPlayerPiecePositions(game.GameState)
	if !slices.Contains(playerPieces, selection.Position) {
		return fmt.Errorf("Error: HandleMaxedOutGraduation: The position selected is not a valid graduation piece")
	}

	game.GameState.Board.graduatePiece(selection.Position, game.GameState)
	log.Println("handleMaxedGrad is changing the State to WAITING")
	game.State = "WAITING"

	return nil
}

func (s *Server) broadcastGameState(game *Game, alreadyLocked bool) {
	stateMsg := Message{
		Type:    "gameState",
		GameID:  game.ID,
		Payload: game.GameState,
		State:   game.State,
	}

	game.send <- stateMsg
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
		return
	}

	game.GameState.Winner = 0
	// Won't be able to broadcast if j
	// s.broadcastGameState(game, true)
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
	addr := flag.String("addr", ":8080", "http service address")
	flag.Parse()

	server := NewServer()
	// Using local mux instead of default as defaultservemux is a global var which can be accessed by any 3rd party package and
	// have routes registered to it
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.handleConnection)
	mux.HandleFunc("/getWaitingGame", server.handleGetWaitingGameID)

	log.Println("Server starting on, ", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
