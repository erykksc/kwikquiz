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
	Username string
	Conn     *websocket.Conn
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
	if _, ok := l.Players[initiator.Username]; ok {
		// TODO: Player already in lobby, check if they want to reconnect
		// (the current connection is unresponsive)
		return errors.New("player already in lobby")
	}

	// If player is not in the lobby, send choose username form
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
		return nil
	}

	// Check if game hasn't started yet
	if !l.StartedAt.IsZero() {
		initiator.Conn.WriteMessage(websocket.TextMessage, []byte("Game already started"))
		return nil
	}

	// Check if the username is already in the lobby
	if _, ok := l.Players[initiator.Username]; ok {
		initiator.Conn.WriteMessage(websocket.TextMessage, []byte("This username is already in the lobby"))
		return nil
	}

	// Remove old username if it exists
	for oldUsername, conn := range l.Players {
		if conn == initiator.Conn {
			delete(l.Players, oldUsername)
			break
		}
	}

	// Add player to the lobby
	l.Players[event.Username] = initiator.Conn

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
