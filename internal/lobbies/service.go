package lobbies

import (
	"github.com/erykksc/kwikquiz/internal/pastgames"
	"github.com/erykksc/kwikquiz/internal/quiz"
)

type Service struct {
	lRepo  Repository           // Lobby Repository
	pgRepo pastgames.Repository // PastGames Repository
	qRepo  quiz.Repository      // Quizzes Repository
}

func NewService(lobbyRepo Repository, pastGamesRepo pastgames.Repository, quizRepo quiz.Repository) Service {
	return Service{
		lRepo:  lobbyRepo,
		pgRepo: pastGamesRepo,
		qRepo:  quizRepo,
	}
}
