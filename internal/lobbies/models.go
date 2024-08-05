package lobbies

import (
	"fmt"
	"html/template"
	"log/slog"
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/game"
	"github.com/erykksc/kwikquiz/internal/quiz"
)

// Lobby is a actively running game session
type Lobby struct {
	mu    sync.Mutex
	Pin   string
	Host  *User
	Users map[common.ClientID]*User
	game.Game
}

type lobbyOptions struct {
	game.GameSettings
	Pin string
}

func (l *Lobby) View() *template.Template {
	if !l.Game.HasStarted() {
		return WaitingRoomView
	}

	if l.InRound() {
		return QuestionView
	}

	if l.HasEnded() {
		return onFinishView
	}

	return AnswerView
}

// NewLobbyOptions returns a new LobbyOptions struct with default values
func NewLobbyOptions() lobbyOptions {
	return lobbyOptions{
		GameSettings: game.GameSettings{
			Quiz: quiz.ExampleQuizMath,
			RoundSettings: game.RoundSettings{
				ReadingTime: 5 * time.Second,
				AnswerTime:  30 * time.Second,
			},
		},
	}
}

func createLobby(options lobbyOptions) *Lobby {
	return &Lobby{
		Pin:   options.Pin, // If it's empty, it will be generated by repository
		Users: make(map[common.ClientID]*User),
		Game:  game.CreateGame(options.GameSettings),
	}
}

func (l *Lobby) sendViewToAll(tmpl *template.Template) {
	vData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(tmpl, vData); err != nil {
		slog.Error("Error sending view to host", "template", tmpl.Name(), "error", err)
	}
	for _, user := range l.Users {
		vData.User = user
		if err := user.writeTemplate(tmpl, vData); err != nil {
			slog.Error("Error sending view to user", "template", tmpl.Name(), "error", err)
		}
	}
}

func (l *Lobby) sendViewToUser(tmpl *template.Template, user *User) error {
	vData := ViewData{
		Lobby: l,
		User:  user,
	}

	if err := user.writeTemplate(tmpl, vData); err != nil {
		return fmt.Errorf("sending view to user: %w", err)
	}
	return nil
}
