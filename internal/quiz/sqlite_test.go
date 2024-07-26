package quiz

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestRepositorySQLite(t *testing.T) {
	testQuiz := Quiz{
		Title:       "Test Quiz",
		Password:    "password",
		Description: "This is a test quiz",
		Questions: []Question{
			{
				Text: "What is the capital of France?",
				Answers: []Answer{
					{Text: "Paris", IsCorrect: true},
					{Text: "London", IsCorrect: false},
				},
			},
		},
	}

	newDB := func() *sqlx.DB {
		db, err := sqlx.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		return db
	}

	t.Run("Initialize", func(t *testing.T) {
		db := newDB()
		defer db.Close()

		repo := NewRepositorySQLite(db)
		err := repo.Initialize()
		if err != nil {
			t.Errorf("Failed to initialize repository: %v", err)
		}
	})

	newRepo := func(db *sqlx.DB) RepositorySQLite {
		repo := NewRepositorySQLite(db)
		err := repo.Initialize()
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}
		return repo
	}

	t.Run("insert nil quiz", func(t *testing.T) {
		db := newDB()
		defer db.Close()
		repo := newRepo(db)

		_, err := repo.Insert(nil)
		assert.Error(t, err)
		assert.Equal(t, "quiz is nil", err.Error())
	})

	t.Run("insert and get valid quiz", func(t *testing.T) {
		db := newDB()
		defer db.Close()
		repo := newRepo(db)
		assert := assert.New(t)

		id, err := repo.Insert(&testQuiz)
		assert.NoError(err)

		// Verify the inserted data
		insertedQuiz, err := repo.Get(id)
		assert.NoError(err)
		assert.Equal(id, insertedQuiz.ID)
		assert.Equal(testQuiz.Title, insertedQuiz.Title)
		assert.Equal(testQuiz.Password, insertedQuiz.Password)
		assert.Equal(testQuiz.Description, insertedQuiz.Description)
		assert.Equal(testQuiz.Questions[0].Text, insertedQuiz.Questions[0].Text)
		assert.Equal(testQuiz.Questions[0].Answers[0].Text, insertedQuiz.Questions[0].Answers[0].Text)
		assert.Equal(testQuiz.Questions[0].Answers[0].IsCorrect, insertedQuiz.Questions[0].Answers[0].IsCorrect)
	})

	t.Run("Upsert", func(t *testing.T) {
		db := newDB()
		defer db.Close()
		repo := newRepo(db)

		testQuiz.ID = 123
		id, err := repo.Upsert(&testQuiz)

		assert := assert.New(t)

		assert.NoError(err)
		assert.Equal(int64(123), id)

		// Verify the inserted data
		upsertedQuiz, err := repo.Get(id)
		assert.NoError(err)
		assert.Equal(id, upsertedQuiz.ID)
		assert.Equal(testQuiz.Title, upsertedQuiz.Title)
		assert.Equal(testQuiz.Password, upsertedQuiz.Password)
		assert.Equal(testQuiz.Description, upsertedQuiz.Description)
		assert.Equal(testQuiz.Questions[0].Text, upsertedQuiz.Questions[0].Text)
		assert.Equal(testQuiz.Questions[0].Answers[0].Text, upsertedQuiz.Questions[0].Answers[0].Text)
		assert.Equal(testQuiz.Questions[0].Answers[0].IsCorrect, upsertedQuiz.Questions[0].Answers[0].IsCorrect)
	})

	t.Run("DeleteQuiz", func(t *testing.T) {
		db := newDB()
		defer db.Close()
		repo := newRepo(db)

		quiz := &Quiz{Title: "Test Quiz", Password: "1234", Description: "This is a test quiz"}
		id, _ := repo.Insert(quiz)
		err := repo.Delete(id)
		if err != nil {
			t.Errorf("Failed to delete quiz: %v", err)
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		db := newDB()
		defer db.Close()
		repo := newRepo(db)

		quiz1 := &Quiz{Title: "Test Quiz 1", Password: "1234", Description: "This is a test quiz"}
		quiz2 := &Quiz{Title: "Test Quiz 2", Password: "1234", Description: "This is a test quiz"}
		repo.Insert(quiz1)
		repo.Insert(quiz2)
		quizzes, err := repo.GetAll()
		if err != nil {
			t.Errorf("Failed to get all quizzes: %v", err)
		}
		if len(quizzes) != 2 {
			t.Errorf("Wrong number of quizzes returned, expected 2, got: %d", len(quizzes))
		}
	})
}
