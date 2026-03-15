//go:build db

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	_ "modernc.org/sqlite"
)

func initDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS games (
			id TEXT PRIMARY KEY,
			state TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return db, err
}

func (s *Server) saveGame(game *Game) {
	data, err := json.Marshal(game.GameState)
	if err != nil {
		log.Printf("saveGame: failed to marshal game state: %v", err)
		return
	}
	_, err = s.db.Exec(`
		INSERT INTO games (id, state, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET state = excluded.state, updated_at = excluded.updated_at
	`, game.ID, string(data))
	if err != nil {
		log.Printf("saveGame: failed to save game %s: %v", game.ID, err)
	}
}

func (s *Server) loadGame(gameID string) (*GameState, error) {
	var stateJSON string
	err := s.db.QueryRow(`SELECT state FROM games WHERE id = ?`, gameID).Scan(&stateJSON)
	if err != nil {
		return nil, err
	}
	var gameState GameState
	if err := json.Unmarshal([]byte(stateJSON), &gameState); err != nil {
		return nil, err
	}
	return &gameState, nil
}

func (s *Server) deleteGame(gameID string) {
	_, err := s.db.Exec(`DELETE FROM games WHERE id = ?`, gameID)
	if err != nil {
		log.Printf("deleteGame: failed to delete game %s: %v", gameID, err)
	}
}
