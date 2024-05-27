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
		// Send the lobby screen to the player
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
		initiator.Conn.WriteMessage(websocket.TextMessage, []byte("Username cannot be empty"))
		return errors.New("new username is empty")
	}

	// Check if game hasn't started yet
	if !l.StartedAt.IsZero() {
		initiator.Conn.WriteMessage(websocket.TextMessage, []byte("Game already started"))
		return errors.New("game already started")
	}

	// Create a set of all usernames in the lobby
	usernames := make(map[string]bool)
	for _, player := range l.Players {
		usernames[player.Username] = true
	}

	// Check if the username is already in the lobby
	if _, ok := usernames[event.Username]; ok {
		initiator.Conn.WriteMessage(websocket.TextMessage, []byte("This username is already in the lobby"))
		return errors.New("username already in the lobby")
	}

	// Update the player's username
	l.Players[initiator.ClientID] = Player{
		Conn:     initiator.Conn,
		Username: event.Username,
	}

	// Send the lobby screen to all players

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
