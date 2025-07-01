package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB(filepath string) {
	var err error
	db, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	// This schema is correct: it has NO total_score column.
	createScoresTableSQL := `CREATE TABLE IF NOT EXISTS scores (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"game_id" INTEGER,
		"player_id" INTEGER,
		"legacy_score" INTEGER,
		"base_cards" INTEGER,
		"extra_vp" INTEGER,
		"basic_events" INTEGER,
		"special_events" INTEGER,
		"prosperity_cards" INTEGER,
		"visitors" INTEGER,
		"journey" INTEGER,
		"garland_award" INTEGER,
		FOREIGN KEY(game_id) REFERENCES games(id),
		FOREIGN KEY(player_id) REFERENCES players(id)
	);`

	// These CREATE statements remain the same and are correct.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS games ( "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "game_date" DATETIME );`)
	if err != nil { log.Fatalf("Error creating games table: %q", err) }
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS players ( "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "name" TEXT UNIQUE );`)
	if err != nil { log.Fatalf("Error creating players table: %q", err) }
	_, err = db.Exec(createScoresTableSQL)
	if err != nil { log.Fatalf("Error creating scores table: %q", err) }

	log.Println("Database initialized successfully")
}

// CORRECTED: The Score struct NO LONGER has a TotalScore field.
// It only describes the raw data that is stored and retrieved.
type Score struct {
	PlayerName      string `json:"player_name"`
	LegacyScore     *int   `json:"legacy_score,omitempty"`
	BaseCards       *int   `json:"base_cards,omitempty"`
	ExtraVP         *int   `json:"extra_vp,omitempty"`
	BasicEvents     *int   `json:"basic_events,omitempty"`
	SpecialEvents   *int   `json:"special_events,omitempty"`
	ProsperityCards *int   `json:"prosperity_cards,omitempty"`
	Visitors        *int   `json:"visitors,omitempty"`
	Journey         *int   `json:"journey,omitempty"`
	GarlandAward    *int   `json:"garland_award,omitempty"`
}

type Game struct {
	ID       int       `json:"id"`
	GameDate time.Time `json:"game_date"`
	Scores   []Score   `json:"scores"`
}

// This function is correct. It only inserts the component scores.
func createGameHandler(w http.ResponseWriter, r *http.Request) {
	var game Game
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&game); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if game.GameDate.IsZero() {
		game.GameDate = time.Now()
	}

	tx, err := db.Begin()
	if err != nil { http.Error(w, "Failed to start transaction", http.StatusInternalServerError); return }

	result, err := tx.Exec("INSERT INTO games (game_date) VALUES (?)", game.GameDate)
	if err != nil { tx.Rollback(); http.Error(w, "Failed to insert game", http.StatusInternalServerError); return }
	gameID, err := result.LastInsertId()
	if err != nil { tx.Rollback(); http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError); return }

	for _, score := range game.Scores {
		var playerID int64
		err := tx.QueryRow("SELECT id FROM players WHERE name = ?", score.PlayerName).Scan(&playerID)
		if err == sql.ErrNoRows {
			res, err := tx.Exec("INSERT INTO players (name) VALUES (?)", score.PlayerName)
			if err != nil { tx.Rollback(); http.Error(w, fmt.Sprintf("Failed to insert new player %s", score.PlayerName), http.StatusInternalServerError); return }
			playerID, err = res.LastInsertId()
			if err != nil { tx.Rollback(); http.Error(w, "Failed to get new player ID", http.StatusInternalServerError); return }
		} else if err != nil {
			tx.Rollback(); http.Error(w, "Failed to query for player", http.StatusInternalServerError); return
		}

		_, err = tx.Exec(`
			INSERT INTO scores (game_id, player_id, base_cards, extra_vp, basic_events, special_events, prosperity_cards, visitors, journey, garland_award)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, gameID, playerID, score.BaseCards, score.ExtraVP, score.BasicEvents, score.SpecialEvents, score.ProsperityCards, score.Visitors, score.Journey, score.GarlandAward)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to insert score", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil { http.Error(w, "Failed to commit transaction", http.StatusInternalServerError); return }
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"gameId": gameID})
}

// CORRECTED: This function is now much simpler.
// It just selects the raw data without any calculation.
func getScoresForGame(gameID int) ([]Score, error) {
	scoreRows, err := db.Query(`
		SELECT
			p.name, s.legacy_score, s.base_cards, s.extra_vp, s.basic_events,
			s.special_events, s.prosperity_cards, s.visitors, s.journey, s.garland_award
		FROM scores s
		JOIN players p ON s.player_id = p.id
		WHERE s.game_id = ?
	`, gameID)
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

// --- The rest of the file (getGamesHandler, getGameHandler, main) remains the same ---
func getGamesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, game_date FROM games ORDER BY game_date DESC")
    if err != nil { 
		log.Printf("Error querying games: %v", err)
		http.Error(w, "Failed to query games", http.StatusInternalServerError); 
		return 
	}
    
	defer rows.Close()
    var games []Game
    for rows.Next() {
        var game Game
        if err := rows.Scan(&game.ID, &game.GameDate); err != nil { http.Error(w, "Failed to scan game row", http.StatusInternalServerError); return }
        scores, err := getScoresForGame(game.ID)
        if err != nil { http.Error(w, "Failed to get scores for game", http.StatusInternalServerError); return }
        game.Scores = scores
        games = append(games, game)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(games)
}
func getGameHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil { http.Error(w, "Invalid game ID", http.StatusBadRequest); return }
    var game Game
    err = db.QueryRow("SELECT id, game_date FROM games WHERE id = ?", id).Scan(&game.ID, &game.GameDate)
    if err == sql.ErrNoRows { http.Error(w, "Game not found", http.StatusNotFound); return }
	if err != nil { http.Error(w, "Failed to query game", http.StatusInternalServerError); return }
    scores, err := getScoresForGame(id)
    if err != nil { http.Error(w, "Failed to retrieve scores", http.StatusInternalServerError); return }
    game.Scores = scores
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(game)
}
func main() {
	initDB("./everdell_games.db")
	defer db.Close()
	r := mux.NewRouter()
	r.HandleFunc("/games", createGameHandler).Methods("POST")
	r.HandleFunc("/games", getGamesHandler).Methods("GET")
	r.HandleFunc("/games/{id}", getGameHandler).Methods("GET")
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:5173"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})
	corsHandler := handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)
	port := "8181"
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil { log.Fatal(err) }
}