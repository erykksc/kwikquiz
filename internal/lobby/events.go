package lobby

import (
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"

	"github.com/gorilla/websocket"
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

func HandleNewWebsocketConn(l *Lobby, conn *websocket.Conn, clientID ClientID) (*User, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Assume that connectedUser at this point only has the Conn and ClientID fields
	// This function should set any other fields that are needed
	connectedUser := &User{
		Conn:     conn,
		ClientID: clientID,
	}

	viewName := l.State.ViewName()

	player, connectedUserIsPlayer := l.Players[clientID]
	switch {
	// Check if it is the first user, if so, he becomes the host
	case l.Host == nil:
		slog.Info("New Host for the Lobby", "Lobby-Pin", l.Pin, "Client-ID", connectedUser.ClientID)
		l.Host = connectedUser
		connectedUser.IsHost = true

	// Check if host is trying to reconnect
	case l.Host.ClientID == connectedUser.ClientID:
		slog.Info("Host reconnecting", "Lobby-Pin", l.Pin, "Client-ID", connectedUser.ClientID)
		// Update the connection
		l.Host.Conn = conn
		connectedUser = l.Host

	// Check if player is trying to reconnect
	case connectedUserIsPlayer:
		slog.Info("Player reconnecting", "Lobby-Pin", l.Pin, "Client-ID", connectedUser.ClientID)
		// Update the connection
		player.Conn = conn
		connectedUser = player

	// New User connecting
	default:
		slog.Info("New Player for Lobby", "Lobby-Pin", l.Pin, "Client-ID", connectedUser.ClientID)
		viewName = ChooseUsernameView
	}

	viewData := ViewData{
		Lobby: l,
		User:  connectedUser,
	}
	tmpl := template.Must(template.ParseFiles(LobbyTemplate))
	if err := connectedUser.WriteTemplate(tmpl, viewName, viewData); err != nil {
		return nil, err
	}

	return connectedUser, nil
}

type LEUSubmittedUsername struct {
	Username string
}

func (e LEUSubmittedUsername) String() string {
	return "GEUserSubmittedUsername: " + e.Username
}

func (event LEUSubmittedUsername) Handle(l *Lobby, initiator *User) error {
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
	if initiator.Username == event.Username {
		slog.Info("Username is the same as the old one", "Username", event.Username)
		tmpl := template.Must(template.ParseFiles(LobbyTemplate))
		viewData := ViewData{
			Lobby: l,
			User:  initiator,
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

	// Check if the initiator is in the players list
	if _, ok := l.Players[initiator.ClientID]; !ok {
		// If not, add him to the players list
		slog.Info("Adding new player to the lobby", "Client-ID", initiator.ClientID)
		initiator.Username = event.Username
		l.Players[initiator.ClientID] = initiator
	} else {
		// If he is, update his username
		slog.Info("Updating username", "old", initiator.Username, "new", event.Username)
		initiator.Username = event.Username
	}

	// l.Players[initiator.ClientID] = &User{
	// 	Conn:     initiator.Conn,
	// 	Username: event.Username,
	// }

	// Send the lobby screen to all players
	viewData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	tmpl := template.Must(template.ParseFiles(LobbyTemplate))
	if err := l.Host.WriteTemplate(tmpl, WaitingRoomView, viewData); err != nil {
		slog.Error("Error writing view to host", "error", err, "host", l.Host)
	}

	for _, player := range l.Players {
		viewData.User = player
		if err := player.WriteTemplate(tmpl, WaitingRoomView, viewData); err != nil {
			slog.Error("Error writing view to user", "error", err, "user", player)
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
		Lobby: l,
		User:  initiator,
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
			Lobby: l,
			User:  initiator,
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
