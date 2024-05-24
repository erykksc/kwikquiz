package lobby

import (
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/gorilla/websocket"
)

type Lobby struct {
	mu              sync.Mutex
	Host            websocket.Conn
	Pin             string
	TimePerQuestion time.Duration // time per question
	common.Game
	CreatedAt              time.Time
	CurrentQuestion        common.Question
	CurrentQuestionTimeout time.Time // timestamp (when the server should not accept answers anymore for the current question, the host can send a request to shorten the answer time)
	// Username -> Conn
	Players map[string]*websocket.Conn
}

type LobbyOptions struct {
	TimePerQuestion time.Duration
	Pin             string
}

func CreateLobby(options LobbyOptions) *Lobby {
	return &Lobby{
		TimePerQuestion: options.TimePerQuestion,
		Pin:             options.Pin,
		Game:            common.Game{},
		CreatedAt:       time.Now(),
		Players:         make(map[string]*websocket.Conn),
	}
}
