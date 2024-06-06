package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

type LobbyEvent interface {
	String() string
	Handle(lobby *Lobby, initiator *User) error
}

type HXData struct {
	HxCurrentURL  string `json:"HX-Current-URL"`
	HxRequest     string `json:"HX-Request"`
	HxTarget      string `json:"HX-Target"`
	HxTrigger     string `json:"HX-Trigger"`
	HxTriggerName string `json:"HX-Trigger-Name"`
}

// ParseLobbyEvent parses a GameEvent from a JSON in a byte slice
func ParseLobbyEvent(data []byte) (LobbyEvent, error) {
	type Headers struct {
		HXData
		EventType string
	}

	type WsRequest struct {
		HEADERS Headers
	}

	var wsRequest WsRequest
	if err := json.Unmarshal(data, &wsRequest); err != nil {
		return nil, err
	}

	switch wsRequest.HEADERS.HxTriggerName {
	case "answer":
		var event LEAnswerSubmitted
		// Parse id from "HxTrigger" in format "answer-<question-id>-<answer-id>"
		_, err := fmt.Sscanf(wsRequest.HEADERS.HxTrigger, "answer-q%d-a%d", &event.QuestionIdx, &event.AnswerIdx)
		if err != nil {
			return nil, err
		}
		return event, nil
	case "skip-to-answer-btn":
		var event LESkipToAnswerRequest
		return event, nil
	case "next-question-btn":
		var event LENextQuestionRequest
		return event, nil
	case "change-username-btn":
		var event LEChangeUsernameRequest
		return event, nil
	case "start-game-btn":
		var event LEGameStartRequest
		return event, nil
	case "new-username-form":
		var event LEUSubmittedUsername
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return event, nil
	default:
		return nil, errors.New("unrecognized trigger name, cannot parse event: " + wsRequest.HEADERS.HxTriggerName)
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

	view := l.State.View()

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
		view = ChooseUsernameView
	}

	viewData := ViewData{
		Lobby: l,
		User:  connectedUser,
	}
	if err := connectedUser.WriteTemplate(view, viewData); err != nil {
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
		initiator.WriteTemplate(LobbyErrorAlertTmpl, "Username cannot be empty")
		return errors.New("new username is empty")
	}

	// Check if game hasn't started yet
	if l.State != LSWaitingForPlayers {
		initiator.WriteTemplate(LobbyErrorAlertTmpl, "Game already started")
		return errors.New("game already started")
	}

	// Check if new username isn't the same as the old one
	if initiator.Username == event.Username {
		slog.Info("Username is the same as the old one", "Username", event.Username)
		viewData := ViewData{
			Lobby: l,
			User:  initiator,
		}
		if err := initiator.WriteTemplate(WaitingRoomView, viewData); err != nil {
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
		initiator.WriteTemplate(LobbyErrorAlertTmpl, "Username already in the lobby")
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
	if err := l.Host.WriteTemplate(WaitingRoomView, viewData); err != nil {
		slog.Error("Error writing view to host", "error", err, "host", l.Host)
	}

	for _, player := range l.Players {
		viewData.User = player
		if err := player.WriteTemplate(WaitingRoomView, viewData); err != nil {
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
		initiator.WriteTemplate(LobbyErrorAlertTmpl, "Game already started")
		return errors.New("Game already started")
	}

	// Send the choose username screen to the player
	viewData := ViewData{
		Lobby: l,
		User:  initiator,
	}
	if err := initiator.WriteTemplate(ChooseUsernameView, viewData); err != nil {
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
		initiator.WriteTemplate(LobbyErrorAlertTmpl, "Only the host can start the game")
		return errors.New("Non-host tried to start the game")
	}

	// Check if there are enough players
	if len(l.Players) == 0 {
		initiator.WriteTemplate(LobbyErrorAlertTmpl, "Not enough players")
		return errors.New("Can't start the game: not enough players")
	}

	// Check if the game has already started
	if l.State != LSWaitingForPlayers {
		viewData := ViewData{
			Lobby: l,
			User:  initiator,
		}
		if err := initiator.WriteTemplate(l.State.View(), viewData); err != nil {
			return err
		}
		return errors.New("Game already started, sending current state to initiator")
	}

	// Check if the quiz has at least one question
	if len(l.Quiz.Questions) == 0 {
		err := initiator.WriteTemplate(LobbyErrorAlertTmpl, "Quiz has no questions")
		if err != nil {
			return err
		}
		return errors.New("Can't start the game: quiz has no questions")
	}

	// Start game: go to the first question
	if err := l.StartNextQuestion(); err != nil {
		return err
	}

	return nil
}

type LESkipToAnswerRequest struct{}

func (e LESkipToAnswerRequest) String() string {
	return "LESkipToAnswerRequest"
}

func (event LESkipToAnswerRequest) Handle(l *Lobby, initiator *User) error {
	l.questionTimer.Cancel()
	return nil
}

type LENextQuestionRequest struct{}

func (e LENextQuestionRequest) String() string {
	return "LENextQuestionRequest"
}

func (event LENextQuestionRequest) Handle(l *Lobby, initiator *User) error {
	if err := l.StartNextQuestion(); err != nil {
		return err
	}
	return nil
}

type LEAnswerSubmitted struct {
	QuestionIdx int // Index of the question in Quiz.Questions
	AnswerIdx   int // Index of the answer in CurrentQuestion.Answers
}

func (e LEAnswerSubmitted) String() string {
	return "LEAnswerSubmitted: " + fmt.Sprint(e.AnswerIdx)
}

func (e LEAnswerSubmitted) Handle(l *Lobby, initiator *User) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if the answer index is valid
	if e.AnswerIdx < 0 || e.AnswerIdx >= len(l.CurrentQuestion.Answers) {
		initiator.WriteTemplate(LobbyErrorAlertTmpl, "Invalid answer index")
		return errors.New("Invalid answer index")
	}

	// Check if the question index is the current one
	if e.QuestionIdx != l.CurrentQuestionIdx {
		slog.Warn("Answer submitted for wrong question", "Client-ID", initiator.ClientID)
		return nil
	}

	// Check if the initiator isn't the host
	if initiator.ClientID == l.Host.ClientID {
		slog.Warn("Host tried to submit an answer")
		return nil
	}

	// Check if the game is in the question state
	if l.State != LSQuestion {
		initiator.WriteTemplate(LobbyErrorAlertTmpl, "Submitted after question timeout")
		return errors.New("Submitted after question timeout")
	}

	// Check if the user has already submitted an answer
	if initiator.SubmittedAnswerIdx != -1 {
		slog.Warn("User tried to submit an answer twice", "Client-ID", initiator.ClientID)
		return nil
	}

	// Update the user's answer
	initiator.SubmittedAnswerIdx = e.AnswerIdx
	initiator.AnswerSubmissionTime = time.Now()

	// Write updated view to the initiator
	viewData := ViewData{
		Lobby: l,
		User:  initiator,
	}
	if err := initiator.WriteNamedTemplate(QuestionView, "answer-options", viewData); err != nil {
		return err
	}

	l.PlayersAnswering--

	// Check if all players have answered
	if l.PlayersAnswering == 0 {
		// End the question
		l.questionTimer.Cancel()
	}

	// Send template for how many people are left to answer
	for _, player := range l.Players {
		viewData.User = player
		if err := player.WriteNamedTemplate(QuestionView, "player-count", viewData); err != nil {
			return err
		}
	}

	return nil
}
