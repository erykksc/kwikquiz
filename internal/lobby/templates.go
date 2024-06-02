package lobby

import (
	"fmt"
)

// Templates
const (
	NotFoundPage        = "static/notfound.html"
	BaseTemplate        = "templates/base.html"
	IndexTemplate       = "templates/index.html"
	LobbiesTemplate     = "templates/lobby/lobbies.html"
	LobbyTemplate       = "templates/lobby/lobby.html"
	LobbyCreateTemplate = "templates/lobby/lobby-create.html"
)

// Views
const (
	ChooseUsernameView = "choose-username-view"
	WaitingRoomView    = "waiting-room-view"
	QuestionView       = "question-view"
	AnswerView         = "answer-view"
	FinalResultsView   = "final-results-view"
	ErrorAlert         = "error-alert"
)

type ViewData struct {
	Lobby  *Lobby
	Player User
	IsHost bool
}

type LobbyState int

const (
	LSWaitingForPlayers LobbyState = iota
	LSQuestion
	LSAnswer
	LSFinalResults
)

// ViewName returns the mapping of the LobbyState to the ViewName
// that displays the state
func (state LobbyState) ViewName() string {
	m := make(map[LobbyState]string)
	m[LSWaitingForPlayers] = WaitingRoomView
	m[LSQuestion] = QuestionView
	m[LSAnswer] = AnswerView
	m[LSFinalResults] = FinalResultsView
	fName, ok := m[state]
	if !ok {
		panic("Undefined ViewName for state:" + fmt.Sprint(state))
	}
	return fName
}
