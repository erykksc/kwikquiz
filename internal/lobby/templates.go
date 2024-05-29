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

// ViewName is the name of the view that is displayed to the user inside the lobby
// Views are sent by websocket to the client
type ViewName string

const (
	ChooseUsernameViewName ViewName = "choose-username-view"
	WaitingRoomViewName    ViewName = "waiting-room-view"
	QuestionViewName       ViewName = "question-view"
	AnswerViewName         ViewName = "answer-view"
	FinalResultsViewName   ViewName = "final-results-view"
)

type LobbyState int

const (
	LSWaitingForPlayers LobbyState = iota
	LSQuestion
	LSAnswer
	LSFinalResults
)

// ViewName returns the mapping of the LobbyState to the ViewName
// that displays the state
func (state LobbyState) ViewName() ViewName {
	m := make(map[LobbyState]ViewName)
	m[LSWaitingForPlayers] = WaitingRoomViewName
	m[LSQuestion] = QuestionViewName
	m[LSAnswer] = AnswerViewName
	m[LSFinalResults] = FinalResultsViewName
	fName, ok := m[state]
	if !ok {
		panic("Undefined ViewName for state:" + fmt.Sprint(state))
	}
	return fName
}
