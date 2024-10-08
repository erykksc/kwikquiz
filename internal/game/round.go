package game

import (
	"errors"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

var ErrRoundAlreadyEnded = errors.New("Round already ended")

type RoundSettings struct {
	ReadingTime time.Duration
	AnswerTime  time.Duration
}

type Round struct {
	mu       sync.RWMutex
	question Question
	startAt  time.Time
	endedAt  time.Time
	players  map[Username]bool
	answers  map[Username]roundAnswer
	finished chan struct{} // channel that closes once a round has finished
	settings RoundSettings
}

func CreateRound(players []Username, question Question, settings RoundSettings) *Round {
	round := Round{
		question: question,
		settings: settings,
		finished: make(chan struct{}),
		players:  make(map[Username]bool),
		answers:  make(map[Username]roundAnswer),
	}
	for _, player := range players {
		round.players[player] = true
	}
	return &round
}

func (round *Round) start() error {
	round.mu.Lock()
	defer round.mu.Unlock()

	if !round.startAt.IsZero() {
		return ErrRoundAlreadyEnded
	}

	if !round.endedAt.IsZero() {
		return ErrRoundAlreadyEnded
	}

	round.startAt = time.Now()

	go func() {
		select {
		case <-time.After(round.settings.ReadingTime + round.settings.AnswerTime):
			round.mu.Lock()
			defer round.mu.Unlock()

			err := round.finishRound()
			if err != nil {
				slog.Error("Error finishing while round after timer", "err", err)
			}

		case <-round.finished:
			// Round was ended early
			return
		}
	}()

	return nil
}

func (round *Round) Question() Question {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return round.question
}

func (round *Round) Answers() map[Username]roundAnswer {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return round.answers
}

func (round *Round) StartedAt() time.Time {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return round.startAt
}

func (round *Round) HasStarted() bool {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return !round.startAt.IsZero()
}

func (round *Round) ReadingTimeout() time.Time {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return round.startAt.Add(round.settings.ReadingTime)
}

func (round *Round) Timeout() time.Time {
	round.mu.RLock()
	defer round.mu.RUnlock()
	duration := round.settings.ReadingTime + round.settings.AnswerTime
	return round.startAt.Add(duration)
}

// FinishEarly closes the finished channel, effectively ending the round early
func (round *Round) FinishEarly() error {
	round.mu.Lock()
	defer round.mu.Unlock()

	return round.finishRound()
}

// Thread unsafe
func (round *Round) finishRound() error {
	if !round.endedAt.IsZero() {
		return ErrRoundAlreadyEnded
	}

	round.endedAt = time.Now()
	close(round.finished)
	return nil
}

func (round *Round) Finished() chan struct{} {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return round.finished
}

func (round *Round) HasFinished() bool {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return !round.endedAt.IsZero()
}

type roundAnswer struct {
	Index       int
	SubmittedAt time.Time
}

func (round *Round) submitAnswer(player Username, answerIndex int) error {
	round.mu.Lock()
	defer round.mu.Unlock()

	rAnswer := roundAnswer{
		Index:       answerIndex,
		SubmittedAt: time.Now(),
	}

	// Check if index is valid
	if !round.question.IsAnswerValid(answerIndex) {
		return errors.New("invalid answer index")
	}

	// Check if player is in the round
	if _, isInPlayers := round.players[player]; !isInPlayers {
		return errors.New("player not in round")
	}

	if round.startAt.IsZero() {
		return errors.New("round has not started")
	}

	if rAnswer.SubmittedAt.Before(round.startAt.Add(round.settings.ReadingTime)) {
		return errors.New("answer submitted before answering allowed")
	}

	// Check if answer was submitted after the round ended
	if !round.endedAt.IsZero() {
		return errors.New("answer submitted after round ended")
	}

	// Check if player has already submitted an answer
	previousAnswer, playerAlreadySubmitted := round.answers[player]
	if playerAlreadySubmitted {
		return errors.New("player has already submitted an answer: answerSubmitted=" + strconv.Itoa(previousAnswer.Index))
	}

	round.answers[player] = rAnswer

	// If all players have answered, finish the round
	if len(round.answers) == len(round.players) {
		err := round.finishRound()
		if err != nil {
			slog.Error("Error finishing round after all players answered", "err", err)
		}
	}

	return nil
}

func (round *Round) PlayersAnswering() int {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return len(round.players) - len(round.answers)
}

func (round *Round) PlayerAnswers() map[Username]roundAnswer {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return round.answers
}

// Results returns the scores of the players in the round, sorted by points in descending order
func (round *Round) Results() (map[Username]int, error) {
	round.mu.RLock()
	defer round.mu.RUnlock()

	if round.endedAt.IsZero() {
		return nil, errors.New("round has not ended")
	}

	// scores := make([]Score, len(round.players))
	scores := make(map[Username]int, len(round.players))
	for username := range round.players {
		answer, hasAnswered := round.answers[username]
		pointsAwarded := 0

		if hasAnswered && round.question.IsAnswerCorrect(answer.Index) {
			time2Answer := answer.SubmittedAt.Sub(round.startAt.Add(round.settings.ReadingTime))
			if time2Answer < time.Millisecond*500 {
				// Maximum points for answering in less than 500ms
				pointsAwarded = 1000
			} else {
				pointsAwarded = int((1 - (float64(time2Answer) / float64(round.settings.AnswerTime) / 2.0)) * 1000)
			}
		}
		scores[username] = pointsAwarded
	}
	return scores, nil
}
