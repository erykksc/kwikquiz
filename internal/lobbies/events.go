package lobbies

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/gorilla/websocket"
)

type lobbyEvent interface {
	String() string
	Handle(lobby *Lobby, initiator *User) error // Handles the event, is executed with the lobby's mutex locked
}

// parseLobbyEvent parses a [lobbyEvent] from a JSON in a byte slice
func parseLobbyEvent(jsonData []byte) (lobbyEvent, error) {
	type WsRequest struct {
		HEADERS common.HX_Headers
	}

	var wsRequest WsRequest
	if err := json.Unmarshal(jsonData, &wsRequest); err != nil {
		return nil, err
	}

	switch wsRequest.HEADERS.HxTriggerName {
	case "answer":
		var event leAnswerSubmitted
		// Parse id from "HxTrigger" in format "answer-<question-id>-<answer-id>"
		_, err := fmt.Sscanf(wsRequest.HEADERS.HxTrigger, "answer-q%d-a%d", &event.QuestionIdx, &event.AnswerIdx)
		if err != nil {
			return nil, err
		}
		return event, nil
	case "skip-to-answer-btn":
		var event leSkipToAnswerRequested
		return event, nil
	case "next-question-btn":
		var event leNextQuestionRequested
		return event, nil
	case "change-username-btn":
		var event leUsernameChangeRequested
		return event, nil
	case "start-game-btn":
		var event leGameStartRequested
		return event, nil
	case "new-username-form":
		var event leNewUsernameSubmitted
		if err := json.Unmarshal(jsonData, &event); err != nil {
			return nil, err
		}
		return event, nil
	default:
		return nil, errors.New("unrecognized trigger name, cannot parse event: " + wsRequest.HEADERS.HxTriggerName)
	}
}

// handleNewWebsocketConn handles a new websocket connection to the lobby
// This function bridges routes and events
func handleNewWebsocketConn(l *Lobby, conn *websocket.Conn, clientID clientID) (*User, error) {
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

	vData := ViewData{
		Lobby: l,
		User:  connectedUser,
	}
	if err := connectedUser.writeTemplate(view, vData); err != nil {
		return nil, err
	}

	return connectedUser, nil
}

// leNewUsernameSubmitted is an event that is triggered when a user submits a new username
type leNewUsernameSubmitted struct {
	Username string
}

func (e leNewUsernameSubmitted) String() string {
	return "GEUserSubmittedUsername: " + e.Username
}

func (event leNewUsernameSubmitted) Handle(l *Lobby, initiator *User) error {
	// Check if the username is empty
	if event.Username == "" {
		initiator.writeTemplate(LobbyErrorAlertTmpl, "Username cannot be empty")
		return errors.New("new username is empty")
	}

	// Check if game hasn't started yet
	if l.State != LsWaitingForPlayers {
		initiator.writeTemplate(LobbyErrorAlertTmpl, "Game already started")
		return errors.New("game already started")
	}

	// Check if new username isn't the same as the old one
	if initiator.Username == event.Username {
		slog.Info("Username is the same as the old one", "Username", event.Username)
		vData := ViewData{
			Lobby: l,
			User:  initiator,
		}
		if err := initiator.writeTemplate(WaitingRoomView, vData); err != nil {
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
		initiator.writeTemplate(LobbyErrorAlertTmpl, "Username already in the lobby")
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

	// Send the lobby screen to all players
	vData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(WaitingRoomView, vData); err != nil {
		slog.Error("Error writing view to host", "error", err, "host", l.Host)
	}

	for _, player := range l.Players {
		vData.User = player
		if err := player.writeTemplate(WaitingRoomView, vData); err != nil {
			slog.Error("Error writing view to user", "error", err, "user", player)
		}
	}
	return nil
}

// leUsernameChangeRequested is an event that is triggered when a user requests to change his username
type leUsernameChangeRequested struct{}

func (e leUsernameChangeRequested) String() string {
	return "GEChangeUsernameRequest"
}

func (event leUsernameChangeRequested) Handle(l *Lobby, initiator *User) error {
	// Check if the game has already started
	if l.State != LsWaitingForPlayers {
		initiator.writeTemplate(LobbyErrorAlertTmpl, "Game already started")
		return errors.New("Game already started")
	}

	// Send the choose username screen to the player
	vData := ViewData{
		Lobby: l,
		User:  initiator,
	}
	if err := initiator.writeTemplate(ChooseUsernameView, vData); err != nil {
		return err
	}

	return nil
}

// leGameStartRequested is an event that is triggered when a user requests to start the game
type leGameStartRequested struct{}

func (e leGameStartRequested) String() string {
	return "LEGameStartRequest"
}

func (event leGameStartRequested) Handle(l *Lobby, initiator *User) error {
	// Check if the initiator is the host
	if l.Host.ClientID != initiator.ClientID {
		initiator.writeTemplate(LobbyErrorAlertTmpl, "Only the host can start the game")
		return errors.New("Non-host tried to start the game")
	}

	// Check if there are enough players
	if len(l.Players) == 0 {
		initiator.writeTemplate(LobbyErrorAlertTmpl, "Not enough players")
		return errors.New("Can't start the game: not enough players")
	}

	// Check if the game has already started
	if l.State != LsWaitingForPlayers {
		vData := ViewData{
			Lobby: l,
			User:  initiator,
		}
		if err := initiator.writeTemplate(l.State.View(), vData); err != nil {
			return err
		}
		return errors.New("Game already started, sending current state to initiator")
	}

	// Check if the quiz has at least one question
	if len(l.Quiz.Questions) == 0 {
		err := initiator.writeTemplate(LobbyErrorAlertTmpl, "Quiz has no questions")
		if err != nil {
			return err
		}
		return errors.New("Can't start the game: quiz has no questions")
	}

	// Start game: go to the first question
	if err := l.startGame(); err != nil {
		return err
	}

	return nil
}

// leSkipToAnswerRequested is an event that is triggered when a user requests to skip to the answer
type leSkipToAnswerRequested struct{}

func (e leSkipToAnswerRequested) String() string {
	return "LESkipToAnswerRequest"
}

func (event leSkipToAnswerRequested) Handle(l *Lobby, initiator *User) error {
	l.questionTimer.Cancel()
	return nil
}

// leNextQuestionRequested is an event that is triggered when a user requests to go to the next question
type leNextQuestionRequested struct{}

func (e leNextQuestionRequested) String() string {
	return "LENextQuestionRequest"
}

func (event leNextQuestionRequested) Handle(l *Lobby, initiator *User) error {
	if err := l.startNextQuestion(); err != nil {
		return err
	}
	return nil
}

// leAnswerSubmitted is an event that is triggered when a user submits an answer
type leAnswerSubmitted struct {
	QuestionIdx int // Index of the question in Quiz.Questions
	AnswerIdx   int // Index of the answer in CurrentQuestion.Answers
}

func (e leAnswerSubmitted) String() string {
	return "LEAnswerSubmitted: " + fmt.Sprint(e.AnswerIdx)
}

func (e leAnswerSubmitted) Handle(l *Lobby, initiator *User) error {
	// Check if the answer index is valid
	if e.AnswerIdx < 0 || e.AnswerIdx >= len(l.CurrentQuestion.Answers) {
		initiator.writeTemplate(LobbyErrorAlertTmpl, "Invalid answer index")
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
	if l.State != LsQuestion {
		initiator.writeTemplate(LobbyErrorAlertTmpl, "Submitted after question timeout")
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
	vData := ViewData{
		Lobby: l,
		User:  initiator,
	}
	if err := initiator.writeNamedTemplate(QuestionView, "answer-options", vData); err != nil {
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
		vData.User = player
		if err := player.writeNamedTemplate(QuestionView, "player-count", vData); err != nil {
			return err
		}
	}

	return nil
}
