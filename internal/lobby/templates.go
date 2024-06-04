package lobby

import (
	"fmt"
	"html/template"
)

const BaseTemplatePath = "templates/base.html"

// Templates are used to render the different pages of the app
var NotFoundTmpl *template.Template
var IndexTmpl *template.Template
var LobbyCreateTmpl *template.Template
var LobbyErrorAlertTmpl *template.Template
var LobbiesTmpl *template.Template
var LobbyTmpl *template.Template

func init() {
	NotFoundTmpl = template.Must(template.ParseFiles("static/notfound.html", BaseTemplatePath))
	IndexTmpl = template.Must(template.ParseFiles("templates/index.html", BaseTemplatePath))
	LobbyCreateTmpl = template.Must(template.ParseFiles("templates/lobby/lobby-create.html", BaseTemplatePath))
	LobbiesTmpl = template.Must(template.ParseFiles("templates/lobby/lobbies.html", BaseTemplatePath))
	LobbyTmpl = template.Must(template.ParseFiles("templates/lobby/lobby.html", BaseTemplatePath))
	LobbyErrorAlertTmpl = LobbyTmpl.Lookup("error-alert")
}

// Views are the templates that are rendered for the different states of the lobby
var ChooseUsernameView *template.Template
var WaitingRoomView *template.Template
var QuestionView *template.Template
var AnswerView *template.Template
var FinalResultsView *template.Template

func init() {
	ChooseUsernameView = template.Must(template.ParseFiles("templates/views/choose-username-view.html", BaseTemplatePath))
	WaitingRoomView = template.Must(template.ParseFiles("templates/views/waiting-room-view.html", BaseTemplatePath))
	QuestionView = template.Must(template.ParseFiles("templates/views/question-view.html", BaseTemplatePath))
	AnswerView = template.Must(template.ParseFiles("templates/views/answer-view.html", BaseTemplatePath))
	FinalResultsView = template.Must(template.ParseFiles("templates/views/final-results-view.html", BaseTemplatePath))
}

type ViewData struct {
	Lobby *Lobby
	User  *User
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
func (state LobbyState) ViewName() *template.Template {
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
