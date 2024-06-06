package lobbies

import "sync"

type errLobbyNotFound struct{}

func (errLobbyNotFound) Error() string {
	return "lobby not found"
}

type errLobbyAlreadyExists struct{}

func (errLobbyAlreadyExists) Error() string {
	return "game already exists"
}

type lobbyRepository interface {
	AddLobby(*lobby) error
	UpdateLobby(*lobby) error
	GetLobby(pin string) (*lobby, error)
	DeleteLobby(pin string) error
	GetAllLobbies() ([]*lobby, error)
}

// In-memory store for games
type inMemoryLobbyRepository struct {
	lobbies map[string]*lobby
	mu      sync.RWMutex
}

func newInMemoryLobbyRepository() *inMemoryLobbyRepository {
	return &inMemoryLobbyRepository{
		lobbies: make(map[string]*lobby),
	}
}

func (s *inMemoryLobbyRepository) AddLobby(l *lobby) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.lobbies[l.Pin]; ok {
		return errLobbyAlreadyExists{}
	}

	s.lobbies[l.Pin] = l
	return nil
}

func (s *inMemoryLobbyRepository) UpdateLobby(l *lobby) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.lobbies[l.Pin]; !ok {
		return errLobbyNotFound{}
	}

	s.lobbies[l.Pin] = l
	return nil
}

func (s *inMemoryLobbyRepository) GetLobby(pin string) (*lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.lobbies[pin]
	if !ok {
		return &lobby{}, errLobbyNotFound{}
	}

	return l, nil
}

func (s *inMemoryLobbyRepository) DeleteLobby(pin string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.lobbies[pin]; !ok {
		return errLobbyNotFound{}
	}

	delete(s.lobbies, pin)
	return nil
}

func (s *inMemoryLobbyRepository) GetAllLobbies() ([]*lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var lobbies []*lobby
	for _, l := range s.lobbies {
		lobbies = append(lobbies, l)
	}

	return lobbies, nil
}
