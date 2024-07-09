package models

import (
	"gorm.io/gorm"
	"time"
)

type PlayerScore struct {
	gorm.Model
	PastGameID uint
	Username   string
	Score      int
}

type PastGame struct {
	gorm.Model
	ID        uint
	StartedAt time.Time
	EndedAt   time.Time
	QuizTitle string
	Scores    []PlayerScore `gorm:"foreignKey:PastGameID"` // sorted by score, descending
}
