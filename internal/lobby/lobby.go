package lobby

import (
	"strconv"
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/gorilla/websocket"
)

type Lobby struct {
	mu              sync.Mutex
	Host            ClientID
	Pin             string
	TimePerQuestion time.Duration // time per question
	common.Game
	CreatedAt              time.Time
	CurrentQuestion        common.Question
	CurrentQuestionTimeout time.Time // timestamp (when the server should not accept answers anymore for the current question, the host can send a request to shorten the answer time)
	Players                map[ClientID]*Player
}

type ClientID string

func NewClientID() (ClientID, error) {
	randomStr, err := common.RandomHexString(200)
	if err != nil {
		return "", err
	}
	clientID := strconv.FormatInt(time.Now().Unix(), 16) + randomStr
	return ClientID(clientID), nil
}

type Player struct {
	Conn     *websocket.Conn
	Username string
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
		Players:         make(map[ClientID]*Player),
	}
}
