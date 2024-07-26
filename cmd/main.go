package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/config"
	"github.com/erykksc/kwikquiz/internal/lobbies"
	"github.com/erykksc/kwikquiz/internal/pastgames"
	"github.com/erykksc/kwikquiz/internal/quiz"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug(fmt.Sprintf("%s %s", r.Method, r.URL.Path))

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func getLoggingHandler(level slog.Leveler) slog.Handler {
	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}
	handler := slog.NewJSONHandler(os.Stderr, opts)

	return handler
}

func main() {
	// Load config from environmental variables
	if err := config.LoadEnv(".env"); err != nil {
		slog.Warn("Couldn't load .env file", "error", err)
	}
	conf, err := config.LoadConfigFromEnv()
	if err != nil {
		panic(err)
	}

	// Set up logging
	var logLevel slog.Leveler = slog.LevelInfo
	if conf.InDevMode {
		slog.Info("Development mode enabled, setting LogLevel to Debug")
		logLevel = slog.LevelDebug
	}
	handler := getLoggingHandler(logLevel)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Open sqlite database connection
	db, err := sqlx.Open("sqlite3", "kwikquiz.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Enforce CASCADE in sqlite, this needs to run before any other query
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatal(err)
	}

	// Setup pastgames Service
	pastGamesRepo := pastgames.NewRepositorySQLite(db)
	err = pastGamesRepo.Initialize()
	if err != nil {
		slog.Error("failed to set up pastgames repo", "err", err)
		panic(err)
	}
	pastGamesService := pastgames.NewService(pastGamesRepo)

	// Setup Quiz Service
	quizRepo := quiz.NewRepositorySQLite(db)
	err = quizRepo.Initialize()
	if err != nil {
		slog.Error("failed to set up quiz repo", "err", err)
		panic(err)
	}
	quizService := quiz.NewService(quizRepo)

	// Setup lobbies Service
	lobbiesRepo := lobbies.NewRepositoryInMemory()
	lobbiesService := lobbies.NewService(lobbiesRepo, pastGamesRepo, quizRepo)

	// Set up routes
	router := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	router.Handle("/quizzes/", quizService.NewQuizzesRouter())
	router.Handle("/lobbies/", lobbiesService.NewLobbiesRouter())
	router.Handle("/past-games/", pastGamesService.NewPastGamesRouter())
	router.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		if err := common.IndexTmpl.Execute(w, nil); err != nil {
			slog.Error("Error rendering template", "error", err)
		}
	})

	// Add example data types
	if conf.InDevMode {
		// Pastgames
		for _, example := range pastgames.GetExamples() {
			pastGamesRepo.Upsert(&example)
		}
		// Quizzes
		for _, example := range quiz.GetExamples() {
			quizRepo.Insert(&example)
		}
		// Lobbies
		for _, example := range lobbies.GetExamples() {
			lobbiesRepo.AddLobby(example)
		}
	}

	// Start server
	port := 3000
	addr := fmt.Sprintf(":%d", port)
	slog.Info("Server listening", "addr", addr)

	err = http.ListenAndServe(addr, loggingMiddleware(router))
	if err != nil {
		slog.Error("Server shutting down", "err", err.Error())
	}
}
