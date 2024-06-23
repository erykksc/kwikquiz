package lobbies

import (
	"fmt"
	"html/template"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/quiz"
)

func tmplParseWithBase(path string) *template.Template {
	return template.Must(template.ParseFiles(path, common.BaseTmplPath))
}

var LobbiesTmpl = tmplParseWithBase("templates/lobbies/lobbies.html")
var LobbyTmpl = tmplParseWithBase("templates/lobbies/lobby.html")
var LobbyErrorAlertTmpl = LobbyTmpl.Lookup("error-alert")

type ViewData struct {
	Lobby *lobby
	User  *user
}

// Views are the templates that are rendered for the different states of the lobby
// All of them require ViewData as the data to render
var ChooseUsernameView = tmplParseWithBase("templates/views/choose-username-view.html")
var WaitingRoomView = tmplParseWithBase("templates/views/waiting-room-view.html")
var QuestionView = tmplParseWithBase("templates/views/question-view.html")

// Decrement function used for checking if the current question is the last one
var AnswerView = template.Must(template.New("answer-view.html").Funcs(template.FuncMap{
	"decrement": func(i int) int {
		return i - 1
	},
}).ParseFiles("templates/views/answer-view.html", common.BaseTmplPath))
var onFinishView = tmplParseWithBase("templates/views/on-finish-view.html")

type OnFinishData struct {
	PastGameID int
	ViewData
}

// This template is used to render the lobby settings inside waitingRoomView
var LobbySettingsTmpl = WaitingRoomView.Lookup("lobby-settings")

type LobbySettingsData struct {
	Quizzes []quiz.QuizMetadata
	Lobby   *lobby
}

type LobbyState int

const (
	LsWaitingForPlayers LobbyState = iota
	LsQuestion
	LsAnswer
)

// View returns the view template for the given state
func (state LobbyState) View() *template.Template {
	switch state {
	case LsWaitingForPlayers:
		return WaitingRoomView
	case LsQuestion:
		return QuestionView
	case LsAnswer:
		return AnswerView
	default:
		panic("Undefined ViewName for state:" + fmt.Sprint(state))
	}
}
