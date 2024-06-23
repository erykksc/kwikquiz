package lobbies

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/pastgames"
	"github.com/erykksc/kwikquiz/internal/quiz"
)

// lobby is a actively running game session
type lobby struct {
	mu                       sync.Mutex
	CreatedAt                time.Time
	StartedAt                time.Time
	FinishedAt               time.Time
	Quiz                     quiz.Quiz
	Host                     *user
	Pin                      string
	TimePerQuestion          time.Duration // Time for players to answer a question
	TimeForReading           time.Duration // Time to read the question before answering is allowed
	Players                  map[clientID]*user
	State                    LobbyState
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
	Quiz            quiz.Quiz
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
		Pin:             options.Pin, // If it's empty, it will be generated by repository
		TimePerQuestion: timePerQuestion,
		TimeForReading:  timeForReading,
		CreatedAt:       time.Now(),
		Players:         make(map[clientID]*user),
		State:           LsWaitingForPlayers,
		Quiz:            options.Quiz,
	}
}

func (l *lobby) startGame() error {
	l.StartedAt = time.Now()
	// Next question increments the index, so we start at -1
	l.CurrentQuestionIdx = -1

	return l.startNextQuestion()
}

func (l *lobby) startNextQuestion() error {
	l.State = LsQuestion
	l.CurrentQuestionIdx++
	l.CurrentQuestionStartTime = time.Now()
	l.CurrentQuestionTimeout = l.CurrentQuestionStartTime.Add(l.TimePerQuestion).Format(time.RFC3339)
	l.PlayersAnswering = len(l.Players)

	// Reset Player answers
	l.Host.SubmittedAnswerIdx = -1
	l.Host.AnswerSubmissionTime = time.Time{}
	for _, player := range l.Players {
		player.SubmittedAnswerIdx = -1
		player.AnswerSubmissionTime = time.Time{}
	}

	// Check if the game has finished
	if l.CurrentQuestionIdx == len(l.Quiz.Questions) {
		return l.endGame()
	}

	l.CurrentQuestion = &l.Quiz.Questions[l.CurrentQuestionIdx]

	// Start the question timer
	l.questionTimer = NewCancellableTimer(l.TimePerQuestion)
	go func() {
		select {
		case <-l.questionTimer.timer.C:
			// Time completed
			slog.Debug("Question timeout reached")
			l.mu.Lock()
			l.showAnswer()
			l.mu.Unlock()
			break
		case <-l.questionTimer.cancelChan:
			// Timer cancelled
			slog.Debug("Question timer cancelled")
			// Doesnt require a lock because cancel
			// can only be triggered by an event
			// and all events handlers are locked
			l.showAnswer()
			break
		}
	}()

	// Send question view to all
	vData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(QuestionView, vData); err != nil {
		slog.Error("Error sending QuestionView to host", "error", err)
	}
	for _, player := range l.Players {
		vData.User = player
		err := player.writeTemplate(QuestionView, vData)
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
	l.State = LsAnswer

	// Add points to players
	for _, player := range l.Players {
		// Check if player submitted an answer
		if player.SubmittedAnswerIdx == -1 {
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
				player.NewPoints = int((1 - (float64(timeToAnswer) / float64(l.TimePerQuestion) / 2.0)) * 1000)
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
	vData := ViewData{
		Lobby: l,
		User:  l.Host,
	}
	if err := l.Host.writeTemplate(AnswerView, vData); err != nil {
		slog.Error("Error sending AnswerView to host", "error", err)
	}

	for _, player := range l.Players {
		vData.User = player
		if err := player.writeTemplate(AnswerView, vData); err != nil {
			slog.Error("Error sending AnswerView to user", "error", err)
		}
	}
	return nil
}

func (l *lobby) endGame() error {
	l.FinishedAt = time.Now()
	if err := lobbiesRepo.DeleteLobby(l.Pin); err != nil {
		return err
	}

	scores := make([]pastgames.PlayerScore, 0, len(l.Players)+1)
	for _, player := range l.Leaderboard {
		scores = append(scores, pastgames.PlayerScore{
			Username: player.Username,
			Score:    player.Score,
		})
	}

	pastGame := pastgames.PastGame{
		StartedAt: l.StartedAt,
		EndedAt:   time.Now(),
		QuizTitle: l.Quiz.Title,
		Scores:    scores,
	}
	id, err := pastgames.PastGamesRepo.AddPastGame(pastGame)
	if err != nil {
		return err
	}

	data := OnFinishData{
		PastGameID: id,
		ViewData: ViewData{
			Lobby: l,
			User:  l.Host,
		},
	}

	// Close websocket connections
	// The redirection to lobby results is handled by the view
	l.Host.writeTemplate(onFinishView, data)
	l.Host.Conn.Close()
	for _, player := range l.Players {
		data.User = player
		err = player.writeTemplate(onFinishView, data)
		if err != nil {
			slog.Error("Error sending OnFinishView to user", "error", err)
		}
		player.Conn.Close()
	}
	return nil
}
