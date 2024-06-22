package models

import (
	"gorm.io/gorm"
)

type QuizModel struct {
	gorm.Model
	ID            int
	Title         string
	Password      string
	Description   string
	QuestionOrder string
	Questions     []QuestionModel `gorm:"foreignKey:ID"`
}

type QuestionModel struct {
	gorm.Model
	Text          string
	Answers       []AnswerModel `gorm:"foreignKey:Text"`
	CorrectAnswer int
}

type AnswerModel struct {
	gorm.Model
	IsCorrect bool
	Text      string
}
