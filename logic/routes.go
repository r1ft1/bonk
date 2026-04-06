package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(4096) // valid moves are tiny JSON; prevent memory exhaustion

	gameID := r.URL.Query().Get("gameID")
	var game *Game
	var playerID string

	if gameID == "" {
		game = s.createGame(conn)
		playerID = "player1"

		// First player starts the writePump for this game
		var wpWg sync.WaitGroup
		wpWg.Add(1)
		go game.writePump(s, &wpWg)
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
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := conn.WriteJSON(Message{
		Type:     "joined",
		GameID:   game.ID,
		PlayerID: playerID,
		Payload:  game.GameState,
	}); err != nil {
		log.Printf("Error sending initial game state: %v", err)
		s.handlePlayerDisconnect(game.ID, playerID)
		return
	}

	// Notify all players when a second player joins
	if playerID == "player2" {
		game.broadcastGameState()
	}

	log.Printf("Player %s joined game %s", playerID, game.ID)
	s.handleGameLoop(conn, game, playerID)
}

func (s *Server) handleGetWaitingGameID(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	s.serverMutex.Lock()
	defer s.serverMutex.Unlock()

	type GameIDs struct {
		IDs []string `json:"ids"`
	}
	var ids GameIDs
	if len(s.waitingGames) != 0 {
		for id := range s.waitingGames {
			ids.IDs = append(ids.IDs, id)
		}
	} else {
		ids.IDs = append(ids.IDs, "No games waiting")
	}
	jsonID, _ := json.Marshal(ids)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonID)
}
