package pastgames

import (
	"github.com/erykksc/kwikquiz/internal/database"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/erykksc/kwikquiz/internal/common"
)

var pastGameTmpl = template.Must(template.ParseFiles("templates/pastgames/pastgame.html", common.BaseTmplPath))

var PastGamesRepo *GormPastGameRepository

// NewPastGamesRouter sets up the routes for the pastgames package.
func NewPastGamesRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/past-games/{gameID}", getPastGameHandler)

	PastGamesRepo = NewGormPastGameRepository(database.DB)

	//if common.DevMode() {
	//	// Add test past game
	//	PastGamesRepo.AddPastGame(ExamplePastGame1)
	//}

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

	pastGame, err := PastGamesRepo.GetPastGameByID(uint(id))
	if err != nil {
		if _, ok := err.(ErrPastGameNotFound); ok {
			http.Error(w, "Past game not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := pastGameTmpl.Execute(w, pastGame); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		slog.Error("Error rendering template", "err", err)
	}
}
