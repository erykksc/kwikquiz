package lobbies

import (
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/quiz"
)

func GetExamples() []*Lobby {
	return []*Lobby{
		&Example1234Lobby,
		&ExampleLobbyOnQuestionView,
	}
}

var Example1234Lobby = Lobby{
	Pin:                      "1234",
	mu:                       sync.Mutex{},
	CreatedAt:                time.Now(),
	StartedAt:                time.Time{},
	FinishedAt:               time.Time{},
	Quiz:                     quiz.ExampleQuizGeography,
	Host:                     &User{},
	TimePerQuestion:          30 * time.Second,
	TimeForReading:           time.Second * 5,
	Players:                  make(map[ClientID]*User),
	State:                    LsWaitingForPlayers,
	questionTimer:            &cancellableTimer{},
	CurrentQuestionStartTime: time.Time{},
	CurrentQuestionTimeout:   time.Time{},
	ReadingTimeout:           time.Time{},
	CurrentQuestionIdx:       0,
	CurrentQuestion:          &quiz.Question{},
	PlayersAnswering:         0,
	Leaderboard:              []*User{},
}
var ExampleLobbyOnQuestionView = Lobby{
	Pin:                      "0001",
	mu:                       sync.Mutex{},
	CreatedAt:                time.Now(),
	StartedAt:                time.Now(),
	FinishedAt:               time.Time{},
	Quiz:                     quiz.ExampleQuizGeography,
	Host:                     &User{},
	TimePerQuestion:          30 * time.Second,
	TimeForReading:           time.Second * 5,
	Players:                  make(map[ClientID]*User),
	State:                    LsQuestion,
	questionTimer:            &cancellableTimer{},
	CurrentQuestionStartTime: time.Now(),
	CurrentQuestionTimeout:   time.Now().Add(100 * time.Second),
	ReadingTimeout:           time.Now(),
	CurrentQuestionIdx:       0,
	CurrentQuestion:          &quiz.ExampleQuizGeography.Questions[0],
	PlayersAnswering:         3,
	Leaderboard:              []*User{},
}
