package game

type User struct {
	ID        string
	Name      string
	Anonymous bool
}

type GameID string

type Game struct {
	ID       GameID
	Host     User
	QuizName string
}

type ErrGameNotFound struct{}

func (ErrGameNotFound) Error() string {
	return "game not found"
}

type ErrGameAlreadyExists struct{}

func (ErrGameAlreadyExists) Error() string {
	return "game already exists"
}

type GameRepository interface {
	AddGame(game Game) error
	UpdateGame(game Game) error
	GetGame(id GameID) (Game, error)
	DeleteGame(id GameID) error
	GetAllGames() ([]Game, error)
}

// In-memory store for games
type inMemoryGameRepository struct {
	games map[GameID]Game
}

func NewInMemoryGameRepository() GameRepository {
	return &inMemoryGameRepository{
		games: make(map[GameID]Game),
	}
}

func (s *inMemoryGameRepository) AddGame(game Game) error {
	if _, ok := s.games[game.ID]; ok {
		return ErrGameAlreadyExists{}
	}

	s.games[game.ID] = game
	return nil
}

func (s *inMemoryGameRepository) UpdateGame(game Game) error {
	if _, ok := s.games[game.ID]; !ok {
		return ErrGameNotFound{}
	}

	s.games[game.ID] = game
	return nil
}

func (s *inMemoryGameRepository) GetGame(id GameID) (Game, error) {
	game, ok := s.games[id]
	if !ok {
		return Game{}, ErrGameNotFound{}
	}

	return game, nil
}

func (s *inMemoryGameRepository) DeleteGame(id GameID) error {
	if _, ok := s.games[id]; !ok {
		return ErrGameNotFound{}
	}

	delete(s.games, id)
	return nil
}

func (s *inMemoryGameRepository) GetAllGames() ([]Game, error) {
	var games []Game
	for _, game := range s.games {
		games = append(games, game)
	}

	return games, nil
}
