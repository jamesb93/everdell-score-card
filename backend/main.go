package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)


type Server struct {
	db     *sql.DB
	logger *slog.Logger
}



func main() {
	var db *sql.DB
	var logger *slog.Logger

	server := &Server {
		db: db,
		logger: logger,
	}

	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("failed to open log file", "error", err)
		os.Exit(1)
	}
	server.logger = slog.New(slog.NewJSONHandler(logFile, nil))

	server.initDB("./everdell_games.db")
	defer server.db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/games", server.createGameHandler).Methods("POST")
	r.HandleFunc("/games", server.getGamesHandler).Methods("GET")
	r.HandleFunc("/games/{id}", server.getGameHandler).Methods("GET")

	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:5173", "https://everdell.fun"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})
	corsHandler := handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)

	port := "8181"
	server.logger.Info("server starting", "port", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		server.logger.Error("server startup failed", "error", err)
		os.Exit(1)
	}
}