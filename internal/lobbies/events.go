package lobbies

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/game"
	"github.com/erykksc/kwikquiz/internal/pastgames"
	"github.com/gorilla/websocket"
)

// Events are either user generated or system generated (for example when question timer expires)
// One event can cause another event
type lobbyEvent interface {
	String() string
	Handle(s Service, lobby *Lobby, initiator *User) error // Handles the event, is executed with the lobby's mutex locked
}

var lobbySystemUser = &User{
	Username: "SYSTEM",
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
	case "finish-game-btn":
		var event leEndGameRequested
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
func handleNewWebsocketConn(l *Lobby, conn *websocket.Conn, clientID common.ClientID) (*User, error) {
	connectedUser := &User{
		Conn:     conn,
		ClientID: clientID,
	}

	// view := l.State.View()
	view := l.View()

	player, connectedUserIsPlayer := l.Users[clientID]
	switch {
	// Check if it is the first user, if so, he becomes the host
	case l.Host == nil:
		slog.Info("New Host for the Lobby", "Lobby-Pin", l.Pin, "Client-ID", connectedUser.ClientID)
		connectedUser.Username = "HOST"
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
	Username game.Username
}

func (e leNewUsernameSubmitted) String() string {
	return "GEUserSubmittedUsername: " + string(e.Username)
}

func (event leNewUsernameSubmitted) Handle(_ Service, l *Lobby, initiator *User) error {
	if initiator.Username == "" {
		err := l.AddPlayer(event.Username)
		if err != nil {
			return err
		}
		initiator.Username = event.Username
		l.Users[initiator.ClientID] = initiator

	} else {
		err := l.ChangeUsername(initiator.Username, event.Username)
		if err != nil {
			return err
		}
		l.Users[initiator.ClientID].Username = event.Username
	}

	// TODO:Send updated player list to all
	vData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(WaitingRoomView, vData); err != nil {
		slog.Error("Error writing view to host", "error", err, "host", l.Host)
	}

	for _, player := range l.Users {
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

func (event leUsernameChangeRequested) Handle(_ Service, l *Lobby, initiator *User) error {
	// Check if the game has already started
	if l.HasStarted() {
		_ = initiator.writeTemplate(LobbyErrorAlertTmpl, "Game already started")
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

func (event leGameStartRequested) Handle(s Service, l *Lobby, initiator *User) error {
	// Check if the initiator is the host
	if l.Host.ClientID != initiator.ClientID {
		_ = initiator.writeTemplate(LobbyErrorAlertTmpl, "Only the host can start the game")
		return errors.New("Non-host tried to start the game")
	}

	err := l.Start()
	if err != nil {
		return err
	}

	roundFinished, err := l.RoundFinished()
	if err != nil {
		return err
	}

	go func() {
		<-roundFinished
		slog.Debug("Round finished, requesting to show answer")

		l.mu.Lock()
		err := leShowAnswerRequested{}.Handle(s, l, lobbySystemUser)
		if err != nil {
			slog.Error("Error handling ShowAnswerRequested", "error", err)
		}
		l.mu.Unlock()
	}()

	// Send question screen to players
	vData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(QuestionView, vData); err != nil {
		return err
	}
	for _, user := range l.Users {
		vData.User = user
		err := user.writeTemplate(QuestionView, vData)
		if err != nil {
			slog.Error("Error sending QuestionView to user", "error", err)
		}
	}
	return err
}

// leSkipToAnswerRequested is an event that is triggered when a user requests to skip to the answer
type leSkipToAnswerRequested struct{}

func (e leSkipToAnswerRequested) String() string {
	return "LESkipToAnswerRequest"
}

func (event leSkipToAnswerRequested) Handle(_ Service, l *Lobby, _ *User) error {
	err := l.FinishRoundEarly()
	if err != nil {
		return err
	}

	vData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(AnswerView, vData); err != nil {
		slog.Error("Error sending AnswerView to host", "error", err)
	}
	for _, user := range l.Users {
		vData.User = user
		err := user.writeTemplate(AnswerView, vData)
		if err != nil {
			slog.Error("Error sending AnswerView to user", "error", err, "client-id", user.ClientID)
		}
	}
	return nil
}

// leNextQuestionRequested is an event that is triggered when a user requests to go to the next question
type leNextQuestionRequested struct{}

func (e leNextQuestionRequested) String() string {
	return "LENextQuestionRequest"
}

func (event leNextQuestionRequested) Handle(s Service, l *Lobby, initiator *User) error {
	err := l.StartNextRound()
	if err != nil {
		return err
	}

	roundFinished, err := l.RoundFinished()
	if err != nil {
		return err
	}

	go func() {
		<-roundFinished
		slog.Debug("Round finished, requesting to show answer")

		l.mu.Lock()
		err := leShowAnswerRequested{}.Handle(s, l, lobbySystemUser)
		if err != nil {
			slog.Error("Error handling ShowAnswerRequested", "error", err)
		}
		l.mu.Unlock()
	}()

	// Send question view to all
	vData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(QuestionView, vData); err != nil {
		slog.Error("Error sending QuestionView to host", "error", err)
	}
	for _, player := range l.Users {
		vData.User = player
		err := player.writeTemplate(QuestionView, vData)
		if err != nil {
			slog.Error("Error sending QuestionView to user", "error", err)
		}
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

func (e leAnswerSubmitted) Handle(_ Service, l *Lobby, initiator *User) error {
	// Check if the question index is the current one
	if e.QuestionIdx != l.RoundNum() {
		return errors.New("Answer submitted for wrong question")
	}

	// Check if the initiator is the host
	if initiator.ClientID == l.Host.ClientID {
		return errors.New("Host tried to submit an answer")
	}

	err := l.SubmitAnswer(initiator.Username, e.AnswerIdx)
	if err != nil {
		return err
	}

	// Write updated view to the initiator
	vData := ViewData{
		Lobby: l,
		User:  initiator,
	}
	if err := initiator.writeNamedTemplate(QuestionView, "answer-options", vData); err != nil {
		return err
	}

	// Send template for how many people are left to answer
	for _, player := range l.Users {
		vData.User = player
		if err := player.writeNamedTemplate(QuestionView, "player-count", vData); err != nil {
			return err
		}
	}

	return nil
}

type leShowAnswerRequested struct{}

func (e leShowAnswerRequested) String() string {
	return "LEShowAnswerRequested"
}

func (e leShowAnswerRequested) Handle(_ Service, l *Lobby, _ *User) error {
	err := l.FinishRoundEarly()
	if err != nil && !errors.Is(err, game.ErrRoundAlreadyEnded) {
		return err
	}

	// Send answer view to all
	vData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(AnswerView, vData); err != nil {
		slog.Error("Error sending AnswerView to host", "error", err)
	}

	for _, player := range l.Users {
		vData.User = player
		if err := player.writeTemplate(AnswerView, vData); err != nil {
			slog.Error("Error sending AnswerView to user", "error", err)
		}
	}
	return nil
}

type leEndGameRequested struct{}

func (e leEndGameRequested) String() string {
	return "LEEndGameRequested"
}

func (e leEndGameRequested) Handle(s Service, l *Lobby, _ *User) error {
	err := l.Finish()
	if err != nil {
		return err
	}

	if err := s.lRepo.DeleteLobby(l.Pin); err != nil {
		return err
	}

	scores := make([]pastgames.PlayerScore, 0, len(l.Users))
	for _, player := range l.Leaderboard() {
		scores = append(scores, pastgames.PlayerScore{
			Username: string(player.Username),
			Score:    player.Points,
		})
	}

	pastGame := pastgames.PastGame{
		StartedAt: l.StartedAt(),
		EndedAt:   l.EndedAt(),
		QuizTitle: l.Quiz().Title(),
		Scores:    scores,
	}
	id, err := s.pgRepo.Insert(&pastGame)
	if err != nil {
		return err
	}

	data := OnFinishData{
		PastGameID: id,
		ViewData: ViewData{
			Lobby: l,
			User:  l.Host,
		},
	}

	// Close websocket connections
	// The redirection to lobby results is handled by the view
	if l.Host.Conn != nil {
		_ = l.Host.writeTemplate(onFinishView, data)
		l.Host.Conn.Close()
	}
	for _, player := range l.Users {
		if player.Conn == nil {
			slog.Error("Player connection is nil", "player", player)
			continue
		}
		data.User = player
		err = player.writeTemplate(onFinishView, data)
		if err != nil {
			slog.Error("Error sending OnFinishView to user", "error", err)
		}
		player.Conn.Close()
	}
	return nil
}
