package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/config"
	"github.com/erykksc/kwikquiz/internal/database"
	"github.com/erykksc/kwikquiz/internal/lobbies"
	"github.com/erykksc/kwikquiz/internal/models"
	"github.com/erykksc/kwikquiz/internal/pastgames"
	"github.com/erykksc/kwikquiz/internal/quiz"
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

func setUpDatabase() error {
	// Load environment variables
	if err := config.LoadEnv(".env"); err != nil {
		return err
	}
	slog.Info("Environment variables loaded")

	// Load config from environment variables
	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	slog.Info("Config for DB loaded")

	// Connect to the database
	database.Connect(cfg)
	slog.Info("Database connected")

	// Migrate the schema
	migrateDatabase()

	return nil
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
	slog.Info("Database migrated")
}

func main() {
	// Set up logging
	var logLevel slog.Leveler = slog.LevelInfo
	if common.DebugOn() {
		slog.Info("Debug mode enabled")
		logLevel = slog.LevelDebug
	}
	handler := getLoggingHandler(logLevel)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Set up database
	if err := setUpDatabase(); err != nil {
		log.Fatalf("failed to set up database: %v", err)
	}

	// Set up routes
	router := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	router.Handle("/quizzes/", quiz.NewQuizzesRouter())
	router.Handle("/lobbies/", lobbies.NewLobbiesRouter())
	router.Handle("/past-games/", pastgames.NewPastGamesRouter())
	router.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		if err := common.IndexTmpl.Execute(w, nil); err != nil {
			slog.Error("Error rendering template", "error", err)
		}
	})

	// Start server
	port := 3000
	addr := fmt.Sprintf(":%d", port)
	slog.Info("Server listening", "addr", addr)

	err := http.ListenAndServe(addr, loggingMiddleware(router))
	if err != nil {
		slog.Error("Server shutting down", "err", err.Error())
	}
}
