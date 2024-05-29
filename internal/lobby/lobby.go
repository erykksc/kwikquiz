package lobby

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"html/template"
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/gorilla/websocket"
)

type Lobby struct {
	mu                     sync.Mutex
	Host                   User
	Pin                    string
	TimePerQuestion        time.Duration // time per question
	Game                   common.Game
	CreatedAt              time.Time
	CurrentQuestionTimeout time.Time // timestamp (when the server should not accept answers anymore for the current question, the host can send a request to shorten the answer time)
	questionTimer          *CancellableTimer
	Players                map[ClientID]*User
	State                  LobbyState
	CurrentQuestion        int
	tmpl                   *template.Template // Store lobby template for quick access

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

type User struct {
	Conn                   *websocket.Conn
	ClientID               ClientID
	Username               string
	CurrentQuestionAnswers []int // Indices of the answers to the current question
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
		Players:         make(map[ClientID]*User),
		State:           LSWaitingForPlayers,
		CurrentQuestion: -1,
		tmpl:            template.Must(template.ParseFiles(LobbyTemplate)),
	}
}

// WriteView writes the view to the connection
// It isn't safe to call this function concurrently on lobby
func (l *Lobby) WriteView(conn *websocket.Conn, viewName ViewName, player User) error {
	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	type FrameData struct {
		Lobby  *Lobby
		Player User
		IsHost bool
	}

	data := FrameData{
		Lobby:  l,
		Player: player,
		IsHost: player.ClientID == l.Host.ClientID,
	}

	if err := l.tmpl.ExecuteTemplate(w, string(viewName), data); err != nil {
		return err
	}

	return nil
}

func (l *Lobby) StartNextQuestion() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.CurrentQuestion++
	l.State = LSQuestion

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

	l.State = LSAnswer

	// TODO: Send the answer screen to host

	// TODO: Send the answer screen to all players
	return nil
}
