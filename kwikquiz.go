package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/lobbies"
	"github.com/erykksc/kwikquiz/internal/pastgames"
	"github.com/erykksc/kwikquiz/internal/quiz"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed static
var staticFS embed.FS

// Variables used for command line parameters
var (
	Port       uint
	InProdMode bool
	InDevMode  bool
)

func init() {
	flag.UintVar(&Port, "port", 3000, "Port to host the app")
	flag.BoolVar(&InProdMode, "prod", false, "Run the app in production mode")
	flag.Parse()

	InDevMode = !InProdMode
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("HTTP Call", "method", r.Method, "url_path", r.URL.Path)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Set up logging
	opts := slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	}
	if InDevMode {
		slog.Info("Development mode enabled, setting LogLevel to Debug and adding source")
		opts.Level = slog.LevelDebug
		opts.AddSource = true
	}
	handler := slog.NewJSONHandler(os.Stderr, &opts)
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
	pastGamesRepo, err := pastgames.NewRepositorySQLite(db)
	if err != nil {
		slog.Error("failed to set up pastgames repo", "err", err)
		panic(err)
	}
	pastGamesService := pastgames.NewService(pastGamesRepo)

	// Setup Quiz Service
	quizRepo, err := quiz.NewRepositorySQLite(db)
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

	fs := http.FileServer(http.FS(staticFS))
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
	if InDevMode {
		// Pastgames
		slog.Debug("Upserting examples pastgames")
		for _, example := range pastgames.GetExamples() {
			_, err := pastGamesRepo.Upsert(&example)
			if err != nil {
				slog.Error("Failed to upsert example pastgame", "err", err)
			}
		}
		// Quizzes
		slog.Debug("Upserting example quizzes")
		for _, example := range quiz.GetExamples() {
			_, err := quizRepo.Upsert(&example)
			if err != nil {
				slog.Error("Failed to upsert example quiz", "err", err)
			}
		}
		// Lobbies
		slog.Debug("Adding example lobbies")
		for _, example := range lobbies.GetExamples() {
			err := lobbiesRepo.AddLobby(example)
			if err != nil {
				slog.Error("Failed to add example lobbies", "err", err)
			}
		}
		slog.Debug("Finished upserting example data")
	}

	// Start server
	addr := fmt.Sprintf(":%d", Port)
	slog.Info("Server listening", "addr", addr)

	err = http.ListenAndServe(addr, loggingMiddleware(router))
	if err != nil {
		slog.Error("Server shutting down", "err", err.Error())
	}
}
