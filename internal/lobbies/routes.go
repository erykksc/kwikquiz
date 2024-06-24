package lobbies

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/quiz"
	"github.com/gorilla/websocket"
)

var lobbiesRepo lobbyRepository = newInMemoryLobbyRepository()

// Returns a handler for routes starting with /lobbies
func NewLobbiesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /lobbies/{$}", getLobbiesHandler)
	mux.HandleFunc("POST /lobbies/{$}", postLobbiesHandler)
	mux.HandleFunc("GET /lobbies/{pin}", getLobbyByPinHandler)
	mux.HandleFunc("/lobbies/{pin}/ws", getLobbyByPinWsHandler)
	mux.HandleFunc("/lobbies/{pin}/settings", lobbySettingsHandler)

	mux.HandleFunc("GET /lobbies/join", getLobbyJoinHandler)

	// Add test lobby
	if common.DevMode() {
		lOptions := lobbyOptions{
			TimePerQuestion: 30 * time.Second,
			Pin:             "1234",
			Quiz:            quiz.ExampleQuizGeography,
		}
		testLobby := createLobby(lOptions)
		lobbiesRepo.AddLobby(testLobby)
		slog.Info("Added test lobby", "lobby-pin", testLobby.Pin)

		// Question View Lobby
		// a lobby with fixed question-view, for design purposes
		lOptions.Pin = "0001"
		qwLobby := createLobby(lOptions)
		qwLobby.StartedAt = time.Now()
		qwLobby.CurrentQuestionStartTime = time.Now()
		qwLobby.CurrentQuestion = &quiz.ExampleQuizGeography.Questions[0]
		qwLobby.CurrentQuestionIdx = 0
		qwLobby.CurrentQuestionTimeout = time.Now().Add(30 * time.Second)
		qwLobby.ReadingTimeout = time.Now()
		qwLobby.State = LsQuestion
		qwLobby.PlayersAnswering = 3
		lobbiesRepo.AddLobby(qwLobby)
		slog.Info("Added question view lobby", "lobby-pin", qwLobby.Pin)

		// Question View Reading Lobby
		// a lobby with fixed question-view, for design purposes
		lOptions.Pin = "0002"
		qwrLobby := createLobby(lOptions)
		qwrLobby.StartedAt = time.Now()
		qwrLobby.CurrentQuestionStartTime = time.Now()
		qwrLobby.CurrentQuestion = &quiz.ExampleQuizGeography.Questions[0]
		qwrLobby.CurrentQuestionIdx = 0
		qwrLobby.CurrentQuestionTimeout = time.Now().Add(300 * time.Second)
		qwrLobby.ReadingTimeout = time.Now().Add(100 * time.Second)
		qwrLobby.State = LsQuestion
		qwrLobby.PlayersAnswering = 3
		lobbiesRepo.AddLobby(qwrLobby)
		slog.Info("Added question view lobby during reading", "lobby-pin", qwrLobby.Pin)
	}

	return mux
}

// getClientIDFromRequest returns the clientID from the request cookie
func getClientIDFromRequest(r *http.Request) (ClientID, error) {
	clientIDCookie, err := r.Cookie("client-id")
	if err == http.ErrNoCookie {
		return "", err
	}
	return ClientID(clientIDCookie.Value), nil
}

// TODO: Make it only accessible by admin
func getLobbiesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	lobbies, err := lobbiesRepo.GetAllLobbies()
	if err != nil {
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if err := LobbiesTmpl.Execute(w, lobbies); err != nil {
		slog.Error("Error rendering template", "err", err)
	}
}

func postLobbiesHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the client isn't a host of another lobby
	clientID, err := getClientIDFromRequest(r)
	if err == nil {
		lobby, err := lobbiesRepo.GetLobbyByHost(clientID)
		if err == nil {
			// Redirect to the lobby
			w.Header().Add("HX-Redirect", "/lobbies/"+lobby.Pin)
			w.WriteHeader(http.StatusFound)
			return
		}
	}
	// Otherwise, create a new lobby

	// TODO: Parse possible arguments
	options := lobbyOptions{}
	newLobby := createLobby(options)
	lobbiesRepo.AddLobby(newLobby)
	slog.Info("Created new lobby", "lobby", newLobby)
	// Redirect to the new lobby
	w.Header().Add("HX-Redirect", "/lobbies/"+newLobby.Pin)
	w.WriteHeader(http.StatusCreated)
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
	cID, err := getClientIDFromRequest(r)
	if err == http.ErrNoCookie {
		// Set new client id if not present
		cID, err = NewClientID()
		if err != nil {
			slog.Error("Error generating new client id", "err", err)
			common.ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	// SET CLIENT ID COOKIE or UPDATE EXPIRATION
	http.SetCookie(w, &http.Cookie{
		Name:  "client-id",
		Value: string(cID),
	})

	if err := LobbyTmpl.Execute(w, &lobby); err != nil {
		slog.Error("Error rendering template", "err", err)
	}
}

// getLobbyByPinWsHandler handles requests to /lobbies/{pin}/ws
func getLobbyByPinWsHandler(w http.ResponseWriter, r *http.Request) {
	clientID, err := getClientIDFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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

	slog.Debug("Handling new ws connection", "clientID", clientID, "Lobby-Pin", lobby.Pin)
	lobby.mu.Lock()
	user, err := handleNewWebsocketConn(lobby, ws, clientID)
	lobby.mu.Unlock()
	if err != nil {
		slog.Error("Error handling user connection", "err", err)
		return
	}

	// HANDLE REQUESTS
	for {
		messageType, message, err := ws.ReadMessage()
		// Handle disconnection
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Info("Client disconnected from websocket", "clientID", user)
			} else if strings.Contains(err.Error(), "use of closed network connection") {
				slog.Info("Server closed the connection", "clientID", user)
			} else {
				slog.Error("Unexpected error while reading ws message, disconnecting", "err", err)
			}
			break
		}
		if messageType != websocket.TextMessage {
			slog.Warn("Received non-text message", "messageType", messageType, "ws", ws)
			continue
		}

		event, err := parseLobbyEvent(message)
		if err != nil {
			slog.Warn("Error parsing lobby event, skipping", "err", err, "message", message)
			continue
		}

		slog.Info("Handling lobby event", "event", event.String(), "initiator", user)

		lobby.mu.Lock()
		if err := event.Handle(lobby, user); err != nil {
			slog.Error("Error handling lobby event", "event", event, "err", err)
		}
		lobby.mu.Unlock()
	}
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
		common.JoinFormTmpl.Execute(w, common.JoinFormData{GamePinError: "Game not found"})
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

// Handler used for getting/updating the lobby settings from the waiting room
func lobbySettingsHandler(w http.ResponseWriter, r *http.Request) {
	clientID, err := getClientIDFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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
	// Check if the client is the host
	if lobby.Host == nil || lobby.Host.ClientID != clientID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Only update settings if the method is PUT
	if r.Method == "PUT" {
		timePerQuestionStr := r.FormValue("time-per-question")
		if timePerQuestionStr != "" {
			timePerQuestion, err := time.ParseDuration(timePerQuestionStr + "s")
			if err != nil {
				slog.Error("Error parsing time-per-question", "err", err)
				common.ErrorHandler(w, r, http.StatusBadRequest)
				return
			}
			lobby.TimePerQuestion = timePerQuestion
			slog.Debug("Updated time-per-question", "lobby.Pin", lobby.Pin, "timePerQuestion", timePerQuestion.String())
		}

		timeForReadingStr := r.FormValue("time-for-reading")
		if timeForReadingStr != "" {
			timeForReading, err := time.ParseDuration(timeForReadingStr + "s")
			if err != nil {
				slog.Error("Error parsing time-for-reading", "err", err)
				common.ErrorHandler(w, r, http.StatusBadRequest)
				return
			}
			lobby.TimeForReading = timeForReading
			slog.Debug("Updated time-for-reading", "lobby.Pin", lobby.Pin, "timeForReading", timeForReading.String())
		}

		quizIDStr := r.FormValue("quiz")
		if quizIDStr != "" {
			quizID, err := strconv.Atoi(quizIDStr)
			if err != nil {
				slog.Error("Error parsing quizID", "err", err)
				common.ErrorHandler(w, r, http.StatusBadRequest)
				return
			}
			quiz, err := quiz.QuizzesRepo.GetQuiz(quizID)
			if err != nil {
				slog.Error("Error getting quiz", "err", err)
				common.ErrorHandler(w, r, http.StatusBadRequest)
				return
			}
			lobby.Quiz = quiz
			slog.Debug("Updated quiz", "lobby.Pin", lobby.Pin, "quizID", quizIDStr, "quiz.Title", lobby.Quiz.Title)
		}
	}

	quizzesMeta, err := quiz.QuizzesRepo.GetAllQuizzesMetadata()
	if err != nil {
		slog.Error("Error getting quizzes metadata", "err", err)
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	err = LobbySettingsTmpl.Execute(w, LobbySettingsData{
		Quizzes: quizzesMeta,
		Lobby:   lobby,
	})
	if err != nil {
		slog.Error("Error rendering template", "err", err)
	}
}
