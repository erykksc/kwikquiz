package lobby

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"html/template"
	"log/slog"
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/gorilla/websocket"
)

type Lobby struct {
	mu                     sync.Mutex
	Host                   *User
	Pin                    string
	TimePerQuestion        time.Duration // time per question
	Game                   common.Game
	CreatedAt              time.Time
	CurrentQuestionTimeout time.Time // timestamp (when the server should not accept answers anymore for the current question, the host can send a request to shorten the answer time)
	questionTimer          *CancellableTimer
	Players                map[ClientID]*User
	State                  LobbyState
	CurrentQuestion        int
}

// WriteTemplateToAll does tmpl.Execute(w, data) on websocket connections
// to the host and all players in the lobby
// slog.Error on every error while writing to websocket.Conn
func (l *Lobby) WriteTemplateToAll(tmpl *template.Template, data any) {
	if err := l.Host.WriteTemplate(tmpl, data); err != nil {
		slog.Error("Error sending view to host", "err", err, "data", data)
	}

	for _, player := range l.Players {
		if err := player.WriteTemplate(tmpl, data); err != nil {
			slog.Error("Error sending view to player", "err", err, "data", data)
		}
	}
}

// WriteTemplateToAll writes the template with the given name  and data
// to host and all players in the lobby
// slog.Error on every error while writing to websocket.Conn
func (l *Lobby) WriteNamedTemplateToAll(tmpl *template.Template, name string, data any) {
	if err := l.Host.WriteNamedTemplate(tmpl, name, data); err != nil {
		slog.Error("Error sending view to host", "err", err, "name", name, "data", data)
	}

	for _, player := range l.Players {
		if err := player.WriteNamedTemplate(tmpl, name, data); err != nil {
			slog.Error("Error sending view to player", "err", err, "name", name, "data", data)
		}
	}
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
	Conn     *websocket.Conn
	ClientID ClientID
	Username string
	IsHost   bool
}

// WriteTemplate does tmpl.Execute(w, data) on websocket connection to the user
func (client *User) WriteTemplate(tmpl *template.Template, data any) error {
	w, err := client.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	if err := tmpl.Execute(w, data); err != nil {
		return err
	}
	return nil
}

// WriteNamedTemplate does tmpl.ExecuteTemplate(w, name, data) on websocket connection to the user
func (client *User) WriteNamedTemplate(tmpl *template.Template, name string, data any) error {
	w, err := client.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		return err
	}
	return nil
}

type LobbyOptions struct {
	TimePerQuestion time.Duration
	Pin             string
}

func CreateLobby(options LobbyOptions) *Lobby {
	return &Lobby{
		Host:            nil,
		TimePerQuestion: options.TimePerQuestion,
		Pin:             options.Pin,
		Game:            common.Game{},
		CreatedAt:       time.Now(),
		Players:         make(map[ClientID]*User),
		State:           LSWaitingForPlayers,
		CurrentQuestion: -1,
	}
}

func (l *Lobby) StartNextQuestion() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.CurrentQuestion++
	l.State = LSQuestion

	// TODO:Check if the game is over (no more questions)

	// Send question view to all
	viewData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	l.WriteTemplateToAll(QuestionView, viewData)

	l.CurrentQuestionTimeout = time.Now().Add(l.TimePerQuestion)

	// Start the question timer
	l.questionTimer = NewCancellableTimer(l.TimePerQuestion)
	go func() {
		select {
		case <-l.questionTimer.timer.C:
			// Time completed
			slog.Debug("Question timeout reached")
			l.ShowAnswer()
			break
		case <-l.questionTimer.cancelChan:
			// Timer cancelled
			slog.Debug("Question timer cancelled")
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

	// Send answer view to all
	viewData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	l.WriteTemplateToAll(AnswerView, viewData)
	return nil
}
