package main

import (
	"fmt"
	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/config"
	"github.com/erykksc/kwikquiz/internal/database"
	"github.com/erykksc/kwikquiz/internal/lobbies"
	"github.com/erykksc/kwikquiz/internal/models"
	"github.com/erykksc/kwikquiz/internal/pastgames"
	"github.com/erykksc/kwikquiz/internal/quiz"
	"log"
	"log/slog"
	"net/http"
	"os"
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

func setUpDatabase() {
	// Load config
	cfg, _ := config.LoadConfig()

	// Connect to the database
	database.Connect(cfg)

	// Migrate the schema
	migrateDatabase()
}

func migrateDatabase() {
	modelsToMigrate := []interface{}{
		&models.Quiz{},
		&models.Question{},
		&models.Answer{},
		&models.PastGame{},
		&models.PlayerScore{},
	}

	for _, model := range modelsToMigrate {
		if err := database.DB.AutoMigrate(model); err != nil {
			log.Fatalf("failed to migrate model %T: %v", model, err)
		}
	}
}

func main() {
	var logLevel slog.Leveler = slog.LevelInfo
	setUpDatabase()

	if common.DebugOn() {
		slog.Info("Debug mode enabled")
		logLevel = slog.LevelDebug
	}

	handler := getLoggingHandler(logLevel)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	fs := http.FileServer(http.Dir("static"))

	router := http.NewServeMux()

	router.Handle("/quizzes/", quiz.NewQuizzesRouter())
	router.Handle("/lobbies/", lobbies.NewLobbiesRouter())
	router.Handle("/past-games/", pastgames.NewPastGamesRouter())
	router.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		if err := common.IndexTmpl.Execute(w, nil); err != nil {
			slog.Error("Error rendering template", "error", err)
		}
	})
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	port := 3000
	addr := fmt.Sprintf(":%d", port)
	slog.Info("Server listening", "addr", addr)

	err := http.ListenAndServe(addr, loggingMiddleware(router))
	if err != nil {
		slog.Error("Server shutting down", "err", err.Error())
	}
}
