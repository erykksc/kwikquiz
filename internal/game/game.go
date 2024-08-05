package game

import (
	"errors"
	"log/slog"
	"slices"
	"sync"
	"time"
)

type GameSettings struct {
	Quiz Quiz
	RoundSettings
}

type Game struct {
	*game
}

type ErrGameAlreadyStarted struct{}

func (_ ErrGameAlreadyStarted) Error() string {
	return "Game already started"
}

type ErrGameFinished struct{}

func (_ ErrGameFinished) Error() string {
	return "Game already finished"
}

type ErrNoMoreQuestions struct{}

func (_ ErrNoMoreQuestions) Error() string {
	return "No more questions"
}

type ErrInvalidUsername struct {
	Reason string
}

func (err ErrInvalidUsername) Error() string {
	return "Invalid username: " + err.Reason
}

type game struct {
	mu        sync.RWMutex
	settings  GameSettings
	startedAt time.Time
	endedAt   time.Time
	quiz      Quiz
	points    map[Username]int
	Round     *Round
	roundNum  int
}

func CreateGame(settings GameSettings) Game {
	game := Game{
		&game{
			points: make(map[Username]int),
		},
	}

	err := game.UpdateSettings(settings)
	if err != nil {
		// This should never happen
		panic(err)
	}
	return game
}

func (game *game) Settings() GameSettings {
	game.mu.RLock()
	defer game.mu.RUnlock()

	return game.settings
}

func (game *game) UpdateSettings(settings GameSettings) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if !game.startedAt.IsZero() {
		return ErrGameAlreadyStarted{}
	}

	game.quiz = settings.Quiz
	game.settings = settings
	return nil
}

func (game *game) Quiz() Quiz {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.quiz
}

func (game *game) StartedAt() time.Time {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.startedAt
}

func (game *game) EndedAt() time.Time {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.endedAt
}

func (game *game) AddPlayer(username Username) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if !game.startedAt.IsZero() {
		return ErrGameAlreadyStarted{}
	}

	isValid, err := username.IsValid()
	if !isValid {
		return ErrInvalidUsername{
			Reason: err.Error(),
		}
	}

	_, isUsernameInGame := game.points[username]
	if isUsernameInGame {
		return errors.New("Username already in game")
	}

	game.points[username] = 0
	return nil
}

func (game *game) ChangeUsername(oldName, newName Username) error {
	if oldName == newName {
		return nil
	}

	game.mu.Lock()
	defer game.mu.Unlock()
	oldUsernamePoints, isOldUsernameInGame := game.points[oldName]
	if !isOldUsernameInGame {
		return errors.New("Username is not in game: " + string(oldName))
	}

	if _, isNewUsernameInGame := game.points[newName]; isNewUsernameInGame {
		return errors.New("New username is already in game: " + string(newName))
	}

	isValid, err := newName.IsValid()
	if !isValid {
		return ErrInvalidUsername{
			Reason: err.Error(),
		}
	}

	game.points[newName] = oldUsernamePoints
	delete(game.points, oldName)
	return nil
}

func (game *game) RemovePlayer(username Username) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if !game.startedAt.IsZero() {
		return ErrGameAlreadyStarted{}
	}

	_, isUsernameInGame := game.points[username]
	if !isUsernameInGame {
		return errors.New("Username not in game")
	}

	delete(game.points, username)
	return nil
}

func (game *game) Start() error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if !game.startedAt.IsZero() {
		return ErrGameAlreadyStarted{}
	}

	if !game.endedAt.IsZero() {
		return ErrGameFinished{}
	}

	err := game.startRound(0)
	if err != nil {
		return err
	}

	game.startedAt = time.Now()
	return nil
}

func (game *game) StartNextRound() error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if !game.endedAt.IsZero() {
		return ErrGameFinished{}
	}

	if game.startedAt.IsZero() {
		return errors.New("Game not started")
	}

	if !game.Round.HasFinished() {
		return errors.New("Previous round not finished")
	}

	if game.roundNum+1 >= game.quiz.QuestionsCount() {
		return ErrNoMoreQuestions{}
	}

	return game.startRound(game.roundNum + 1)
}

// Finish finishes the game if it isn't finished already
func (game *game) Finish() error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if !game.endedAt.IsZero() {
		return ErrGameFinished{}
	}

	game.endedAt = time.Now()
	return nil
}

func (game *game) HasStarted() bool {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return !game.startedAt.IsZero()
}

func (game *game) HasEnded() bool {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return !game.endedAt.IsZero()
}

func (game *game) IsFinished() bool {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return !game.endedAt.IsZero()
}

func (game *game) RoundFinished() (chan struct{}, error) {
	game.mu.RLock()
	defer game.mu.RUnlock()

	if game.Round == nil {
		return nil, errors.New("Not in round")
	}

	return game.Round.Finished(), nil
}

func (game *game) StartRound(num int) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.startRound(num)
}

// startRound starts a specific round in the game, first round is of index 0
// Thread unsafe
func (game *game) startRound(num int) error {
	if game.Round != nil {
		if !game.Round.HasFinished() {
			return errors.New("Round not finished")
		}
	}

	if !game.endedAt.IsZero() {
		return ErrGameFinished{}
	}

	if len(game.points) == 0 {
		return errors.New("No players in game")
	}

	question, err := game.quiz.GetQuestion(num)
	if err != nil {
		return err
	}

	newRound := CreateRound(game.players(), question, game.settings.RoundSettings)
	game.Round = newRound
	game.roundNum = num
	err = newRound.start()
	if err != nil {
		return err
	}

	go func() {
		<-game.Round.Finished()
		slog.Debug("Round finished, adding points")
		game.mu.Lock()

		results, err := game.Round.Results()
		if err != nil {
			slog.Error("Error getting round results", "err", err)
			return
		}
		for username, points := range results {
			game.points[username] = points
		}

		game.mu.Unlock()
	}()
	return nil
}

func (game *game) FinishRoundEarly() error {
	game.mu.RLock()
	defer game.mu.RUnlock()
	if game.Round == nil {
		return errors.New("Not in round")
	}

	return game.Round.FinishEarly()
}

func (game *game) PlayerInGame(u Username) bool {
	game.mu.RLock()
	defer game.mu.RUnlock()
	_, exists := game.points[u]
	return exists
}

// Players returns the players in the game
func (game *game) Players() []Username {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.players()
}

// Thread unsafe
func (game *game) players() []Username {
	players := make([]Username, len(game.points))
	i := 0
	for username := range game.points {
		players[i] = username
		i++
	}

	return players
}

// RoundNum returns current round number
func (game *game) RoundNum() int {
	game.mu.RLock()
	defer game.mu.RUnlock()
	if game.startedAt.IsZero() {
		return -1
	}
	return int(game.roundNum)
}

func (game *game) Scores() map[Username]int {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.points
}

func (game *game) Leaderboard() []Score {
	game.mu.RLock()
	defer game.mu.RUnlock()
	leaderboard := make([]Score, len(game.points))
	for username, points := range game.points {
		leaderboard = append(leaderboard, Score{
			Points:   points,
			Username: username,
		})
	}
	// sort leaderboard
	slices.SortFunc(leaderboard, func(i, j Score) int {
		return i.Points - j.Points
	})

	return leaderboard
}

func (game *game) LastRoundPoints() (map[Username]int, error) {
	if game.Round == nil {
		return nil, errors.New("There is no last round")
	}

	return game.Round.Results()
}

func (game *game) SubmitAnswer(username Username, answerIndex int) error {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.Round == nil {
		return errors.New("Not in round")
	}
	return game.Round.submitAnswer(username, answerIndex)
}

// InRound returns true if the round is running (players are answering)
func (game *game) InRound() bool {
	game.mu.RLock()
	defer game.mu.RUnlock()
	return game.Round.HasStarted() && !game.Round.HasFinished()
}
