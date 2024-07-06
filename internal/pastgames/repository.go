package pastgames

import (
	"errors"
	"github.com/erykksc/kwikquiz/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ErrPastGameNotFound struct{}

func (ErrPastGameNotFound) Error() string {
	return "past game not found"
}

type PastGameRepository interface {
	AddPastGame(game models.PastGame) (uint, error)
	GetPastGameByID(id int) (PastGame, error)
	GetAllPastGames() ([]models.PastGame, error)
}

// InMemoryPastGameRepository In-mem store for past games
type GormPastGameRepository struct {
	DB *gorm.DB
}

func NewGormPastGameRepository(db *gorm.DB) *GormPastGameRepository {
	return &GormPastGameRepository{DB: db}
}

func (repo *GormPastGameRepository) AddPastGame(game models.PastGame) (uint, error) {
	result := repo.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&game)
	if result == nil || result.Error != nil {
		return 0, result.Error
	}
	return game.ID, nil
}

func (repo *GormPastGameRepository) GetPastGameByID(id uint) (models.PastGame, error) {
	var game models.PastGame
	result := repo.DB.Preload("Scores").First(&game, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.PastGame{}, ErrPastGameNotFound{}
	}
	return game, result.Error
}

func (repo *GormPastGameRepository) GetAllPastGames() ([]models.PastGame, error) {
	var games []models.PastGame
	result := repo.DB.Preload("Scores").Find(&games)
	return games, result.Error
}
