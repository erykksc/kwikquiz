package main

import (
	"fmt"
	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/config"
	"github.com/erykksc/kwikquiz/internal/database"
	"github.com/erykksc/kwikquiz/internal/lobbies"
	"github.com/erykksc/kwikquiz/internal/models"
	"github.com/erykksc/kwikquiz/internal/quiz"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var DEBUG = common.DebugOn()

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

func setUpDataBase() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load config
	cfg := config.LoadConfig()

	// Connect to database
	database.Connect(cfg)

	// Migrate the schema
	if err := database.DB.AutoMigrate(&models.Quiz{}, &models.Question{}, &models.Answer{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}

func testAddToDB() {
	// Save the example quiz to the database
	err := database.DB.Create(&quiz.ExampleQuizGeography).Error
	if err != nil {
		log.Fatalf("Failed to create example quiz: %v", err)
	}
}

func main() {
	var logLevel slog.Leveler = slog.LevelInfo
	if DEBUG {
		slog.Info("Debug mode enabled")
		logLevel = slog.LevelDebug
	}

	handler := getLoggingHandler(logLevel)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	setUpDataBase()

	testAddToDB()

	fs := http.FileServer(http.Dir("static"))

	router := http.NewServeMux()

	router.Handle("/quizzes/", quiz.NewQuizzesRouter())
	router.Handle("/lobbies/", lobbies.NewLobbiesRouter())
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
