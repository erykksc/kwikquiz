package models

import "gorm.io/gorm"

type Quiz struct {
	gorm.Model
	ID          int
	Title       string
	Password    string
	Description string
	Questions   []Question `gorm:"foreignKey:QuizID"`
}

type Question struct {
	gorm.Model
	Text          string
	Answers       []Answer `gorm:"foreignKey:QuestionID"`
	CorrectAnswer int
}

type Answer struct {
	gorm.Model
	IsCorrect bool
	Text      string
	// later we can add img, video etc. to allow multimodal questions
}
