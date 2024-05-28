package lobby

import (
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"

	"github.com/gorilla/websocket"
)

// Initiator is a user that initiates a lobby event
type Initiator struct {
	ClientID
	Conn *websocket.Conn
}

type LobbyEvent interface {
	String() string
	Handle(lobby *Lobby, initiator *Initiator) error
}

// ParseLobbyEvent parses a GameEvent from a JSON in a byte slice
func ParseLobbyEvent(data []byte) (LobbyEvent, error) {
	type Headers struct {
		EventType string
	}

	type WsRequest struct {
		HEADERS Headers
	}

	var wsRequest WsRequest
	if err := json.Unmarshal(data, &wsRequest); err != nil {
		return nil, err
	}

	switch wsRequest.HEADERS.EventType {
	case "LEUserConnected":
		var event LEUserConnected
		return event, nil
	case "LESubmittedUsername":
		var event LEUSubmittedUsername
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return event, nil
	case "LEGameStarted":
		var event LEGameStarted
		return event, nil
	default:
		return nil, errors.New("unknown event type")
	}
}

// LEUserConnected is a game event that is broadcasted when a user connects to the lobby
type LEUserConnected struct {
}

func (e LEUserConnected) String() string {
	return "GEUserConnected"
}

func (event LEUserConnected) Handle(l *Lobby, initiator *Initiator) error {
	slog.Debug("Handling User Connected", "event", event, "initiator", initiator)
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if the player is already in the lobby
	if player, ok := l.Players[initiator.ClientID]; ok {
		// Player already in the lobby, assuming they want to reconnect
		player.Conn = initiator.Conn
		// TODO: Send the current state screen to the player instead of the lobby screen every time
		tmpl := template.Must(template.ParseFiles(LobbyTemplate, BaseTemplate))
		w, err := initiator.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		defer w.Close()
		if err := tmpl.ExecuteTemplate(w, "player-lobby-screen", l); err != nil {
			return err
		}
		return nil
	}

	tmpl := template.Must(template.ParseFiles(LobbyTemplate, BaseTemplate))
	w, err := initiator.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	if err := tmpl.ExecuteTemplate(w, "username-form", nil); err != nil {
		return err
	}

	return nil
}

type LEUSubmittedUsername struct {
	Username string
}

func (e LEUSubmittedUsername) String() string {
	return "GEUserSubmittedUsername: " + e.Username
}

func (event LEUSubmittedUsername) Handle(l *Lobby, initiator *Initiator) error {
	slog.Debug("Handling Submitted Username", "event", event, "initiator", initiator)
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if the username is empty
	if event.Username == "" {
		// TODO: Send error message to the initiator: "Username cannot be empty"
		return errors.New("new username is empty")
	}

	// Check if game hasn't started yet
	if l.State != WaitingForPlayers {
		// TODO: Send error message to the initiator: "Game already started"
		return errors.New("game already started")
	}

	// Create a set of all usernames in the lobby
	usernames := make(map[string]bool)
	for _, player := range l.Players {
		usernames[player.Username] = true
	}

	// Check if the username is already in the lobby
	if _, ok := usernames[event.Username]; ok {
		// TODO: Send error message to the initiator: "Username already in the lobby"
		return errors.New("username already in the lobby")
	}

	// Update the player's username
	l.Players[initiator.ClientID] = &Player{
		Conn:     initiator.Conn,
		Username: event.Username,
	}

	// Send the lobby screen to all players
	// TODO: Make it more robust, so that if one player fails to receive the message, the others still do
	// maybe try with a retry mechanism
	tmpl := template.Must(template.ParseFiles(LobbyTemplate, BaseTemplate))
	for _, player := range l.Players {
		w, err := player.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}

		if err := tmpl.ExecuteTemplate(w, "player-lobby-screen", l); err != nil {
			return err
		}
		w.Close()
	}
	return nil
}

type LEGameStarted struct{}

func (e LEGameStarted) String() string {
	return "GEGameStarted"
}

func (event LEGameStarted) Handle(l *Lobby, initiator *Initiator) error {
	// Check if the initiator is the host
	if l.Host != initiator.ClientID {
		// TODO: Send error message to the initiator "Only the host can start the game"
		return errors.New("Non-host tried to start the game")
	}

	// Check if there are enough players
	if len(l.Players) == 0 {
		// TODO: Send error message to the initiator "Not enough players"
		return errors.New("Not enough players")
	}

	// Check if the game has already started
	if l.State != WaitingForPlayers {
		// TODO: Send the current state to the initiator
		return errors.New("Game already started")
	}

	// TODO: CHeck if the quiz is choosen

	// TODO: Check if the quiz has at least one question

	// Start game: go to the first question
	if err := l.StartNextQuestion(); err != nil {
		return err
	}

	return nil
}
