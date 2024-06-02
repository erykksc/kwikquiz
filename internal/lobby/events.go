package lobby

import (
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
)

type LobbyEvent interface {
	String() string
	Handle(lobby *Lobby, initiator *User) error
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
	case "LEGameStartRequest":
		var event LEGameStartRequest
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

func (event LEUserConnected) Handle(l *Lobby, initiator *User) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if it is the first user, if so, he becomes the host
	if l.Host.ClientID == "" {
		slog.Info("New Host for Lobby", "Lobby-Pin", l.Pin, "Client-ID", initiator.ClientID)
		l.Host.ClientID = initiator.ClientID
		l.Host.Conn = initiator.Conn
		viewData := ViewData{
			Lobby:  l,
			Player: l.Host,
			IsHost: true,
		}
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		err := l.Host.WriteTemplate(tmpl, WaitingRoomView, viewData)
		if err != nil {
			return err
		}
		return nil
	}

	// Check if host is trying to reconnect
	if l.Host.ClientID == initiator.ClientID {
		// Host reconnecting
		l.Host.Conn = initiator.Conn
		viewData := ViewData{
			Lobby:  l,
			Player: l.Host,
			IsHost: true,
		}
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		err := l.Host.WriteTemplate(tmpl, l.State.ViewName(), viewData)
		if err != nil {
			return err
		}
		slog.Info("Host Reconnected", "Lobby-Pin", l.Pin, "Client-ID", initiator.ClientID)
		return nil
	}

	// Check if the ClientID is already in the lobby
	if player, ok := l.Players[initiator.ClientID]; ok {
		slog.Info("Player Reconnected", "Lobby-Pin", l.Pin, "Client-ID", initiator.ClientID)
		// Player already in the lobby, assuming they want to reconnect
		player.Conn = initiator.Conn
		viewData := ViewData{
			Lobby:  l,
			Player: *player,
			IsHost: false,
		}
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		err := player.WriteTemplate(tmpl, l.State.ViewName(), viewData)
		if err != nil {
			return err
		}
		return nil
	}

	// New User connecting
	slog.Info("New User for Lobby", "Lobby-Pin", l.Pin, "Client-ID", initiator.ClientID)

	viewData := ViewData{
		Lobby:  l,
		Player: *initiator,
		IsHost: false,
	}
	tmpl := template.Must(template.ParseFiles(LobbyTemplate))
	err := initiator.WriteTemplate(tmpl, ChooseUsernameView, viewData)
	if err != nil {
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

func (event LEUSubmittedUsername) Handle(l *Lobby, initiator *User) error {
	slog.Debug("Handling Submitted Username", "event", event, "initiator", initiator)
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if the username is empty
	if event.Username == "" {
		// TODO: Send error message to the initiator: "Username cannot be empty"
		return errors.New("new username is empty")
	}

	// Check if game hasn't started yet
	if l.State != LSWaitingForPlayers {
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
	l.Players[initiator.ClientID] = &User{
		Conn:     initiator.Conn,
		Username: event.Username,
	}

	// Send the lobby screen to all players
	viewData := ViewData{
		Lobby:  l,
		Player: *initiator,
		IsHost: true,
	}
	tmpl := template.Must(template.ParseFiles(LobbyTemplate))
	if err := l.Host.WriteTemplate(tmpl, WaitingRoomView, viewData); err != nil {
		slog.Error("Error writing view to host", "error", err)
	}

	viewData.IsHost = false
	for _, player := range l.Players {
		viewData.Player = *player
		if err := player.WriteTemplate(tmpl, WaitingRoomView, viewData); err != nil {
			slog.Error("Error writing view to player", "error", err)
		}
	}
	return nil
}

type LEGameStartRequest struct{}

func (e LEGameStartRequest) String() string {
	return "GEGameStarted"
}

func (event LEGameStartRequest) Handle(l *Lobby, initiator *User) error {
	// Check if the initiator is the host
	if l.Host.ClientID != initiator.ClientID {
		// TODO: Send error message to the initiator "Only the host can start the game"
		return errors.New("Non-host tried to start the game")
	}

	// Check if there are enough players
	if len(l.Players) == 0 {
		// TODO: Send error message to the initiator "Not enough players"
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		initiator.WriteTemplate(tmpl, ErrorAlert, "Not enough players")
		return errors.New("Not enough players")
	}

	// Check if the game has already started
	if l.State != LSWaitingForPlayers {
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
