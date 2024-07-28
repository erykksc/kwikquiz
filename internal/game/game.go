package game

import (
	"errors"
	"time"
)

type GameSettings struct {
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

type game struct {
	settings    GameSettings
	startedAt   time.Time
	endedAt     time.Time
	quiz        Quiz
	points      map[Username]int
	round       *Round
	roundNum    uint
	leaderboard []Score //Scores are sorted descending by points
}

func CreateGame(quiz Quiz, settings GameSettings) Game {
	return Game{
		&game{
			quiz:     quiz,
			points:   make(map[Username]int),
			settings: settings,
		},
	}
}

func (game *game) AddPlayer(username Username) error {
	if !game.startedAt.IsZero() {
		return ErrGameAlreadyStarted{}
	}

	_, isUsernameInGame := game.points[username]
	if isUsernameInGame {
		return errors.New("Username already in game")
	}

	game.points[username] = 0
	return nil
}

func (game *game) ChangeUsername(oldName, newName Username) error {
	oldUsernamePoints, isOldUsernameInGame := game.points[oldName]
	if !isOldUsernameInGame {
		return errors.New("Username is not in game: " + string(oldName))
	}

	if _, isNewUsernameInGame := game.points[newName]; isNewUsernameInGame {
		return errors.New("New username is already in game: " + string(newName))
	}

	game.points[newName] = oldUsernamePoints
	delete(game.points, oldName)
	return nil
}

func (game *game) RemovePlayer(username Username) error {
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

func (game *game) Start() (*Round, error) {
	if !game.startedAt.IsZero() {
		return nil, ErrGameAlreadyStarted{}
	}

	if !game.endedAt.IsZero() {
		return nil, ErrGameFinished{}
	}

	game.startedAt = time.Now()

	return game.StartRound(0)
}

func (game *game) StartNextRound() (*Round, error) {
	if !game.endedAt.IsZero() {
		return nil, ErrGameFinished{}
	}
	if game.roundNum+1 >= game.quiz.QuestionsCount() {
		return nil, ErrNoMoreQuestions{}
	}

	return game.StartRound(game.roundNum + 1)
}

// Finish finishes the game if it isn't finished already
func (game *game) Finish() error {
	if !game.endedAt.IsZero() {
		return ErrGameFinished{}
	}

	game.endedAt = time.Now()
	return nil
}

func (game game) IsFinished() bool {
	return !game.endedAt.IsZero()
}

// startRound starts a specific round in the game, first round is of index 0
func (game *game) StartRound(num uint) (*Round, error) {
	if !game.endedAt.IsZero() {
		return nil, ErrGameFinished{}
	}

	question, err := game.quiz.GetQuestion(num)
	if err != nil {
		return nil, err
	}

	newRound := CreateRound(game.GetPlayers(), question, game.settings.RoundSettings)
	game.round = newRound
	game.roundNum = num
	return newRound, newRound.Start()
}

func (game game) PlayerExists(u Username) bool {
	_, exists := game.points[u]
	return exists
}

// GetPlayers returns the players in the game
func (game game) GetPlayers() []Username {
	players := make([]Username, len(game.points))
	i := 0
	for username := range game.points {
		players[i] = username
		i++
	}

	return players
}
