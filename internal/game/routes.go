package game

import (
	"github.com/erykksc/kwikquiz/internal/common"
	"html/template"
	"log/slog"
	"net/http"
)

const (
	NotFoundPage        = "static/notfound.html"
	BaseTemplate        = "templates/base.html"
	GamesTemplate       = "templates/games.html"
	GamesGidTemplate    = "templates/games-gid.html"
	IndexTemplate       = "templates/index.html"
	GamesCreateTemplate = "templates/games-create.html"
)

var gamesRepo GameRepository = NewInMemoryGameRepository()
var DEBUG = common.DebugOn()

// Returns a handler for routes starting with /games
func NewGamesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /games/{$}", getGamesHandler)
	mux.HandleFunc("GET /games/{gid}", getGamesGidHandler)
	mux.HandleFunc("POST /games/{$}", postGamesHandler)
	mux.HandleFunc("PUT /games/{gid}", putGameGidHandler)
	mux.HandleFunc("DELETE /games/{gid}", deleteGamesGidHandler)

	mux.HandleFunc("GET /games/join", getGamesJoinHandler)
	mux.HandleFunc("GET /games/create", getGamesCreateHandler)

	// TODO: Remove this after testing
	if DEBUG {
		gamesRepo.AddGame(Game{
			ID:       "1234",
			Host:     User{Name: "erykk"},
			QuizName: "Test Quiz",
		})
		slog.Info("Adding test game", "gid", "1234")
	}

	return mux
}

func getGamesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	games, err := gamesRepo.GetAllGames()
	if err != nil {
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles(GamesTemplate, BaseTemplate))

	tmpl.Execute(w, games)
}

func getGamesGidHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	gid := r.PathValue("gid")

	game, err := gamesRepo.GetGame(GameID(gid))
	if err != nil {
		switch err.(type) {
		case ErrGameNotFound:
			common.ErrorHandler(w, r, http.StatusNotFound)
			return
		default:
			common.ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	tmpl := template.Must(template.ParseFiles(GamesGidTemplate, BaseTemplate))
	tmpl.Execute(w, game)
}

type createGameForm struct {
	Gid       string
	Username  string
	Quizname  string
	FormError string
}

func postGamesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	gid := r.FormValue("gid") // Game ID
	username := r.FormValue("username")
	quizname := r.FormValue("quizname")

	if gid == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("gid in form is required"))
		return
	}

	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("username in form is required"))
		return
	}

	if quizname == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("quizname in form is required"))
		return
	}

	// Create a new game
	game := Game{
		ID:       GameID(gid),
		Host:     User{Name: username},
		QuizName: quizname,
	}

	if err := gamesRepo.AddGame(game); err != nil {
		slog.Error("Error adding game", "error", err)
		tmpl := template.Must(template.ParseFiles(GamesCreateTemplate, BaseTemplate))
		err = tmpl.ExecuteTemplate(w, "create-form", createGameForm{
			Gid:       gid,
			Username:  username,
			Quizname:  quizname,
			FormError: err.Error(),
		})
		if err != nil {
			slog.Error("Error rendering template", "error", err)
		}
		return
	}

	// Redirect to the game
	w.Header().Add("HX-Redirect", "/games/"+gid)
	w.WriteHeader(http.StatusCreated)
}

func putGameGidHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	common.ErrorHandler(w, r, http.StatusNotImplemented)
}

func deleteGamesGidHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	common.ErrorHandler(w, r, http.StatusNotImplemented)
}

type joinFormData struct {
	GamePinError string
}

func getGamesJoinHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	gid := r.URL.Query().Get("gid")
	if gid == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("gid in form is required"))
		return
	}

	_, err := gamesRepo.GetGame(GameID(gid))
	switch err.(type) {
	case nil:
		// Do nothing
	case ErrGameNotFound:
		w.WriteHeader(http.StatusNotFound)
		tmpl := template.Must(template.ParseFiles(IndexTemplate, BaseTemplate))
		tmpl.ExecuteTemplate(w, "join-form", joinFormData{GamePinError: "Game not found"})
		return
	default:
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Redirect", "/games/"+gid)

	// Redirect to the game if it's not an HX request
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/games/"+gid, http.StatusFound)
	}
}

func getGamesCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	tmpl := template.Must(template.ParseFiles(GamesCreateTemplate, BaseTemplate))
	tmpl.Execute(w, nil)
}
