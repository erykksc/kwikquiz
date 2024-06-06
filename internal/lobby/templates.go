package lobby

import (
	"fmt"
	"html/template"
)

const BaseTemplatePath = "templates/base.html"

func tmplParseWithBase(path string) *template.Template {
	return template.Must(template.ParseFiles(path, BaseTemplatePath))
}

// Templates are used to render the different pages of the app
var NotFoundTmpl *template.Template
var IndexTmpl *template.Template
var LobbyCreateTmpl *template.Template
var LobbyErrorAlertTmpl *template.Template
var LobbiesTmpl *template.Template
var LobbyTmpl *template.Template

func init() {
	NotFoundTmpl = tmplParseWithBase("static/notfound.html")
	IndexTmpl = tmplParseWithBase("templates/index.html")
	LobbyCreateTmpl = tmplParseWithBase("templates/lobby/lobby-create.html")
	LobbiesTmpl = tmplParseWithBase("templates/lobby/lobbies.html")
	LobbyTmpl = tmplParseWithBase("templates/lobby/lobby.html")
	LobbyErrorAlertTmpl = LobbyTmpl.Lookup("error-alert")
}

type ViewData struct {
	Lobby *Lobby
	User  *User
}

// Views are the templates that are rendered for the different states of the lobby
// All of them require ViewData as the data to render
var ChooseUsernameView *template.Template
var WaitingRoomView *template.Template
var QuestionView *template.Template

type AnswerViewData struct {
	ViewData
}

var AnswerView *template.Template
var FinalResultsView *template.Template

func init() {
	ChooseUsernameView = tmplParseWithBase("templates/views/choose-username-view.html")
	WaitingRoomView = tmplParseWithBase("templates/views/waiting-room-view.html")
	QuestionView = tmplParseWithBase("templates/views/question-view.html")
	AnswerView = tmplParseWithBase("templates/views/answer-view.html")
	FinalResultsView = tmplParseWithBase("templates/views/final-results-view.html")
}

type LobbyState int

const (
	LSWaitingForPlayers LobbyState = iota
	LSQuestion
	LSAnswer
	LSFinalResults
)

// View returns the view template for the given state
func (state LobbyState) View() *template.Template {
	switch state {
	case LSWaitingForPlayers:
		return WaitingRoomView
	case LSQuestion:
		return QuestionView
	case LSAnswer:
		return AnswerView
	case LSFinalResults:
		return FinalResultsView
	default:
		panic("Undefined ViewName for state:" + fmt.Sprint(state))
	}
}
