package pastgames

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type RepositorySQLite struct {
	*repositorySQLite
}

func NewRepositorySQLite(db *sqlx.DB) (RepositorySQLite, error) {
	repo := RepositorySQLite{
		&repositorySQLite{
			db: db,
		},
	}
	return repo, repo.createTables()
}

type repositorySQLite struct {
	db *sqlx.DB
}

func (repo *repositorySQLite) createTables() error {
	const schema = `
		CREATE TABLE IF NOT EXISTS past_game (
			id INTEGER PRIMARY KEY,
			started_at DATETIME,
			ended_at DATETIME,
			quiz_title TEXT
		);

		CREATE TABLE IF NOT EXISTS player_score (
			id INTEGER PRIMARY KEY,
			past_game_id INTEGER REFERENCES past_game(id) ON DELETE CASCADE,
			username TEXT,
			score INTEGER
		);

		CREATE INDEX IF NOT EXISTS idx_player_score_past_game_id ON player_score(past_game_id);
	`
	_, err := repo.db.Exec(schema)
	return err
}

func (repo *repositorySQLite) Insert(game *PastGame) (int64, error) {
	if game == nil {
		return 0, errors.New("game is nil")
	}
	tx, err := repo.db.Beginx()
	if err != nil {
		return 0, err
	}
	// Rollback if no tx.Commit (if there is commit, this is no-op)
	defer tx.Rollback() //nolint

	// Insert the game
	res, err := tx.NamedExec(`
        INSERT INTO past_game (started_at, ended_at, quiz_title)
		VALUES (:started_at, :ended_at, :quiz_title)
    `, &game)
	if err != nil {
		return 0, err
	}

	insertedGameID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert all scores
	for _, score := range game.Scores {
		_, err := tx.Exec(`
			INSERT INTO player_score (past_game_id, username, score)
			VALUES (?, ?, ?)
		`, insertedGameID, score.Username, score.Score)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()

	return insertedGameID, err
}

func (repo *repositorySQLite) Upsert(game *PastGame) (int64, error) {
	if game == nil {
		return 0, errors.New("game is nil")
	}

	tx, err := repo.db.Beginx()
	if err != nil {
		return 0, err
	}
	// Rollback if no tx.Commit (if there is commit, this is no-op)
	defer tx.Rollback() //nolint

	_, err = tx.NamedExec(`
        INSERT INTO past_game (id, started_at, ended_at, quiz_title)
		VALUES (:id, :started_at, :ended_at, :quiz_title)
        ON CONFLICT(id) DO UPDATE SET
        started_at = EXCLUDED.started_at,
        ended_at = EXCLUDED.ended_at,
        quiz_title = EXCLUDED.quiz_title
    `, &game)
	if err != nil {
		return 0, err
	}

	// Delete all scores as updating isn't an option
	_, err = tx.Exec(`
		DELETE FROM player_score WHERE past_game_id = ?
	`, game.ID)
	if err != nil {
		return 0, err
	}

	// Insert all scores
	for _, score := range game.Scores {
		_, err := tx.Exec(`
			INSERT INTO player_score (past_game_id, username, score)
			VALUES (?, ?, ?)
		`, game.ID, score.Username, score.Score)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()
	return game.ID, err
}

func (repo *repositorySQLite) GetByID(id int64) (*PastGame, error) {
	query := "SELECT * FROM past_game WHERE id=?"
	var game PastGame
	err := repo.db.Get(&game, query, id)
	if err != nil {
		return nil, err
	}

	query = "SELECT username, score FROM player_score WHERE past_game_ID=?"
	err = repo.db.Select(&game.Scores, query, game.ID)
	return &game, err
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
