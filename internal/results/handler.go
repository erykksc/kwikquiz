package results

import (
	"html/template"
	"log/slog"
	"net/http"
	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/lobby"
)

const (
	LeaderboardTemplate = "templates/leaderboard.html"
	BaseTemplate        = "templates/base.html"
)

var LobbiesRepo lobby.LobbyRepository

func inc(i int) int {
	return i + 1
}

func GetLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	pin := r.URL.Query().Get("pin")
	if pin == "" {
		common.ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	lobbyInstance, err := LobbiesRepo.GetLobby(pin)
	if err != nil {
		if _, ok := err.(*lobby.ErrLobbyNotFound); ok {
			common.ErrorHandler(w, r, http.StatusNotFound)
			return
		}
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	leaderboard := lobbyInstance.Game.GetLeaderboard()

	tmpl := template.Must(template.New("leaderboard").Funcs(template.FuncMap{
		"inc": func(i int) int { return i + 1 },
	}).ParseFiles(LeaderboardTemplate, BaseTemplate))

	if err := tmpl.ExecuteTemplate(w, "base", leaderboard); err != nil {
		slog.Error("Error rendering template", "err", err)
	}
}

