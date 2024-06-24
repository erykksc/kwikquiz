package quiz

import (
	"errors"
	"github.com/erykksc/kwikquiz/internal/models"
	"gorm.io/gorm"
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
	result := r.DB.Create(&q)
	if result.Error != nil {
		return 0, result.Error
	}
	return q.ID, nil
}

func (r *GormQuizRepository) UpdateQuiz(q models.Quiz) (uint, error) {
	result := r.DB.Save(&q)
	if result.Error != nil {
		return 0, result.Error
	}
	return q.ID, nil
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
