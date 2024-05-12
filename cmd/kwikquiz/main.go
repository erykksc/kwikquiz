package main

import (
	"fmt"
	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/game"
	"log/slog"
	"net/http"
	"os"
)

var DEBUG = common.DebugOn()

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info(fmt.Sprintf("%s %s", r.Method, r.URL.Path))

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func getLoggingHandler(level slog.Leveler) slog.Handler {
	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)

	return handler
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

	fs := http.FileServer(http.Dir("static"))

	router := http.NewServeMux()

	router.Handle("/games/", game.NewGamesRouter())
	router.HandleFunc("/{$}", common.IndexHandler)
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	port := 3000
	addr := fmt.Sprintf(":%d", port)
	slog.Info("Server listening", "addr", addr)

	err := http.ListenAndServe(addr, loggingMiddleware(router))
	if err != nil {
		slog.Error("Server shutting down", "err", err.Error())
	}
}
