package lobbies

import (
	"fmt"
	"github.com/erykksc/kwikquiz/internal/models"
	"html/template"
	"path/filepath"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
)

func tmplParseWithBase(path string) *template.Template {
	return template.Must(template.ParseFiles(path, common.BaseTmplPath))
}

var LobbiesTmpl = tmplParseWithBase("templates/lobbies/lobbies.html")
var LobbyTmpl = tmplParseWithBase("templates/lobbies/lobby.html")
var LobbyErrorAlertTmpl = LobbyTmpl.Lookup("error-alert")

type ViewData struct {
	Lobby *Lobby
	User  *User
}

func ParseViewWithFuncs(path string) *template.Template {
	// get base name of the path
	baseName := filepath.Base(path)
	viewTmpl := template.Must(template.New(baseName).Funcs(template.FuncMap{
		"formatAsISO": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		// Decrement function used for checking if the current question is the last one
		"decrement": func(i int) int {
			return i - 1
		},
	}).ParseFiles(path, common.BaseTmplPath))

	return viewTmpl
}

// Views are the templates that are rendered for the different states of the lobby
// All of them require ViewData as the data to render
var ChooseUsernameView = ParseViewWithFuncs("templates/views/choose-username-view.html")
var WaitingRoomView = ParseViewWithFuncs("templates/views/waiting-room-view.html")
var QuestionView = ParseViewWithFuncs("templates/views/question-view.html")
var AnswerView = ParseViewWithFuncs("templates/views/answer-view.html")
var onFinishView = tmplParseWithBase("templates/views/on-finish-view.html")

type OnFinishData struct {
	PastGameID uint
	ViewData
}

// This template is used to render the lobby settings inside waitingRoomView
var LobbySettingsTmpl = WaitingRoomView.Lookup("lobby-settings")

type LobbySettingsData struct {
	Quizzes   []models.QuizMetadata
	Lobby     *Lobby
	LobbyJSON string
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
