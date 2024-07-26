package quiz

type Quiz struct {
	ID          int64 `db:"quiz_id"`
	Title       string
	Password    string
	Description string
	Questions   []Question `gorm:"foreignKey:QuizID"`
}

type Question struct {
	id      int64    `db:"question_id"`
	QuizID  int64    `db:"quiz_id"`
	Text    string   `db:"question_text"`
	Answers []Answer `gorm:"foreignKey:QuestionID"`
}

type Answer struct {
	ID         int64  `db:"answer_id"`
	QuestionID int64  `db:"question_id"`
	IsCorrect  bool   `db:"is_correct"`
	Text       string `db:"answer_text"`
	LaTeX      string `db:"latex"`
	ImageName  string
	Image      []byte
}

// It is used for faster lookups if only limited data is needed
type QuizMetadata struct {
	ID    uint `db:"quiz_id"`
	Title string
}
