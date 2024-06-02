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
	case "LEChangeUsernameRequest":
		var event LEChangeUsernameRequest
		return event, nil
	default:
		return nil, errors.New("unknown event type")
	}
}

// LEUserConnected is a game event that is broadcasted when a user connects to the lobby
type LEUserConnected struct {
}

func (e LEUserConnected) String() string {
	return "LEUserConnected"
}

func (event LEUserConnected) Handle(l *Lobby, initiator *User) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	viewData := ViewData{
		Lobby:  l,
		Player: *initiator,
		IsHost: false,
	}

	player, initiatorIsPlayer := l.Players[initiator.ClientID]
	switch {
	// Check if it is the first user, if so, he becomes the host
	case l.Host == nil:
		slog.Info("New Host for the Lobby", "Lobby-Pin", l.Pin, "Client-ID", initiator.ClientID)
		l.Host = initiator
		viewData.IsHost = true

	// Check if host is trying to reconnect
	case l.Host.ClientID == initiator.ClientID:
		slog.Info("Host reconnecting", "Lobby-Pin", l.Pin, "Client-ID", initiator.ClientID)
		l.Host.Conn = initiator.Conn
		viewData.IsHost = true

	// Check if player is trying to reconnect
	case initiatorIsPlayer:
		slog.Info("Player reconnecting", "Lobby-Pin", l.Pin, "Client-ID", initiator.ClientID)
		player.Conn = initiator.Conn
		initiator = player

	// New User connecting
	default:
		slog.Info("New Player for Lobby", "Lobby-Pin", l.Pin, "Client-ID", initiator.ClientID)
		l.Players[initiator.ClientID] = initiator
	}

	tmpl := template.Must(template.ParseFiles(LobbyTemplate))
	err := initiator.WriteTemplate(tmpl, l.State.ViewName(), viewData)
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
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		initiator.WriteTemplate(tmpl, ErrorAlert, "Username cannot be empty")
		return errors.New("new username is empty")
	}

	// Check if game hasn't started yet
	if l.State != LSWaitingForPlayers {
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		initiator.WriteTemplate(tmpl, ErrorAlert, "Game already started")
		return errors.New("game already started")
	}

	// Check if new username isn't the same as the old one
	if l.Players[initiator.ClientID].Username == event.Username {
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		viewData := ViewData{
			Lobby:  l,
			Player: *initiator,
			IsHost: l.Host.ClientID == initiator.ClientID,
		}
		if err := initiator.WriteTemplate(tmpl, WaitingRoomView, viewData); err != nil {
			return err
		}
		return nil
	}

	// Create a set of all usernames in the lobby
	usernames := make(map[string]bool)
	for _, player := range l.Players {
		usernames[player.Username] = true
	}

	// Check if the username is already in the lobby
	if _, ok := usernames[event.Username]; ok {
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		initiator.WriteTemplate(tmpl, ErrorAlert, "Username already in the lobby")
		return errors.New("username already in the lobby")
	}

	// Update the player's username
	initiator.Username = event.Username

	// l.Players[initiator.ClientID] = &User{
	// 	Conn:     initiator.Conn,
	// 	Username: event.Username,
	// }

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

type LEChangeUsernameRequest struct{}

func (e LEChangeUsernameRequest) String() string {
	return "GEChangeUsernameRequest"
}

func (event LEChangeUsernameRequest) Handle(l *Lobby, initiator *User) error {
	// Check if the game has already started
	if l.State != LSWaitingForPlayers {
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		initiator.WriteTemplate(tmpl, ErrorAlert, "Game already started")
		return errors.New("Game already started")
	}

	// Send the choose username screen to the player
	viewData := ViewData{
		Lobby:  l,
		Player: *initiator,
		IsHost: false, // host can't change username
	}
	tmpl := template.Must(template.ParseFiles(LobbyTemplate))
	if err := initiator.WriteTemplate(tmpl, ChooseUsernameView, viewData); err != nil {
		return err
	}

	return nil
}

type LEGameStartRequest struct{}

func (e LEGameStartRequest) String() string {
	return "LEGameStartRequest"
}

func (event LEGameStartRequest) Handle(l *Lobby, initiator *User) error {
	// Check if the initiator is the host
	if l.Host.ClientID != initiator.ClientID {
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		initiator.WriteTemplate(tmpl, ErrorAlert, "Only the host can start the game")
		return errors.New("Non-host tried to start the game")
	}

	// Check if there are enough players
	if len(l.Players) == 0 {
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		initiator.WriteTemplate(tmpl, ErrorAlert, "Not enough players")
		return errors.New("Can't start the game: not enough players")
	}

	// Check if the game has already started
	if l.State != LSWaitingForPlayers {
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		viewData := ViewData{
			Lobby:  l,
			Player: *initiator,
			IsHost: l.Host.ClientID == initiator.ClientID,
		}
		if err := initiator.WriteTemplate(tmpl, l.State.ViewName(), viewData); err != nil {
			return err
		}
		return errors.New("Game already started, sending current state to initiator")
	}

	// TODO: CHeck if the quiz is choosen

	// TODO: Check if the quiz has at least one question

	// Start game: go to the first question
	if err := l.StartNextQuestion(); err != nil {
		return err
	}

	return nil
}
