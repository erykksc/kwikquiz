package lobbies

import (
	"fmt"
	"math/rand"
	"sync"
)

type errLobbyNotFound struct{}

func (errLobbyNotFound) Error() string {
	return "lobby not found"
}

type errLobbyAlreadyExists struct{}

func (errLobbyAlreadyExists) Error() string {
	return "game already exists"
}

type Repository interface {
	AddLobby(*Lobby) error
	GetLobby(pin string) (*Lobby, error)
	DeleteLobby(pin string) error
	GetAllLobbies() ([]*Lobby, error)
	GetLobbyByHost(ClientID) (*Lobby, error)
}

// In-memory store for games
type inMemoryLobbyRepository struct {
	lobbies map[string]*Lobby
	mu      sync.RWMutex
}

func newInMemoryLobbyRepository() *inMemoryLobbyRepository {
	return &inMemoryLobbyRepository{
		lobbies: make(map[string]*Lobby),
	}
}

// AddLobby adds a new lobby to the in-memory store
// If the lobby has a pin it tries to add it to the store
// If the lobby doesn't have a pin, it updates the Pin field with a new pin
func (s *inMemoryLobbyRepository) AddLobby(l *Lobby) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	l.mu.Lock()
	defer l.mu.Unlock()

	// If the lobby doesn't have a pin, create one
	if l.Pin == "" {
		for l.Pin == "" || s.lobbies[l.Pin] != nil {
			// Generate a new pin, a random 4 digit number
			newPin := rand.Intn(10000)
			l.Pin = fmt.Sprintf("%04d", newPin)
		}
	}

	if _, ok := s.lobbies[l.Pin]; ok {
		return errLobbyAlreadyExists{}
	}

	s.lobbies[l.Pin] = l
	return nil
}

func (s *inMemoryLobbyRepository) GetLobby(pin string) (*Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.lobbies[pin]
	if !ok {
		return &Lobby{}, errLobbyNotFound{}
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

func (s *inMemoryLobbyRepository) GetAllLobbies() ([]*Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var lobbies []*Lobby
	for _, l := range s.lobbies {
		lobbies = append(lobbies, l)
	}

	return lobbies, nil
}

func (s *inMemoryLobbyRepository) GetLobbyByHost(host ClientID) (*Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, l := range s.lobbies {
		if l.Host != nil && l.Host.ClientID == host {
			return l, nil
		}
	}

	return nil, errLobbyNotFound{}
}
