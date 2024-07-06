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
	result := r.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&q)
	if result.Error != nil {
		return 0, result.Error
	}
	return q.ID, nil
}

//func (r *GormQuizRepository) UpdateQuiz(q models.Quiz) (uint, error) {
//	result := r.DB.Save(&q)
//	if result.Error != nil {
//		return 0, result.Error
//	}
//	return q.ID, nil
//}

func (r *GormQuizRepository) UpdateQuiz(q models.Quiz) (uint, error) {
	// Start a transaction
	tx := r.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update the quiz
	if err := tx.Model(&q).Updates(models.Quiz{
		Title:       q.Title,
		Password:    q.Password,
		Description: q.Description,
	}).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// Get existing questions
	var existingQuestions []models.Question
	if err := tx.Where("quiz_id = ?", q.ID).Find(&existingQuestions).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// Update or create questions
	for _, question := range q.Questions {
		var existingQuestion models.Question
		if err := tx.Where("id = ? AND quiz_id = ?", question.ID, q.ID).First(&existingQuestion).Error; err == nil {
			// Update existing question
			if err := tx.Model(&existingQuestion).Updates(models.Question{
				Text: question.Text,
			}).Error; err != nil {
				tx.Rollback()
				return 0, err
			}

			// Update or create answers
			if err := updateAnswers(tx, existingQuestion.ID, question.Answers); err != nil {
				tx.Rollback()
				return 0, err
			}
		} else if err == gorm.ErrRecordNotFound {
			// Create new question
			newQuestion := models.Question{
				QuizID: q.ID,
				Text:   question.Text,
			}
			if err := tx.Create(&newQuestion).Error; err != nil {
				tx.Rollback()
				return 0, err
			}

			// Create new answers
			if err := updateAnswers(tx, newQuestion.ID, question.Answers); err != nil {
				tx.Rollback()
				return 0, err
			}
		} else {
			tx.Rollback()
			return 0, err
		}
	}

	// Delete questions that are no longer present
	for _, existingQuestion := range existingQuestions {
		found := false
		for _, question := range q.Questions {
			if existingQuestion.ID == question.ID {
				found = true
				break
			}
		}
		if !found {
			if err := tx.Delete(&existingQuestion).Error; err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return q.ID, nil
}

func updateAnswers(tx *gorm.DB, questionID uint, answers []models.Answer) error {
	// Get existing answers
	var existingAnswers []models.Answer
	if err := tx.Where("question_id = ?", questionID).Find(&existingAnswers).Error; err != nil {
		return err
	}

	// Update or create answers
	for _, answer := range answers {
		var existingAnswer models.Answer
		if err := tx.Where("id = ? AND question_id = ?", answer.ID, questionID).First(&existingAnswer).Error; err == nil {
			// Update existing answer
			if err := tx.Model(&existingAnswer).Updates(models.Answer{
				IsCorrect: answer.IsCorrect,
				Text:      answer.Text,
				LaTeX:     answer.LaTeX,
				ImageName: answer.ImageName,
				Image:     answer.Image,
			}).Error; err != nil {
				return err
			}
		} else if err == gorm.ErrRecordNotFound {
			// Create new answer
			newAnswer := models.Answer{
				QuestionID: questionID,
				IsCorrect:  answer.IsCorrect,
				Text:       answer.Text,
				LaTeX:      answer.LaTeX,
				ImageName:  answer.ImageName,
				Image:      answer.Image,
			}
			if err := tx.Create(&newAnswer).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Delete answers that are no longer present
	for _, existingAnswer := range existingAnswers {
		found := false
		for _, answer := range answers {
			if existingAnswer.ID == answer.ID {
				found = true
				break
			}
		}
		if !found {
			if err := tx.Delete(&existingAnswer).Error; err != nil {
				return err
			}
		}
	}

	return nil
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
