package pastgames

import (
	"log"
	"math"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func setup() (RepositorySQLite, func()) {
	// Create a new instance of repositorySQLite
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		log.Fatalln(err)
	}

	repo, err := NewRepositorySQLite(db)
	if err != nil {
		log.Fatalln(err)
	}

	// Return the repository and a teardown function
	return repo, func() {
		db.Close()
	}
}

func TestRepositorySQLite_Insert(t *testing.T) {
	repo, teardown := setup()
	defer teardown()

	t.Run("insert nil game", func(t *testing.T) {
		_, err := repo.Insert(nil)
		if err == nil {
			t.Error("Expected an error, got nil")
		}
		if err.Error() != "game is nil" {
			t.Errorf("Expected error 'game is nil', got '%s'", err.Error())
		}
	})

	t.Run("insert valid game", func(t *testing.T) {
		game := &PastGame{
			StartedAt: time.Now().Add(-time.Hour),
			EndedAt:   time.Now(),
			QuizTitle: "Test Quiz",
			Scores: []PlayerScore{
				{
					Username: "player1",
					Score:    10,
				},
				{
					Username: "player2",
					Score:    15,
				},
			},
		}

		id, err := repo.Insert(game)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if id == 0 {
			t.Error("Expected non-zero ID, got 0")
		}

		// Verify the inserted data
		insertedGame, err := repo.GetByID(id)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
		if !game.StartedAt.Equal(insertedGame.StartedAt) {
			t.Errorf("Expected started_at '%s', got '%s'", game.StartedAt, insertedGame.StartedAt)
		}
		if !game.EndedAt.Equal(insertedGame.EndedAt) {
			t.Errorf("Expected ended_at '%s', got '%s'", game.EndedAt, insertedGame.EndedAt)
		}
		if game.QuizTitle != insertedGame.QuizTitle {
			t.Errorf("Expected quiz_title '%s', got '%s'", game.QuizTitle, insertedGame.QuizTitle)
		}
		// Compare length of scores
		if len(game.Scores) != len(insertedGame.Scores) {
			t.Errorf("Expected %d scores, got %d", len(game.Scores), len(insertedGame.Scores))
			return
		}
		// Assuming scores are sorted by username
		prevScore := math.MinInt
		for i, score := range game.Scores {
			if score.Username != insertedGame.Scores[i].Username {
				t.Errorf("Expected username '%s', got '%s'", score.Username, insertedGame.Scores[i].Username)
			}
			if score.Score != insertedGame.Scores[i].Score {
				t.Errorf("Expected score '%d', got '%d'", score.Score, insertedGame.Scores[i].Score)
			}

			if score.Score < prevScore {
				t.Errorf("Expected scores to be sorted, got '%d' after '%d'", score.Score, prevScore)
			}
			prevScore = score.Score
		}
	})
}

func TestRepositorySQLite_Upsert(t *testing.T) {
	repo, teardown := setup()
	defer teardown()

	t.Run("upsert nil game", func(t *testing.T) {
		_, err := repo.Upsert(nil)
		if err == nil {
			t.Error("Expected an error, got nil")
		}
		if err.Error() != "game is nil" {
			t.Errorf("Expected error 'game is nil', got '%s'", err.Error())
		}
	})

	t.Run("upsert valid game", func(t *testing.T) {
		game := &PastGame{
			ID:        123,
			StartedAt: time.Now().Add(-time.Hour),
			EndedAt:   time.Now(),
			QuizTitle: "Test Quiz",
			Scores: []PlayerScore{
				{
					Username: "player1",
					Score:    10,
				},
				{
					Username: "player2",
					Score:    15,
				},
			},
		}

		id, err := repo.Upsert(game)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if id != 123 {
			t.Errorf("Expected ID 123, got %d", id)
		}

		// Verify the upserted data
		upsertedGame, err := repo.GetByID(id)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
		if !game.StartedAt.Equal(upsertedGame.StartedAt) {
			t.Errorf("Expected started_at '%s', got '%s'", game.StartedAt, upsertedGame.StartedAt)
		}
		if !game.EndedAt.Equal(upsertedGame.EndedAt) {
			t.Errorf("Expected ended_at '%s', got '%s'", game.EndedAt, upsertedGame.EndedAt)
		}
		if game.QuizTitle != upsertedGame.QuizTitle {
			t.Errorf("Expected quiz_title '%s', got '%s'", game.QuizTitle, upsertedGame.QuizTitle)
		}

		if len(game.Scores) != len(upsertedGame.Scores) {
			t.Errorf("Expected %d scores, got %d", len(game.Scores), len(upsertedGame.Scores))
			return
		}

		prevScore := math.MinInt
		for i, score := range game.Scores {
			if score.Username != upsertedGame.Scores[i].Username {
				t.Errorf("Expected username '%s', got '%s'", score.Username, upsertedGame.Scores[i].Username)
			}
			if score.Score != upsertedGame.Scores[i].Score {
				t.Errorf("Expected score '%d', got '%d'", score.Score, upsertedGame.Scores[i].Score)
			}
			if score.Score < prevScore {
				t.Errorf("Expected scores to be sorted, got '%d' after '%d'", score.Score, prevScore)
			}
			prevScore = score.Score
		}
	})
}
