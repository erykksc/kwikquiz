package lobby

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/gorilla/websocket"
)

type Lobby struct {
	mu                     sync.Mutex
	Host                   ClientID
	Pin                    string
	TimePerQuestion        time.Duration // time per question
	Game                   common.Game
	CreatedAt              time.Time
	CurrentQuestionTimeout time.Time // timestamp (when the server should not accept answers anymore for the current question, the host can send a request to shorten the answer time)
	questionTimer          *CancellableTimer
	Players                map[ClientID]*Player
	State                  LobbyState
	CurrentQuestion        int
}

type ClientID string

func NewClientID() (ClientID, error) {
	// Generate 8 bytes from the timestamp (64 bits)
	timestamp := time.Now().Unix()
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))

	// Generate 8 random bytes (64 bits)
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Combine the two byte slices into 128 bits
	combinedBytes := append(timestampBytes, randomBytes...)

	// Encode the 128 bits into a base64 string
	encoded := base64.StdEncoding.EncodeToString(combinedBytes)
	return ClientID(encoded), nil
}

type Player struct {
	Conn     *websocket.Conn
	Username string
}

type LobbyOptions struct {
	TimePerQuestion time.Duration
	Pin             string
}

type LobbyState int

const (
	WaitingForPlayers = iota
	Question
	Answer
	FinalResults
)

func CreateLobby(options LobbyOptions) *Lobby {
	return &Lobby{
		TimePerQuestion: options.TimePerQuestion,
		Pin:             options.Pin,
		Game:            common.Game{},
		CreatedAt:       time.Now(),
		Players:         make(map[ClientID]*Player),
		State:           WaitingForPlayers,
		CurrentQuestion: -1,
	}
}

func (l *Lobby) StartNextQuestion() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.CurrentQuestion++
	l.State = Question

	// TODO:Check if the game is over (no more questions)

	// TODO: Send the question screen to host

	// TODO: Send the question screen to all players

	l.CurrentQuestionTimeout = time.Now().Add(l.TimePerQuestion)

	// Start the question timer
	l.questionTimer = NewCancellableTimer(l.TimePerQuestion)
	go func() {
		select {
		case <-l.questionTimer.timer.C:
			// Time completed
			l.ShowAnswer()
			break
		case <-l.questionTimer.cancelChan:
			// Timer cancelled
			l.ShowAnswer()
			break
		}
	}()
	return nil
}

func (l *Lobby) ShowAnswer() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.State = Answer

	// TODO: Send the answer screen to host

	// TODO: Send the answer screen to all players
	return nil
}
