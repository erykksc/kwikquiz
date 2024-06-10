package lobbies

import (
	"github.com/erykksc/kwikquiz/internal/quiz"
	"log/slog"
	"net/http"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/gorilla/websocket"
)

var lobbiesRepo lobbyRepository = newInMemoryLobbyRepository()

// Returns a handler for routes starting with /games
func NewLobbiesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /lobbies/{$}", getLobbiesHandler)
	mux.HandleFunc("GET /lobbies/{pin}", getLobbyByPinHandler)
	mux.HandleFunc("/lobbies/{pin}/ws", getLobbyByPinWsHandler)

	mux.HandleFunc("GET /lobbies/join", getLobbyJoinHandler)
	mux.HandleFunc("GET /lobbies/create", getLobbyCreateHandler)
	mux.HandleFunc("POST /lobbies/create", postLobbyCreateHandler)

	// Add test lobby
	if common.DevMode() {
		lOptions := lobbyOptions{
			TimePerQuestion: 30 * time.Second,
			Pin:             "1234",
		}
		testLobby := createLobby(lOptions)
		testLobby.Quiz = &quiz.Quiz{
			Title:       "Geography",
			Description: "This is a quiz about capitals around the world",
			Questions: []*quiz.Question{
				{
					Text: "What is the capital of France?",
					Answers: []*quiz.Answer{
						{Text: "Paris", IsCorrect: true},
						{Text: "Berlin", IsCorrect: false},
						{Text: "Warsaw", IsCorrect: false},
						{Text: "Barcelona", IsCorrect: false},
					},
				},
				{
					Text: "On which continent is Russia?",
					Answers: []*quiz.Answer{
						{Text: "Europe", IsCorrect: true},
						{Text: "Asia", IsCorrect: true},
						{Text: "North America", IsCorrect: false},
						{Text: "South America", IsCorrect: false},
					},
				},
			},
		}

		lobbiesRepo.AddLobby(testLobby)
		slog.Info("Adding test lobby", "lobby", testLobby)
	}

	return mux
}

// TODO: Make it only accessible by admin
func getLobbiesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	lobbies, err := lobbiesRepo.GetAllLobbies()
	if err != nil {
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if err := lobbiesTmpl.Execute(w, lobbies); err != nil {
		slog.Error("Error rendering template", "err", err)
	}
}

// getLobbyByPinHandler handles requests to /lobbies/{pin}
func getLobbyByPinHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	pin := r.PathValue("pin")

	lobby, err := lobbiesRepo.GetLobby(pin)
	if err != nil {
		switch err.(type) {
		case errLobbyNotFound:
			common.ErrorHandler(w, r, http.StatusNotFound)
			return
		default:
			common.ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	// GET CLIENT ID from COOKIE
	var cID clientID
	clientIDCookie, err := r.Cookie("client-id")
	if err == http.ErrNoCookie {
		// Generate new client id
		cID, err = newClientID()
		if err != nil {
			slog.Error("Error generating new client id", "err", err)
			common.ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	} else {
		cID = clientID(clientIDCookie.Value)
	}

	// SET CLIENT ID COOKIE or UPDATE EXPIRATION
	http.SetCookie(w, &http.Cookie{
		Name:    "client-id",
		Value:   string(cID),
		Expires: time.Now().Add(6 * time.Hour),
	})

	if err := lobbyTmpl.Execute(w, &lobby); err != nil {
		slog.Error("Error rendering template", "err", err)
	}
}

// getLobbyByPinWsHandler handles requests to /lobbies/{pin}/ws
func getLobbyByPinWsHandler(w http.ResponseWriter, r *http.Request) {
	pin := r.PathValue("pin")

	lobby, err := lobbiesRepo.GetLobby(pin)
	switch err.(type) {
	case nil:
		break
	case errLobbyNotFound:
		slog.Error("Error trying to connect to not existing lobby", "err", err)
		common.ErrorHandler(w, r, http.StatusNotFound)
		return
	default:
		slog.Error("Error getting lobby", "err", err)
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// UPGRADE CONNECTION TO WEBSOCKET
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 4 * 1024,
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade to websocket", "err", err)
		return
	}

	// GET CLIENT ID from COOKIE
	clientIDCookie, err := r.Cookie("client-id")
	if err == http.ErrNoCookie {
		slog.Error("Client ID cookie not found")
		common.ErrorHandler(w, r, http.StatusForbidden)
		return
	}
	clientID := clientID(clientIDCookie.Value)

	slog.Debug("Handling new ws connection", "clientID", clientID, "Lobby-Pin", lobby.Pin)
	user, err := handleNewWebsocketConn(lobby, ws, clientID)
	if err != nil {
		slog.Error("Error handling user connection", "err", err)
		return
	}

	// HANDLE REQUESTS
	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			slog.Error("Error reading ws message", "err", err)
			break
		}
		if messageType != websocket.TextMessage {
			slog.Warn("Received non-text message", "messageType", messageType, "ws", ws)
			continue
		}

		event, err := parseLobbyEvent(message)
		if err != nil {
			slog.Error("Error parsing lobby event", "err", err)
			continue
		}

		slog.Info("Handling lobby event", "event", event.String(), "initiator", user)

		if err := event.Handle(lobby, user); err != nil {
			slog.Error("Error handling lobby event", "event", event, "err", err)
		}
	}
}

type joinFormData struct {
	GamePinError string
}

// getLobbyJoinHandler handles requests to /lobbies/join
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
	case errLobbyNotFound:
		w.WriteHeader(http.StatusNotFound)
		common.IndexTmpl.ExecuteTemplate(w, "join-form", joinFormData{GamePinError: "Game not found"})
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

// getLobbyCreateHandler handles requests to /lobbies/create
func getLobbyCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	if err := lobbyCreateTmpl.Execute(w, nil); err != nil {
		slog.Error("Error rendering template", "error", err)
	}
}

type createGameForm struct {
	Pin       string
	Username  string
	FormError string
}

// postLobbyCreateHandler handles requests to POST /lobbies/create
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

	newLobby := createLobby(lobbyOptions{})

	if err := lobbiesRepo.AddLobby(newLobby); err != nil {
		slog.Error("Error adding game", "error", err)
		err = lobbyCreateTmpl.ExecuteTemplate(w, "create-form", createGameForm{
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
