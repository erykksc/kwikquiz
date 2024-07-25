package pastgames

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/jmoiron/sqlx"
)

var pastGameTmpl = common.ParseTmplWithFuncs("templates/pastgames/pastgame.html")

var pastGamesListTmpl = template.Must(template.ParseFiles("templates/pastgames/search_pastgames.html", common.BaseTmplPath))

var Repo Repository

func InitRepo(db *sqlx.DB) error {
	sqliteRepo := NewPastGameRepositorySQLite(db)
	err := sqliteRepo.Initialize()
	if err != nil {
		return err
	}

	Repo = sqliteRepo
	return nil
}

// NewPastGamesRouter sets up the routes for the pastgames package.
func NewPastGamesRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/past-games/{gameID}", getPastGameHandler)
	mux.HandleFunc("/past-games/{$}", browsePastGamesHandler)

	return mux
}

func getPastGameHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the game ID from the URL
	gameID := r.PathValue("gameID")

	id, err := strconv.Atoi(gameID)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	pastGame, err := Repo.GetByID(int64(id))
	if err != nil {
		if _, ok := err.(ErrPastGameNotFound); ok {
			http.Error(w, "Past game not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		slog.Error("Error getting past game", "err", err)
		return
	}

	err = Repo.HydrateScores(pastGame)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		slog.Error("Error hydrating past game scores", "err", err)
		return
	}

	if err := pastGameTmpl.Execute(w, pastGame); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		slog.Error("Error rendering template", "err", err)
	}
}

func browsePastGamesHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var pastGames []PastGame
	var err error
	if query != "" {
		pastGames, err = Repo.BrowsePastGamesByID(query)
	} else {
		pastGames, err = Repo.GetAll()
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		slog.Error("Error searching past games", "err", err)
		return
	}

	data := struct {
		Query string
		Games []PastGame
	}{
		Query: query,
		Games: pastGames,
	}

	if err := pastGamesListTmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		slog.Error("Error rendering template", "err", err)
	}
}
