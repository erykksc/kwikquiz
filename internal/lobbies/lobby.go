package lobbies

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

// lobby is a actively running game session
type lobby struct {
	common.Game
	mu                       sync.Mutex
	Host                     *user
	Pin                      string
	TimePerQuestion          time.Duration
	CreatedAt                time.Time
	Players                  map[clientID]*user
	State                    lobbyState
	questionTimer            *cancellableTimer
	CurrentQuestionStartTime time.Time
	CurrentQuestionTimeout   string // ISO 8601 String
	CurrentQuestionIdx       int
	CurrentQuestion          *common.Question
	PlayersAnswering         int // Number of players who haven't submitted an answer
}

type lobbyOptions struct {
	TimePerQuestion time.Duration
	Pin             string
}

func createLobby(options lobbyOptions) *lobby {
	return &lobby{
		Host:               nil,
		TimePerQuestion:    options.TimePerQuestion,
		Pin:                options.Pin,
		Game:               common.Game{},
		CreatedAt:          time.Now(),
		Players:            make(map[clientID]*user),
		State:              lsWaitingForPlayers,
		CurrentQuestionIdx: -1,
	}
}

type clientID string

func newClientID() (clientID, error) {
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
	return clientID(encoded), nil
}

type user struct {
	Conn                 *websocket.Conn
	ClientID             clientID
	Username             string
	IsHost               bool
	SubmittedAnswerIdx   int
	AnswerSubmissionTime time.Time
	Score                int64
	NewPoints            int64
}

// writeTemplate does tmpl.Execute(w, data) on websocket connection to the user
func (client *user) writeTemplate(tmpl *template.Template, data any) error {
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

// writeNamedTemplate does tmpl.ExecuteTemplate(w, name, data) on websocket connection to the user
func (client *user) writeNamedTemplate(tmpl *template.Template, name string, data any) error {
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

func (l *lobby) startNextQuestion() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Reset Player answers
	l.Host.SubmittedAnswerIdx = -1
	l.Host.AnswerSubmissionTime = time.Time{}
	for _, player := range l.Players {
		player.SubmittedAnswerIdx = -1
		player.AnswerSubmissionTime = time.Time{}
	}

	l.CurrentQuestionIdx++
	l.State = lsQuestion
	l.CurrentQuestionStartTime = time.Now()
	l.CurrentQuestionTimeout = l.CurrentQuestionStartTime.Add(l.TimePerQuestion).Format(time.RFC3339)
	l.PlayersAnswering = len(l.Players)

	// Check if the game has finished
	if l.CurrentQuestionIdx == len(l.Quiz.Questions) {
		l.State = lsFinalResults
		// Send final results view to all
		vData := viewData{
			Lobby: l,
			User:  l.Host,
		}
		if err := l.Host.writeTemplate(finalResultsView, vData); err != nil {
			slog.Error("Error sending FinalResultsView to host", "error", err)
		}
		for _, player := range l.Players {
			vData.User = player
			err := player.writeTemplate(finalResultsView, vData)
			if err != nil {
				slog.Error("Error sending FinalResultsView to user", "error", err)
			}
		}
		return nil
	}

	l.CurrentQuestion = &l.Quiz.Questions[l.CurrentQuestionIdx]

	// Start the question timer
	l.questionTimer = NewCancellableTimer(l.TimePerQuestion)
	go func() {
		select {
		case <-l.questionTimer.timer.C:
			// Time completed
			slog.Debug("Question timeout reached")
			l.showAnswer()
			break
		case <-l.questionTimer.cancelChan:
			// Timer cancelled
			slog.Debug("Question timer cancelled")
			l.showAnswer()
			break
		}
	}()

	// Send question view to all
	vData := viewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(questionView, vData); err != nil {
		slog.Error("Error sending QuestionView to host", "error", err)
	}
	for _, player := range l.Players {
		vData.User = player
		err := player.writeTemplate(questionView, vData)
		if err != nil {
			slog.Error("Error sending QuestionView to user", "error", err)
		}
	}
	return nil
}

func (l *lobby) showAnswer() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.State = lsAnswer

	// Add points to players
	for _, player := range l.Players {
		if player.SubmittedAnswerIdx == -1 {
			// Player didn't submit an answer
			player.NewPoints = 0
			continue
		}
		submittedAnswer := &l.CurrentQuestion.Answers[player.SubmittedAnswerIdx]
		// Give points based on time to answer
		if submittedAnswer.IsCorrect {
			timeToAnswer := player.AnswerSubmissionTime.Sub(l.CurrentQuestionStartTime)
			if timeToAnswer < time.Millisecond*500 {
				// Maximum points for answering in less than 500ms
				player.NewPoints = 1000
			} else {
				player.NewPoints = int64((1 - (float64(timeToAnswer) / float64(l.TimePerQuestion) / 2.0)) * 1000)
			}
			player.Score += player.NewPoints
		}
	}

	// Send answer view to all
	vData := viewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(answerView, vData); err != nil {
		slog.Error("Error sending AnswerView to host", "error", err)
	}

	for _, player := range l.Players {
		vData.User = player
		if err := player.writeTemplate(answerView, vData); err != nil {
			slog.Error("Error sending AnswerView to user", "error", err)
		}
	}
	return nil
}
