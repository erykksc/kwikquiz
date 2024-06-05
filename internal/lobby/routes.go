package lobby

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/gorilla/websocket"
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
func NewLobbiesRouter() http.Handler {
	mux := http.NewServeMux()

	/*mux.HandleFunc("GET /lobbies/{$}", getLobbiesHandler)
	mux.HandleFunc("GET /lobbies/{pin}", getLobbyByPinHandler)
	mux.HandleFunc("/lobbies/{pin}/ws", getLobbyByPinWsHandler)

	mux.HandleFunc("GET /lobbies/join", getLobbyJoinHandler)
	mux.HandleFunc("GET /lobbies/create", getLobbyCreateHandler)
	mux.HandleFunc("POST /lobbies/create", postLobbyCreateHandler)*/

	mux.HandleFunc("/lobbies", getLobbiesHandler)
	mux.HandleFunc("/lobbies/", getLobbyByPinHandler)
	mux.HandleFunc("/lobbies/ws", getLobbyByPinWsHandler)

	mux.HandleFunc("/lobbies/join", getLobbyJoinHandler)
	mux.HandleFunc("/lobbies/create", getLobbyCreateHandler)
	mux.HandleFunc("POST /lobbies/create", postLobbyCreateHandler)

	// Add test lobby
	if DEBUG {
		lOptions := LobbyOptions{
			TimePerQuestion: 30 * time.Second,
			Pin:             "1234",
		}
		testLobby := CreateLobby(lOptions)

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
	tmpl := template.Must(template.ParseFiles(LobbiesTemplate, BaseTemplate))

	if err := tmpl.Execute(w, lobbies); err != nil {
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
		case *ErrLobbyNotFound:
			common.ErrorHandler(w, r, http.StatusNotFound)
			return
		default:
			common.ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	// GET CLIENT ID from COOKIE
	var clientID ClientID
	clientIDCookie, err := r.Cookie("client-id")
	if err == http.ErrNoCookie {
		// Generate new client id
		clientID, err = NewClientID()
		if err != nil {
			slog.Error("Error generating new client id", "err", err)
			common.ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	} else {
		clientID = ClientID(clientIDCookie.Value)
	}

	// SET CLIENT ID COOKIE or UPDATE EXPIRATION
	http.SetCookie(w, &http.Cookie{
		Name:    "client-id",
		Value:   string(clientID),
		Expires: time.Now().Add(6 * time.Hour),
	})

	tmpl := template.Must(template.ParseFiles(LobbyTemplate, BaseTemplate))
	if err := tmpl.Execute(w, &lobby); err != nil {
		slog.Error("Error rendering template", "err", err)
	}
}

// getLobbyByPinWsHandler handles requests to /lobbies/{pin}/ws
func getLobbyByPinWsHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handling request", "method", r.Method, "path", r.URL.Path)
	//pin := r.PathValue("pin")
	pin := r.URL.Path[len("/lobbies/"):]

	lobby, err := lobbiesRepo.GetLobby(pin)
	switch err.(type) {
	case nil:
		break
	case *ErrLobbyNotFound:
		slog.Error("Error trying to connet to not existing lobby", "err", err)
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
	slog.Info("Upgrading connection")
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
	clientID := ClientID(clientIDCookie.Value)

	player := Initiator{ClientID: clientID, Conn: ws}

	(LEUserConnected{}).Handle(lobby, &player)

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
		slog.Debug("Received ws message", "message", string(message))

		event, err := ParseLobbyEvent(message)
		if err != nil {
			slog.Error("Error parsing lobby event", "err", err)
			continue
		}

		if err := event.Handle(lobby, &player); err != nil {
			slog.Error("Error handling lobby event", "err", err)
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
	case *ErrLobbyNotFound:
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

// getLobbyCreateHandler handles requests to /lobbies/create
func getLobbyCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	tmpl := template.Must(template.ParseFiles(LobbyCreateTemplate, BaseTemplate))
	if err := tmpl.Execute(w, nil); err != nil {
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

	lobby := CreateLobby(LobbyOptions{})

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
