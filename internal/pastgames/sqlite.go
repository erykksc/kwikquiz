package pastgames

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type RepositorySQLite struct {
	*repositorySQLite
}

func NewPastGameRepositorySQLite(db *sqlx.DB) RepositorySQLite {
	return RepositorySQLite{
		&repositorySQLite{
			db: db,
		},
	}
}

type repositorySQLite struct {
	db *sqlx.DB
}

func (repo *repositorySQLite) Initialize() error {
	const schema = `
		CREATE TABLE IF NOT EXISTS past_game (
			id INTEGER PRIMARY KEY,
			started_at DATETIME,
			ended_at DATETIME,
			quiz_title TEXT
		);

		CREATE TABLE IF NOT EXISTS player_score (
			id INTEGER PRIMARY KEY,
			past_game_id INTEGER,
			username TEXT,
			score INTEGER,
			FOREIGN KEY(past_game_id) REFERENCES past_games(id)
		);
	`
	_, err := repo.db.Exec(schema)
	return err
}
func (repo *repositorySQLite) Insert(game *PastGame) (int64, error) {
	tx, err := repo.db.Beginx()
	if err != nil {
		return 0, err
	}

	// Insert the game
	res, err := tx.NamedExec(`
        INSERT INTO past_game (id, started_at, ended_at, quiz_title)
		VALUES (:id, :started_at, :ended_at, :quiz_title)
    `, game)
	if err != nil {
		return 0, err
	}

	insertedGameID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert all scores
	for _, score := range game.Scores {
		_, err := tx.NamedExec(`
			INSERT INTO player_score (id, past_game_id, username, score)
			VALUES (:id, :past_game_id, :username, :score)
		`, &score)

		if err != nil {
			rErr := tx.Rollback()
			if rErr != nil {
				slog.Error("During handling an insert error with rollback, another error appeared", "err", rErr)
			}
			return 0, err
		}
	}

	err = tx.Commit()

	return insertedGameID, err
}

func (repo *repositorySQLite) Upsert(game *PastGame) (int64, error) {
	tx, err := repo.db.Beginx()
	if err != nil {
		return 0, err
	}

	res, err := tx.NamedExec(`
        INSERT INTO past_game (id, started_at, ended_at, quiz_title)
		VALUES (:id, :started_at, :ended_at, :quiz_title)
        ON CONFLICT(id) DO UPDATE SET
        started_at = EXCLUDED.started_at,
        ended_at = EXCLUDED.ended_at,
        quiz_title = EXCLUDED.quiz_title
    `, game)
	if err != nil {
		return 0, err
	}
	upsertedGameID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Delete all scores as updating isn't an option
	res, err = repo.db.Exec(`
		DELETE FROM player_score WHERE past_game_id = ?
	`)

	// Insert all scores
	for _, score := range game.Scores {
		_, err := tx.Exec(`
			INSERT INTO player_score (past_game_id, username, score)
			VALUES (?, ?, ?)
		`, upsertedGameID, score.Username, score.Score)

		if err != nil {
			rErr := tx.Rollback()
			if rErr != nil {
				slog.Error("During handling an insert error with rollback, another error appeared", "err", rErr)
			}
			return 0, err
		}
	}

	err = tx.Commit()

	return upsertedGameID, err
}

func (repo *repositorySQLite) GetByID(id int64) (*PastGame, error) {
	query := "SELECT * FROM past_game WHERE id=?"
	var game PastGame
	err := repo.db.Get(&game, query, id)
	if err != nil {
		return nil, err
	}
	return &game, err
}

func (repo *repositorySQLite) HydrateScores(game *PastGame) error {
	if game.ID == 0 {
		return errors.New("game ID is not set")
	}
	query := "SELECT username, score FROM player_score WHERE past_game_ID=?"
	err := repo.db.Select(&game.Scores, query, game.ID)
	return err
}

// GetAllPastGames returns all past games
// NOTE: PastGame.Scores are unhydrated, use HydrateScores to get them
func (repo *repositorySQLite) GetAll() ([]PastGame, error) {
	query := "SELECT * FROM past_game"
	var games []PastGame
	err := repo.db.Select(&games, query)
	return games, err
}

func (repo *repositorySQLite) BrowsePastGamesByID(query string) ([]PastGame, error) {
	sQuery := "SELECT * FROM past_game WHERE CAST(id AS TEXT) LIKE ?"
	var games []PastGame
	err := repo.db.Select(&games, sQuery, fmt.Sprintf("%%%s%%", query))
	return games, err
}
