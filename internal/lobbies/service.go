package lobbies

import "github.com/erykksc/kwikquiz/internal/pastgames"

type Service struct {
	// Lobby Repository
	lRepo Repository
	// PastGames Repository
	pgRepo pastgames.Repository
}

func NewService(lobbyRepo Repository, pastGamesRepo pastgames.Repository) Service {
	return Service{
		lRepo:  lobbyRepo,
		pgRepo: pastGamesRepo,
	}
}
