package main

import (
	"database/sql"
	"os"
)

func (s *Server) initDB(filepath string) {
	var err error
	s.db, err = sql.Open("sqlite3", filepath)
	if err != nil {
		s.logger.Error("failed to open database connection", "error", err)
		os.Exit(1)
	}

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

	if _, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS games ( "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "game_date" DATETIME );`); err != nil {
		s.logger.Error("error creating games table", "error", err)
		os.Exit(1)
	}
	if _, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS players ( "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "name" TEXT UNIQUE );`); err != nil {
		s.logger.Error("error creating players table", "error", err)
		os.Exit(1)
	}
	if _, err = s.db.Exec(createScoresTableSQL); err != nil {
		s.logger.Error("error creating scores table", "error", err)
		os.Exit(1)
	}

	s.logger.Info("database initialized successfully")
}