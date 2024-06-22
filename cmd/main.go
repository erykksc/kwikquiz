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
	err := database.DB.AutoMigrate(&models.QuizModel{}, &models.QuestionModel{}, &models.AnswerModel{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}

func testInsertDataToDatabase() {
	var ExampleQuizGeography = models.QuizModel{
		ID:          1,
		Title:       "Geography",
		Description: "This is a quiz about capitals around the world",
		Questions: []models.QuestionModel{
			{
				Text: "What is the capital of France?",
				Answers: []models.AnswerModel{
					{Text: "Paris", IsCorrect: true},
					{Text: "Berlin", IsCorrect: false},
					{Text: "Warsaw", IsCorrect: false},
					{Text: "Barcelona", IsCorrect: false},
				},
			},
			{
				Text: "On which continent is Russia?",
				Answers: []models.AnswerModel{
					{Text: "Europe", IsCorrect: true},
					{Text: "Asia", IsCorrect: true},
					{Text: "North America", IsCorrect: false},
					{Text: "South America", IsCorrect: false},
				},
			},
		},
	}
	result := database.DB.Create(&ExampleQuizGeography)
	fmt.Println(result.RowsAffected)
	fmt.Println(result.Error)

}

func main() {
	var logLevel slog.Leveler = slog.LevelInfo
	if common.DebugOn() {
		slog.Info("Debug mode enabled")
		logLevel = slog.LevelDebug
	}

	handler := getLoggingHandler(logLevel)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	setUpDatabase()
	testInsertDataToDatabase()

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
