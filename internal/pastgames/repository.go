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
	GetPastGameByID(id int) (models.PastGame, error)
	GetAllPastGames() ([]models.PastGame, error)
	BrowsePastGamesByID(query string) ([]models.PastGame, error)
}

// InMemoryPastGameRepository In-mem store for past games
type GormPastGameRepository struct {
	DB *gorm.DB
}

func NewGormPastGameRepository(db *gorm.DB) *GormPastGameRepository {
	return &GormPastGameRepository{DB: db}
}

func (repo *GormPastGameRepository) AddPastGame(game models.PastGame) (uint, error) {
	var gameID uint
	err := repo.DB.Transaction(func(tx *gorm.DB) error {
		// Create or update the past game
		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&game).Error; err != nil {
			return err
		}
		gameID = game.ID

		// Ensure the unique index exists on player_scores
		if err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_past_game_user ON player_scores (past_game_id, username)").Error; err != nil {
			return err
		}

		// Add or update player scores
		for i := range game.Scores {
			game.Scores[i].PastGameID = gameID
			if err := tx.Exec(`
                INSERT INTO player_scores (past_game_id, username, score, created_at, updated_at)
                VALUES (?, ?, ?, NOW(), NOW())
                ON CONFLICT (past_game_id, username)
                DO UPDATE SET score = EXCLUDED.score, updated_at = NOW()
            `, game.Scores[i].PastGameID, game.Scores[i].Username, game.Scores[i].Score).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return gameID, nil
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

func (repo *GormPastGameRepository) BrowsePastGamesByID(query string) ([]models.PastGame, error) {
	var games []models.PastGame
	result := repo.DB.Preload("Scores").
		Where("id::text LIKE ?", "%"+query+"%").
		Find(&games)
	return games, result.Error
}
