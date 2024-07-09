package quiz

import (
	"errors"
	"github.com/erykksc/kwikquiz/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ErrQuizNotFound struct{}

func (ErrQuizNotFound) Error() string { return "Quiz not found" }

type ErrQuizAlreadyExists struct{}

func (ErrQuizAlreadyExists) Error() string { return "Quiz already exists" }

type QuizRepository interface {
	AddQuiz(models.Quiz) (uint, error)
	UpdateQuiz(models.Quiz) (uint, error)
	GetQuiz(id uint) (models.Quiz, error)
	DeleteQuiz(id uint) error
	GetAllQuizzes() ([]models.Quiz, error)
	GetAllQuizzesMetadata() ([]models.QuizMetadata, error)
}

// DB for quizzes
type GormQuizRepository struct {
	DB *gorm.DB
}

func NewGormQuizRepository(db *gorm.DB) *GormQuizRepository {
	return &GormQuizRepository{DB: db}
}

func (r *GormQuizRepository) AddQuiz(q models.Quiz) (uint, error) {
	var quizID uint
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// Attempt to find an existing quiz by title
		var existingQuiz models.Quiz
		if err := tx.Where("title = ?", q.Title).First(&existingQuiz).Error; err == nil {
			// Quiz exists, return without making changes
			return nil
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			// Some other error occurred
			return err
		}

		// Create the quiz, using OnConflict to handle any conflicts
		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&q).Error; err != nil {
			return err
		}
		quizID = q.ID

		// Add or update questions and answers with OnConflict
		for i := range q.Questions {
			q.Questions[i].QuizID = quizID
			if err := tx.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&q.Questions[i]).Error; err != nil {
				return err
			}
			for j := range q.Questions[i].Answers {
				q.Questions[i].Answers[j].QuestionID = q.Questions[i].ID
				if err := tx.Clauses(clause.OnConflict{
					UpdateAll: true,
				}).Create(&q.Questions[i].Answers[j]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return quizID, nil
}

func (r *GormQuizRepository) UpdateQuiz(q models.Quiz) (uint, error) {
	var quizID uint
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch existing quiz
		var existingQuiz models.Quiz
		if err := tx.First(&existingQuiz, q.ID).Error; err != nil {
			return err
		}

		// Update quiz fields
		existingQuiz.Title = q.Title
		existingQuiz.Password = q.Password
		existingQuiz.Description = q.Description

		// Delete existing questions and answers
		if err := tx.Where("quiz_id = ?", q.ID).Delete(&models.Question{}).Error; err != nil {
			return err
		}

		// Create new questions and answers
		existingQuiz.Questions = q.Questions
		for i := range existingQuiz.Questions {
			existingQuiz.Questions[i].QuizID = existingQuiz.ID
			for j := range existingQuiz.Questions[i].Answers {
				existingQuiz.Questions[i].Answers[j].QuestionID = existingQuiz.Questions[i].ID
			}
		}

		// Save the updated quiz with new questions and answers
		if err := tx.Save(&existingQuiz).Error; err != nil {
			return err
		}

		quizID = existingQuiz.ID
		return nil
	})

	if err != nil {
		return 0, err
	}

	return quizID, nil
}

func (r *GormQuizRepository) GetQuiz(id uint) (models.Quiz, error) {
	var quiz models.Quiz
	result := r.DB.Preload("Questions.Answers").First(&quiz, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Quiz{}, ErrQuizNotFound{}
		}
		return models.Quiz{}, result.Error
	}
	return quiz, nil
}

func (r *GormQuizRepository) DeleteQuiz(id uint) error {
	result := r.DB.Delete(&models.Quiz{}, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrQuizNotFound{}
		}
		return result.Error
	}
	return nil
}

func (r *GormQuizRepository) GetAllQuizzes() ([]models.Quiz, error) {
	var quizzes []models.Quiz
	result := r.DB.Preload("Questions.Answers").Find(&quizzes)
	if result.Error != nil {
		return nil, result.Error
	}
	return quizzes, nil
}

func (r *GormQuizRepository) GetAllQuizzesMetadata() ([]models.QuizMetadata, error) {
	var quizzes []models.Quiz
	result := r.DB.Find(&quizzes)
	if result.Error != nil {
		return nil, result.Error
	}

	var metadata []models.QuizMetadata
	for _, quiz := range quizzes {
		metadata = append(metadata, models.QuizMetadata{
			ID:    quiz.ID,
			Title: quiz.Title,
		})
	}
	return metadata, nil
}
