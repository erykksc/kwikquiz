package pastgames

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ErrPastGameNotFound struct{}

func (ErrPastGameNotFound) Error() string {
	return "past game not found"
}

type PastGameRepository interface {
	AddPastGame(game PastGame) (uint, error)
	GetPastGameByID(id int) (PastGame, error)
	GetAllPastGames() ([]PastGame, error)
	BrowsePastGamesByID(query string) ([]PastGame, error)
}

// InMemoryPastGameRepository In-mem store for past games
type GormPastGameRepository struct {
	DB *gorm.DB
}

func NewGormPastGameRepository(db *gorm.DB) *GormPastGameRepository {
	return &GormPastGameRepository{DB: db}
}

func (repo *GormPastGameRepository) AddPastGame(game PastGame) (uint, error) {
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

func (repo *GormPastGameRepository) GetPastGameByID(id uint) (PastGame, error) {
	var game PastGame
	result := repo.DB.Preload("Scores").First(&game, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return PastGame{}, ErrPastGameNotFound{}
	}
	return game, result.Error
}

func (repo *GormPastGameRepository) GetAllPastGames() ([]PastGame, error) {
	var games []PastGame
	result := repo.DB.Preload("Scores").Find(&games)
	return games, result.Error
}

func (repo *GormPastGameRepository) BrowsePastGamesByID(query string) ([]PastGame, error) {
	var games []PastGame
	result := repo.DB.Preload("Scores").
		Where("id::text LIKE ?", "%"+query+"%").
		Find(&games)
	return games, result.Error
}
