package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func (s *Server) getScoresForGame(gameID int) ([]Score, error) {
	scoreRows, err := s.db.Query(`
        SELECT
            p.name, s.legacy_score, s.base_cards, s.extra_vp, s.basic_events,
            s.special_events, s.prosperity_cards, s.visitors, s.journey, s.garland_award
        FROM scores s
        JOIN players p ON s.player_id = p.id
        WHERE s.game_id = ?`, gameID)
	if err != nil {
		return nil, err
	}
	defer scoreRows.Close()

	var scores []Score
	for scoreRows.Next() {
		var s Score
		if err := scoreRows.Scan(
			&s.PlayerName, &s.LegacyScore, &s.BaseCards, &s.ExtraVP, &s.BasicEvents,
			&s.SpecialEvents, &s.ProsperityCards, &s.Visitors, &s.Journey, &s.GarlandAward,
		); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	return scores, nil
}

func (s *Server) getGamesHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("received request to get all games", "method", r.Method, "path", r.URL.Path)
	rows, err := s.db.Query("SELECT id, game_date FROM games ORDER BY game_date DESC")
	if err != nil {
		s.logger.Error("failed to query games", "error", err)
		http.Error(w, "Failed to query games", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var game Game
		if err := rows.Scan(&game.ID, &game.GameDate); err != nil {
			s.logger.Error("failed to scan game row", "error", err, "game_id", game.ID)
			http.Error(w, "Failed to scan game row", http.StatusInternalServerError)
			return
		}
		scores, err := s.getScoresForGame(game.ID)
		if err != nil {
			s.logger.Error("failed to get scores for game", "error", err, "game_id", game.ID)
			http.Error(w, "Failed to get scores for game", http.StatusInternalServerError)
			return
		}
		game.Scores = scores
		games = append(games, game)
	}
	s.logger.Info("successfully retrieved all games", "count", len(games))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

func (s* Server) getGameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}
	handlerLogger := s.logger.With("game_id", id)
	handlerLogger.Info("received request to get game", "method", r.Method, "path", r.URL.Path)

	var game Game
	err = s.db.QueryRow("SELECT id, game_date FROM games WHERE id = ?", id).Scan(&game.ID, &game.GameDate)
	if err == sql.ErrNoRows {
		handlerLogger.Warn("game not found")
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}
	if err != nil {
		handlerLogger.Error("failed to query game", "error", err)
		http.Error(w, "Failed to query game", http.StatusInternalServerError)
		return
	}

	scores, err := s.getScoresForGame(id)
	if err != nil {
		handlerLogger.Error("failed to retrieve scores for game", "error", err)
		http.Error(w, "Failed to retrieve scores", http.StatusInternalServerError)
		return
	}
	game.Scores = scores
	handlerLogger.Info("successfully retrieved game")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func (s *Server) createGameHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("received request to create game", "method", r.Method, "path", r.URL.Path)
	var game Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		s.logger.Error("failed to decode request payload", "error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if game.GameDate.IsZero() {
		game.GameDate = time.Now()
	}

	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error("failed to start database transaction", "error", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec("INSERT INTO games (game_date) VALUES (?)", game.GameDate)
	if err != nil {
		s.logger.Error("failed to insert game record", "error", err)
		http.Error(w, "Failed to insert game", http.StatusInternalServerError)
		return
	}
	gameID, err := result.LastInsertId()
	if err != nil {
		s.logger.Error("failed to get last insert ID for game", "error", err)
		http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
		return
	}
	gameLogger := s.logger.With("game_id", gameID)

	for _, score := range game.Scores {
		playerLogger := gameLogger.With("player_name", score.PlayerName)
		var playerID int64
		err := tx.QueryRow("SELECT id FROM players WHERE name = ?", score.PlayerName).Scan(&playerID)
		if err == sql.ErrNoRows {
			playerLogger.Info("player not found, creating new player")
			res, err := tx.Exec("INSERT INTO players (name) VALUES (?)", score.PlayerName)
			if err != nil {
				playerLogger.Error("failed to insert new player", "error", err)
				http.Error(w, "Failed to insert new player", http.StatusInternalServerError)
				return
			}
			playerID, err = res.LastInsertId()
			if err != nil {
				playerLogger.Error("failed to get new player ID", "error", err)
				http.Error(w, "Failed to get new player ID", http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			playerLogger.Error("failed to query for player", "error", err)
			http.Error(w, "Failed to query for player", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
            INSERT INTO scores (game_id, player_id, base_cards, extra_vp, basic_events, special_events, prosperity_cards, visitors, journey, garland_award)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, gameID, playerID, score.BaseCards, score.ExtraVP, score.BasicEvents, score.SpecialEvents, score.ProsperityCards, score.Visitors, score.Journey, score.GarlandAward)
		if err != nil {
			playerLogger.Error("failed to insert score", "error", err)
			http.Error(w, "Failed to insert score", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		gameLogger.Error("failed to commit transaction", "error", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}
	gameLogger.Info("successfully created game")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"gameId": gameID})
}