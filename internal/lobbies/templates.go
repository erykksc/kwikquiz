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

var lobbyErrorAlertTmpl *template.Template
var lobbiesTmpl *template.Template
var lobbyTmpl *template.Template

func init() {
	lobbiesTmpl = tmplParseWithBase("templates/lobbies/lobbies.html")
	lobbyTmpl = tmplParseWithBase("templates/lobbies/lobby.html")
	lobbyErrorAlertTmpl = lobbyTmpl.Lookup("error-alert")
}

type viewData struct {
	Lobby *lobby
	User  *user
}

// Views are the templates that are rendered for the different states of the lobby
// All of them require ViewData as the data to render
var chooseUsernameView *template.Template
var waitingRoomView *template.Template
var questionView *template.Template
var answerView *template.Template
var onFinishView *template.Template

type OnFinishData struct {
	PastGameID int
	viewData
}

// This template is used to render the lobby settings inside waitingRoomView
var lobbySettingsTmpl *template.Template

type lobbySettingsData struct {
	Quizzes []quiz.QuizMetadata
	Lobby   *lobby
}

func init() {
	chooseUsernameView = tmplParseWithBase("templates/views/choose-username-view.html")
	waitingRoomView = tmplParseWithBase("templates/views/waiting-room-view.html")
	questionView = tmplParseWithBase("templates/views/question-view.html")
	// Decrement function used for checking if the current question is the last one
	answerView = template.Must(template.New("answer-view.html").Funcs(template.FuncMap{
		"decrement": func(i int) int {
			return i - 1
		},
	}).ParseFiles("templates/views/answer-view.html", common.BaseTmplPath))
	onFinishView = tmplParseWithBase("templates/views/on-finish-view.html")
	lobbySettingsTmpl = waitingRoomView.Lookup("lobby-settings")
}

type lobbyState int

const (
	lsWaitingForPlayers lobbyState = iota
	lsQuestion
	lsAnswer
)

// View returns the view template for the given state
func (state lobbyState) View() *template.Template {
	switch state {
	case lsWaitingForPlayers:
		return waitingRoomView
	case lsQuestion:
		return questionView
	case lsAnswer:
		return answerView
	default:
		panic("Undefined ViewName for state:" + fmt.Sprint(state))
	}
}
