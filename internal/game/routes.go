package game

import (
	"github.com/erykksc/kwikquiz/internal/common"
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

const (
	NotFoundPage        = "static/notfound.html"
	BaseTemplate        = "templates/base.html"
	IndexTemplate       = "templates/index.html"
	LobbiesTemplate     = "templates/lobbies.html"
	LobbyTemplate       = "templates/lobby.html"
	LobbyCreateTemplate = "templates/lobby-create.html"
)

var lobbiesRepo LobbyRepository = NewInMemoryLobbyRepository()
var DEBUG = common.DebugOn()

// Returns a handler for routes starting with /games
func NewGamesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /lobbies/{$}", getLobbiesHandler)
	mux.HandleFunc("GET /lobbies/{pin}", getLobbyByPinHandler)

	mux.HandleFunc("GET /lobbies/join", getLobbyJoinHandler)
	mux.HandleFunc("GET /lobbies/create", getLobbyCreateHandler)
	mux.HandleFunc("POST /lobbies/create", postLobbyCreateHandler)

	// TODO: Remove this after testing
	if DEBUG {
		testLobby := Lobby{
			Pin:       "1234",
			CreatedAt: time.Now(),
			Game: Game{
				Hostname:        "erykk",
				TimePerQuestion: 30 * time.Second,
			},
		}
		lobbiesRepo.AddLobby(testLobby)
		slog.Info("Adding test lobby", "lobby", testLobby)
	}

	return mux
}

func getLobbiesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	games, err := lobbiesRepo.GetAllLobbies()
	if err != nil {
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles(LobbiesTemplate, BaseTemplate))

	tmpl.Execute(w, games)
}

func getLobbyByPinHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	pin := r.PathValue("pin")

	game, err := lobbiesRepo.GetLobby(pin)
	if err != nil {
		switch err.(type) {
		case ErrLobbyNotFound:
			common.ErrorHandler(w, r, http.StatusNotFound)
			return
		default:
			common.ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	tmpl := template.Must(template.ParseFiles(LobbyTemplate, BaseTemplate))
	tmpl.Execute(w, game)
}

type joinFormData struct {
	GamePinError string
}

func getLobbyJoinHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	pin := r.URL.Query().Get("pin")
	if pin == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("pin in form is required"))
		return
	}

	_, err := lobbiesRepo.GetLobby(pin)
	switch err.(type) {
	case nil:
		// Do nothing
	case ErrLobbyNotFound:
		w.WriteHeader(http.StatusNotFound)
		tmpl := template.Must(template.ParseFiles(IndexTemplate, BaseTemplate))
		tmpl.ExecuteTemplate(w, "join-form", joinFormData{GamePinError: "Game not found"})
		return
	default:
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Redirect", "/lobbies/"+pin)

	// Redirect to the game if it's not an HX request
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/lobbies/"+pin, http.StatusFound)
	}
}

func getLobbyCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	tmpl := template.Must(template.ParseFiles(LobbyCreateTemplate, BaseTemplate))
	tmpl.Execute(w, nil)
}

type createGameForm struct {
	Pin       string
	Username  string
	FormError string
}

func postLobbyCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	pin := r.FormValue("pin") // Game Pin
	username := r.FormValue("username")

	if pin == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("pin in form is required"))
		return
	}

	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("username in form is required"))
		return
	}

	// Create a new game
	game := Game{
		Hostname:        username,
		TimePerQuestion: 30 * time.Second,
	}

	lobby := Lobby{
		Pin:       pin,
		Game:      game,
		CreatedAt: time.Now(),
	}

	if err := lobbiesRepo.AddLobby(lobby); err != nil {
		slog.Error("Error adding game", "error", err)
		tmpl := template.Must(template.ParseFiles(LobbyCreateTemplate, BaseTemplate))
		err = tmpl.ExecuteTemplate(w, "create-form", createGameForm{
			Pin:       pin,
			Username:  username,
			FormError: err.Error(),
		})
		if err != nil {
			slog.Error("Error rendering template", "error", err)
		}
		return
	}

	// Redirect to the game
	w.Header().Add("HX-Redirect", "/lobbies/"+pin)
	w.WriteHeader(http.StatusCreated)
}
