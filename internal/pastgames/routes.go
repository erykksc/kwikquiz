package pastgames

import (
	"net/http"
	"strings"
)

// NewPastGamesRouter sets up the routes for the pastgames package.
func NewPastGamesRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/past-games/", getPastGameHandler)
	return mux
}

func getPastGameHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the game ID from the URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 2 {
		http.NotFound(w, r)
		return
	}

	gameID := pathParts[1]
	r.URL.RawQuery = "game-id=" + gameID
	GetPastGameHandler(w, r)
}

