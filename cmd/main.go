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

func setUpDatabase(conf config.Config) error {
	// Connect to the database
	database.Connect(conf)
	slog.Info("Database connected")

	// Drop existing tables (Should only be used development!)
	//database.DB.Migrator().DropTable(&models.Answer{}, &models.Question{}, &models.Quiz{}, &models.PastGame{},
	//	&models.PlayerScore{})

	// Migrate the schema
	migrateDatabase()

	quiz.InitRepo()
	lobbies.InitRepo()

	return nil
}

func migrateDatabase() {
	modelsToMigrate := []interface{}{
		&quiz.Quiz{},
		&quiz.Question{},
		&quiz.Answer{},
		// &pastgames.PastGame{},
		// &pastgames.PlayerScore{},
	}

	for _, model := range modelsToMigrate {
		if err := database.DB.AutoMigrate(model); err != nil {
			log.Fatalf("failed to migrate model %T: %v", model, err)
		}
	}
	slog.Info("Database migrated")
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

	// Set up database
	if err := setUpDatabase(conf); err != nil {
		log.Fatalf("failed to set up database: %v", err)
	}

	db, err := sqlx.Open("sqlite3", "kwikquiz.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	pastGamesRepo := pastgames.NewRepositorySQLite(db)
	err = pastGamesRepo.Initialize()
	if err != nil {
		slog.Error("failed to set up pastgames repo", "err", err)
		panic(err)
	}

	pastgames.Init(pastGamesRepo)

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

	// Add example data types
	if conf.InDevMode {
		// Pastgames
		for _, example := range pastgames.GetExamples() {
			pastgames.Repo.Upsert(&example)
		}
		// Quizzes
		for _, example := range quiz.GetExamples() {
			quiz.QuizzesRepo.AddQuiz(example)
		}
		// Lobbies
		for _, example := range lobbies.GetExamples() {
			lobbies.LobbiesRepo.AddLobby(example)
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
