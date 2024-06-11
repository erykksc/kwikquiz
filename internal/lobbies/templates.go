package lobbies

import (
	"fmt"
	"html/template"

	"github.com/erykksc/kwikquiz/internal/common"
)

func tmplParseWithBase(path string) *template.Template {
	return template.Must(template.ParseFiles(path, common.BaseTmplPath))
}

// Templates are used to render the different pages of the app
var lobbyCreateTmpl *template.Template

// The form used for creating the lobby/session
var createLobbyFormTmpl *template.Template

type createLobbyFormData struct {
	Pin       string
	Username  string
	FormError string
}

var lobbyErrorAlertTmpl *template.Template
var lobbiesTmpl *template.Template
var lobbyTmpl *template.Template

func init() {
	lobbyCreateTmpl = tmplParseWithBase("templates/lobbies/lobby-create.html")
	createLobbyFormTmpl = lobbyCreateTmpl.Lookup("create-lobby-form")
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
var finalResultsView *template.Template

func init() {
	chooseUsernameView = tmplParseWithBase("templates/views/choose-username-view.html")
	waitingRoomView = tmplParseWithBase("templates/views/waiting-room-view.html")
	questionView = tmplParseWithBase("templates/views/question-view.html")
	answerView = tmplParseWithBase("templates/views/answer-view.html")
	finalResultsView = tmplParseWithBase("templates/views/final-results-view.html")
}

type lobbyState int

const (
	lsWaitingForPlayers lobbyState = iota
	lsQuestion
	lsAnswer
	lsFinalResults
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
	case lsFinalResults:
		return finalResultsView
	default:
		panic("Undefined ViewName for state:" + fmt.Sprint(state))
	}
}
