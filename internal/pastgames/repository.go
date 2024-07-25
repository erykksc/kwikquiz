package pastgames

type ErrPastGameNotFound struct{}

func (ErrPastGameNotFound) Error() string {
	return "past game not found"
}

type PastGameRepository interface {
	Insert(game *PastGame) (int64, error)
	Upsert(game *PastGame) (int64, error)
	GetByID(id int64) (*PastGame, error)
	GetAll() ([]PastGame, error)
	HydrateScores(*PastGame) error
	BrowsePastGamesByID(query string) ([]PastGame, error)
}
