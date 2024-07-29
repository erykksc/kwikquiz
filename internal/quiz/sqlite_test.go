package quiz

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestRepositorySQLite(t *testing.T) {
	testQuiz := Quiz{
		TitleField:  "Test Quiz",
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

	t.Run("call NewRepositorySQLite", func(t *testing.T) {
		db := newDB()
		defer db.Close()

		_, err := NewRepositorySQLite(db)
		if err != nil {
			t.Errorf("Failed to initialize repository: %v", err)
		}
	})

	newRepo := func(db *sqlx.DB) RepositorySQLite {
		repo, err := NewRepositorySQLite(db)
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
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("insert and get valid quiz", func(t *testing.T) {
		db := newDB()
		defer db.Close()
		repo := newRepo(db)

		id, err := repo.Insert(&testQuiz)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify the inserted data
		insertedQuiz, err := repo.Get(id)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if id != insertedQuiz.ID {
			t.Errorf("Expected ID %v, got: %v", id, insertedQuiz.ID)
		}

		if testQuiz.TitleField != insertedQuiz.TitleField {
			t.Errorf("Expected Title %s, got %s", testQuiz.TitleField, insertedQuiz.TitleField)
		}

		if testQuiz.Password != insertedQuiz.Password {
			t.Errorf("Expected Password %s, got %s", testQuiz.Password, insertedQuiz.Password)
		}

		if testQuiz.Description != insertedQuiz.Description {
			t.Errorf("Expected Description %s, got %s", testQuiz.Description, insertedQuiz.Description)
		}

		if testQuiz.Questions[0].Text != insertedQuiz.Questions[0].Text {
			t.Errorf("Expected Question Text %s, got %s", testQuiz.Questions[0].Text, insertedQuiz.Questions[0].Text)
		}

		if testQuiz.Questions[0].Answers[0].Text != insertedQuiz.Questions[0].Answers[0].Text {
			t.Errorf("Expected Answer Text %s, got %s", testQuiz.Questions[0].Answers[0].Text, insertedQuiz.Questions[0].Answers[0].Text)
		}

		if testQuiz.Questions[0].Answers[0].IsCorrect != insertedQuiz.Questions[0].Answers[0].IsCorrect {
			t.Errorf("Expected Answer IsCorrect %v, got: %v", testQuiz.Questions[0].Answers[0].IsCorrect, insertedQuiz.Questions[0].Answers[0].IsCorrect)
		}
	})

	t.Run("Upsert", func(t *testing.T) {
		db := newDB()
		defer db.Close()
		repo := newRepo(db)

		testQuiz.ID = 123
		id, err := repo.Upsert(&testQuiz)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if id != int64(123) {
			t.Errorf("Expected ID %v, got: %v", 123, id)
		}

		// Verify the inserted data
		upsertedQuiz, err := repo.Get(id)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if id != upsertedQuiz.ID {
			t.Errorf("Expected ID %v, got: %v", id, upsertedQuiz.ID)
		}

		if testQuiz.TitleField != upsertedQuiz.TitleField {
			t.Errorf("Expected Title %s, got %s", testQuiz.TitleField, upsertedQuiz.TitleField)
		}

		if testQuiz.Password != upsertedQuiz.Password {
			t.Errorf("Expected Password %s, got %s", testQuiz.Password, upsertedQuiz.Password)
		}

		if testQuiz.Description != upsertedQuiz.Description {
			t.Errorf("Expected Description %s, got %s", testQuiz.Description, upsertedQuiz.Description)
		}

		if testQuiz.Questions[0].Text != upsertedQuiz.Questions[0].Text {
			t.Errorf("Expected Question Text %s, got %s", testQuiz.Questions[0].Text, upsertedQuiz.Questions[0].Text)
		}

		if testQuiz.Questions[0].Answers[0].Text != upsertedQuiz.Questions[0].Answers[0].Text {
			t.Errorf("Expected Answer Text %s, got %s", testQuiz.Questions[0].Answers[0].Text, upsertedQuiz.Questions[0].Answers[0].Text)
		}

		if testQuiz.Questions[0].Answers[0].IsCorrect != upsertedQuiz.Questions[0].Answers[0].IsCorrect {
			t.Errorf("Expected Answer IsCorrect %v, got: %v", testQuiz.Questions[0].Answers[0].IsCorrect, upsertedQuiz.Questions[0].Answers[0].IsCorrect)
		}
	})

	t.Run("DeleteQuiz", func(t *testing.T) {
		db := newDB()
		defer db.Close()
		repo := newRepo(db)

		quiz := &Quiz{TitleField: "Test Quiz", Password: "1234", Description: "This is a test quiz"}
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

		quiz1 := &Quiz{TitleField: "Test Quiz 1", Password: "1234", Description: "This is a test quiz"}
		quiz2 := &Quiz{TitleField: "Test Quiz 2", Password: "1234", Description: "This is a test quiz"}
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
