package quiz

import (
	"errors"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type RepositorySQLite struct {
	*repositorySQLite
}

type repositorySQLite struct {
	db *sqlx.DB
}

func NewRepositorySQLite(db *sqlx.DB) (RepositorySQLite, error) {
	repo := RepositorySQLite{
		&repositorySQLite{
			db: db,
		},
	}

	return repo, repo.createTables()
}

func (repo *repositorySQLite) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS quiz (
		quiz_id     INTEGER PRIMARY KEY,
		title       TEXT,
		password    TEXT,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS question (
		question_id   INTEGER PRIMARY KEY,
		quiz_id       INTEGER REFERENCES quiz(quiz_id) ON DELETE CASCADE,
		question_text TEXT
	);

	CREATE TABLE IF NOT EXISTS answer (
		answer_id   INTEGER PRIMARY KEY,
		question_id INTEGER REFERENCES question(question_id) ON DELETE CASCADE,
		is_correct  INTEGER,
		answer_text TEXT,
		latex       TEXT
	);
	`

	_, err := repo.db.Exec(schema)
	return err
}

func (repo *repositorySQLite) Insert(quiz *Quiz) (int64, error) {
	if quiz == nil {
		return 0, errors.New("quiz is nil")
	}

	tx, err := repo.db.Beginx()
	if err != nil {
		return 0, err
	}
	// Rollback if no tx.Commit (if there is commit, this is no-op)
	defer tx.Rollback()

	// Insert the game
	res, err := tx.NamedExec(`
        INSERT INTO quiz (title, password, description)
		VALUES (:title, :password, :description)
    `, &quiz)
	if err != nil {
		return 0, err
	}

	insertedQuizID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert all questions
	for _, question := range quiz.Questions {
		res, err := tx.Exec(`
			INSERT INTO question (quiz_id, question_text)
			VALUES (?, ?)
		`, insertedQuizID, question.Text)
		if err != nil {
			return 0, err
		}

		insertedQuestionID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}

		for i := range question.Answers {
			question.Answers[i].QuestionID = insertedQuestionID
		}

		res, err = tx.NamedExec(`
			INSERT INTO answer (question_id, is_correct, answer_text, latex)
			VALUES (:question_id, :is_correct, :answer_text, :latex)
		`, question.Answers)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()

	return insertedQuizID, err
}
func (repo *repositorySQLite) Upsert(quiz *Quiz) (int64, error) {
	if quiz == nil {
		return 0, errors.New("quiz is nil")
	}
	tx, err := repo.db.Beginx()
	if err != nil {
		return 0, err
	}
	// Rollback if no tx.Commit (if there is commit, this is no-op)
	defer tx.Rollback()

	// Insert the game
	_, err = tx.NamedExec(`
		INSERT INTO quiz (quiz_id, title, password, description)
		VALUES (:quiz_id, :title, :password, :description)
		ON CONFLICT(quiz_id) DO UPDATE SET
		title = EXCLUDED.title,
		password = EXCLUDED.password,
		description = EXCLUDED.description
	`, &quiz)
	if err != nil {
		return 0, err
	}

	// Remove all questions (and answers because of CASCADE)
	_, err = tx.NamedExec(`
		DELETE FROM question WHERE quiz_id = :quiz_id
	`, &quiz)
	if err != nil {
		return 0, err
	}

	// Insert all questions
	for _, question := range quiz.Questions {
		res, err := tx.Exec(`
			INSERT INTO question (quiz_id, question_text)
			VALUES (?, ?)
		`, quiz.ID, question.Text)
		if err != nil {
			return 0, err
		}

		insertedQuestionID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}

		for i := range question.Answers {
			question.Answers[i].QuestionID = insertedQuestionID
		}

		res, err = tx.NamedExec(`
			INSERT INTO answer (question_id, is_correct, answer_text, latex)
			VALUES (:question_id, :is_correct, :answer_text, :latex)
		`, question.Answers)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()

	return quiz.ID, err
}

func (repo *repositorySQLite) Update(quiz *Quiz) (int64, error) {
	if quiz == nil {
		return 0, errors.New("quiz is nil")
	}
	tx, err := repo.db.Beginx()
	if err != nil {
		return 0, err
	}
	// Rollback if no tx.Commit (if there is commit, this is no-op)
	defer func() {
		rErr := tx.Rollback()
		if rErr != nil {
			slog.Error("During handling an insert error with rollback, another error appeared", "err", rErr)
		}
	}()

	// Insert the game
	res, err := tx.NamedExec(`
		UPDATE quiz SET
		title = :title,
		password = :password,
		description = :description
		WHERE quiz_id = :quiz_id
	`, &quiz)
	if err != nil {
		return 0, err
	}

	upsertedQuizID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Remove all questions (and answers because of CASCADE)
	_, err = tx.Exec(`
		DELETE FROM question WHERE quiz_id = ?
	`, upsertedQuizID)
	if err != nil {
		return 0, err
	}

	// Insert all questions
	for _, question := range quiz.Questions {
		res, err := tx.Exec(`
			INSERT INTO question (quiz_id, question_text)
			VALUES (?, ?)
		`, upsertedQuizID, question.Text)
		if err != nil {
			return 0, err
		}

		insertedQuestionID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}

		for i := range question.Answers {
			question.Answers[i].QuestionID = insertedQuestionID
		}

		res, err = tx.NamedExec(`
			INSERT INTO answer (question_id, is_correct, answer_text, latex)
			VALUES (:question_id, :is_correct, :answer_text, :latex)
		`, question.Answers)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()

	return upsertedQuizID, err
}

func (repo *repositorySQLite) Get(id int64) (*Quiz, error) {
	query := "SELECT * FROM quiz WHERE quiz_id = ?"
	var quiz Quiz
	err := repo.db.Get(&quiz, query, id)
	if err != nil {
		return nil, nil
	}

	// Hydrate the quiz Questions (the answers should also be hydrated)
	query = `
		SELECT question.*, answer.*
		FROM question
		LEFT JOIN answer
		ON question.question_id = answer.question_id
		WHERE question.quiz_id = ?
		ORDER BY question.question_id, answer.answer_id
	`
	rows, err := repo.db.Queryx(query, id)
	if err != nil {
		return nil, err
	}
	// Scan join query result
	var currentQst *Question
	for rows.Next() {
		type Result struct {
			*Question
			Answer
		}
		var res Result
		err := rows.StructScan(&res)
		if err != nil {
			return nil, err
		}

		if currentQst == nil {
			currentQst = res.Question
		}

		if res.Question.id != currentQst.id {
			quiz.Questions = append(quiz.Questions, *currentQst)
			currentQst = res.Question
		}

		currentQst.Answers = append(currentQst.Answers, res.Answer)
	}
	if currentQst != nil {
		quiz.Questions = append(quiz.Questions, *currentQst)
	}

	return &quiz, err
}

func (repo *repositorySQLite) Delete(id int64) error {
	query := "DELETE FROM quiz WHERE quiz_id=?"
	_, err := repo.db.Exec(query, id)
	return err
}

// NOTE: This function will return unhydrated Quizzes
func (repo *repositorySQLite) GetAll() ([]Quiz, error) {
	query := "SELECT * FROM quiz"
	var quizzes []Quiz
	err := repo.db.Select(&quizzes, query)
	return quizzes, err
}

func (repo *repositorySQLite) GetAllQuizzesMetadata() ([]QuizMetadata, error) {
	query := "SELECT quiz_id, title FROM quiz"
	var quizzes []QuizMetadata
	err := repo.db.Select(&quizzes, query)
	return quizzes, err
}
