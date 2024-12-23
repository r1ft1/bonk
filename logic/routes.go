package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	var gameID = r.URL.Query().Get("gameID")
	var game *Game
	var playerID string

	if gameID == "" {
		game = s.createGame(conn)
		playerID = "player1"
	} else {
		game = s.joinGame(conn, gameID)
		playerID = "player2"
		if game == nil {
			conn.WriteJSON(Message{Type: "error", Payload: "Could not join game"})
			conn.Close()
			return
		}
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
	log.Print(s.games)

	// Start game loop
	s.handleGameLoop(conn, game, playerID)
}

func (s *Server) handleGetWaitingGameID(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /getWaitingGame")
	enableCors(&w)
	s.serverMutex.Lock()
	defer s.serverMutex.Unlock()

	//Send list of waiting game IDs to client
	type GameIDs struct {
		IDs []string `json:"ids"`
	}
	var IDs GameIDs
	if len(s.waitingGames) != 0 {
		for id := range s.waitingGames {
			IDs.IDs = append(IDs.IDs, id)
		}
	} else {
		IDs.IDs = append(IDs.IDs, "No games waiting")
	}
	jsonID, _ := json.Marshal(IDs)
	log.Printf(string(jsonID))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonID)
}
