package game

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"
)

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
	RoundSettings
}

func CreateRound(players []Username, question Question, settings RoundSettings) *Round {
	round := Round{
		question:      question,
		RoundSettings: settings,
		finished:      make(chan struct{}),
		players:       make(map[Username]bool),
		answers:       make(map[Username]roundAnswer),
	}
	for _, player := range players {
		round.players[player] = true
	}
	return &round
}

func (round *Round) Start() error {
	round.mu.Lock()
	defer round.mu.Unlock()

	if !round.startAt.IsZero() {
		return errors.New("Round already started")
	}

	if !round.endedAt.IsZero() {
		return errors.New("Round already ended")
	}

	round.startAt = time.Now()

	go func() {
		select {
		case <-time.After(round.ReadingTime + round.AnswerTime):
			round.mu.Lock()
			defer round.mu.Unlock()

			round.endedAt = time.Now()
			close(round.finished)
		case <-round.finished:
			// Round was ended early
			return
		}
	}()

	return nil
}

// FinishEarly closes the finished channel, effectively ending the round early
func (round *Round) FinishEarly() error {
	round.mu.Lock()
	defer round.mu.Unlock()

	if !round.endedAt.IsZero() {
		return errors.New("Round already ended")
	}

	round.endedAt = time.Now()
	close(round.finished)

	return nil
}

func (round *Round) Finished() chan struct{} {
	return round.finished
}

type roundAnswer struct {
	index       int
	submittedAt time.Time
}

func (round *Round) SubmitAnswer(player Username, answerIndex int) error {
	round.mu.Lock()
	defer round.mu.Unlock()

	rAnswer := roundAnswer{
		index:       answerIndex,
		submittedAt: time.Now(),
	}

	// Check if player is in the round
	if _, isInPlayers := round.players[player]; !isInPlayers {
		return errors.New("player not in round")
	}

	if round.startAt.IsZero() {
		return errors.New("round has not started")
	}

	if rAnswer.submittedAt.Before(round.startAt.Add(round.ReadingTime)) {
		return errors.New("answer submitted before answering allowed")
	}

	// Check if answer was submitted after the round ended
	if !round.endedAt.IsZero() {
		return errors.New("answer submitted after round ended")
	}

	// Check if player has already submitted an answer
	previousAnswer, playerAlreadySubmitted := round.answers[player]
	if playerAlreadySubmitted {
		return errors.New("player has already submitted an answer: answerSubmitted=" + strconv.Itoa(previousAnswer.index))
	}

	round.answers[player] = rAnswer
	return nil
}

// GetResults returns the scores of the players in the round, sorted by points in descending order
func (round *Round) GetResults() ([]Score, error) {
	round.mu.RLock()
	defer round.mu.RUnlock()

	if round.endedAt.IsZero() {
		return nil, errors.New("round has not ended")
	}

	scores := make([]Score, len(round.players))
	i := 0
	for username := range round.players {
		answer, hasAnswered := round.answers[username]
		pointsAwarded := 0

		if hasAnswered && round.question.IsAnswerCorrect(answer.index) {
			time2Answer := answer.submittedAt.Sub(round.startAt.Add(round.ReadingTime))
			if time2Answer < time.Millisecond*500 {
				// Maximum points for answering in less than 500ms
				pointsAwarded = 1000
			} else {
				pointsAwarded = int((1 - (float64(time2Answer) / float64(round.AnswerTime) / 2.0)) * 1000)
			}
		}
		scores[i] = Score{
			Points: pointsAwarded,
			Player: username,
		}
		i++
	}

	// Sort by points
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Points > scores[j].Points
	})

	return scores, nil
}

func (round *Round) PlayersStillAnswering() []Username {
	round.mu.RLock()
	defer round.mu.RUnlock()

	answering := []Username{}
	for username := range round.players {
		_, answered := round.answers[username]

		if !answered {
			answering = append(answering, username)
		}
	}

	return answering
}
