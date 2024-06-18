package pastgames

import (
    "sync"
)

type ErrPastGameNotFound struct{}

func (ErrPastGameNotFound) Error() string {
    return "past game not found"
}

type PastGameRepository interface {
    AddPastGame(game PastGame) error
    GetPastGameByID(id int) (PastGame, error)
    GetAllPastGames() ([]PastGame, error)
}

// InMemoryPastGameRepository In-mem store for past games
type InMemoryPastGameRepository struct {
    pastGames map[int]PastGame
    mutex     sync.RWMutex
    highestID int
}

func NewInMemoryPastGameRepository() *InMemoryPastGameRepository {
    return &InMemoryPastGameRepository{
        pastGames: make(map[int]PastGame),
        highestID: 0,
    }
}

func (repo *InMemoryPastGameRepository) AddPastGame(game PastGame) error {
    repo.mutex.Lock()
    defer repo.mutex.Unlock()

    if game.ID == 0 {
        // Assign a unique ID
        repo.highestID++
        game.ID = repo.highestID
    } else if game.ID > repo.highestID {
        repo.highestID = game.ID
    }

    repo.pastGames[game.ID] = game
    return nil
}

func (repo *InMemoryPastGameRepository) GetPastGameByID(id int) (PastGame, error) {
    repo.mutex.RLock()
    defer repo.mutex.RUnlock()

    game, ok := repo.pastGames[id]
    if !ok {
        return PastGame{}, ErrPastGameNotFound{}
    }
    return game, nil
}

func (repo *InMemoryPastGameRepository) GetAllPastGames() ([]PastGame, error) {
    repo.mutex.RLock()
    defer repo.mutex.RUnlock()

    var games []PastGame
    for _, game := range repo.pastGames {
        games = append(games, game)
    }
    return games, nil
}
