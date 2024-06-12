package lobbies

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/quiz"
)

// lobby is a actively running game session
type lobby struct {
	CreatedAt       time.Time
	StartedAt       time.Time
	FinishedAt      time.Time
	Quiz            *quiz.Quiz // TODO: This shouldn't be a pointer
	mu              sync.Mutex
	Host            *user
	Pin             string
	TimePerQuestion time.Duration // Time to read the question before answering is allowed
	TimeForReading  time.Duration
	Players         map[clientID]*user

	State                    lobbyState
	questionTimer            *cancellableTimer
	CurrentQuestionStartTime time.Time
	CurrentQuestionTimeout   string // ISO 8601 String
	CurrentQuestionIdx       int
	CurrentQuestion          *quiz.Question
	PlayersAnswering         int     // Number of players who haven't submitted an answer
	Leaderboard              []*user // Players sorted by score
}

type lobbyOptions struct {
	TimePerQuestion time.Duration
	TimeForReading  time.Duration
	Pin             string
}

func createLobby(options lobbyOptions) *lobby {
	// Default time per question is 30 seconds
	var timePerQuestion time.Duration
	if options.TimePerQuestion != 0 {
		timePerQuestion = options.TimePerQuestion
	} else {
		timePerQuestion = 30 * time.Second
	}

	// Default time for reading is 5 seconds
	var timeForReading time.Duration
	if options.TimeForReading != 0 {
		timeForReading = options.TimeForReading
	} else {
		timeForReading = 5 * time.Second
	}

	return &lobby{
		Host:               nil,
		Pin:                options.Pin,
		TimePerQuestion:    timePerQuestion,
		TimeForReading:     timeForReading,
		CreatedAt:          time.Now(),
		Players:            make(map[clientID]*user),
		State:              lsWaitingForPlayers,
		CurrentQuestionIdx: -1,
	}
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
		return l.endGame()
	}

	l.CurrentQuestion = l.Quiz.Questions[l.CurrentQuestionIdx]

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

// ByScore implements sort.Interface for []*user based on the Score field
// User for calculating leaderboard
type ByScore []*user

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Less(i, j int) bool { return a[i].Score < a[j].Score }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

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
		submittedAnswer := l.CurrentQuestion.Answers[player.SubmittedAnswerIdx]
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

	// Calculate leaderboard
	l.Leaderboard = make([]*user, 0, len(l.Players))
	for _, player := range l.Players {
		l.Leaderboard = append(l.Leaderboard, player)
	}
	sort.Sort(sort.Reverse(ByScore(l.Leaderboard)))

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

func (l *lobby) endGame() error {

	// Close the lobby
	l.FinishedAt = time.Now()
	if err := lobbiesRepo.DeleteLobby(l.Pin); err != nil {
		return err
	}

	// TODO: Save the game results (Eren's package)

	// Close websocket connections
	// The redirection to lobby results is handled by the client
	l.Host.writeTemplate(onFinishView, nil)
	l.Host.Conn.Close()
	for _, player := range l.Players {
		player.writeTemplate(onFinishView, nil)
		player.Conn.Close()
	}

	return nil
}
