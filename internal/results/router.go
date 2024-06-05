package results

import (
    "net/http"
)
	
func NewResultsRouter() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/{pin}", GetLeaderboardHandler)
    return mux
}
