package pastgames

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/erykksc/kwikquiz/internal/common"
)

const (
	PastGameTemplate = "templates/pastgames/pastgame.html"
	BaseTemplate     = "templates/base.html"
)

var pastGameRepo PastGameRepository = NewInMemoryPastGameRepository()

// NewPastGamesRouter sets up the routes for the pastgames package.
func NewPastGamesRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/past-games/{gameID}", getPastGameHandler)

	if common.DevMode() {
		// Add test past game
		pastGameRepo.AddPastGame(ExamplePastGame1)
	}

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

	pastGame, err := pastGameRepo.GetPastGameByID(id)
	if err != nil {
		if _, ok := err.(ErrPastGameNotFound); ok {
			http.Error(w, "Past game not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(PastGameTemplate, BaseTemplate))
	if err := tmpl.Execute(w, pastGame); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		slog.Error("Error rendering template", "err", err)
	}
}
