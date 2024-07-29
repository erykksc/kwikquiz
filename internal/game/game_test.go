package game

import (
	"errors"
	"testing"
	"time"
)

// Mock Quiz implementation
type MockQuiz struct {
	questions []Question
}

func (q MockQuiz) Title() string {
	return "MockQuiz"
}

func (q MockQuiz) GetQuestion(idx int) (Question, error) {
	if idx >= len(q.questions) {
		return nil, errors.New("question not found")
	}
	return q.questions[idx], nil
}

func (q MockQuiz) QuestionsCount() int {
	return len(q.questions)
}

func createMockGame() Game {
	settings := GameSettings{
		Quiz: MockQuiz{
			questions: []Question{
				MyQuestion{},
				MyQuestion{},
			},
		},
		RoundSettings: RoundSettings{
			ReadingTime: 1 * time.Second,
			AnswerTime:  1 * time.Second,
		},
	}
	return CreateGame(settings)
}

// Contains checks if a slice contains a specific element
func Contains(slice []Username, item Username) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func TestCreateGame(t *testing.T) {
	game := createMockGame()
	if game.IsFinished() {
		t.Errorf("New game should not be finished")
	}
}

func TestAddPlayer(t *testing.T) {
	game := createMockGame()

	if err := game.AddPlayer("Alice"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := game.AddPlayer("Alice"); err == nil {
		t.Errorf("Expected error for adding existing player, got nil")
	}

	game.Start()
	if err := game.AddPlayer("Bob"); err == nil {
		t.Errorf("Expected error for adding player after game started, got nil")
	}
}

func TestChangeUsername(t *testing.T) {
	game := createMockGame()
	game.AddPlayer("Alice")

	if !game.PlayerInGame("Alice") {
		t.Errorf("Alice should be in players")
	}

	if err := game.ChangeUsername("Alice", "Bob"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if game.PlayerInGame("Alice") {
		t.Errorf("Alice should not be in players")
	}

	if !game.PlayerInGame("Bob") {
		t.Errorf("Bob should be in players")
	}

	if err := game.ChangeUsername("Charlie", "Dave"); err == nil {
		t.Errorf("Expected error for changing non-existent username, got nil")
	}

}

func TestRemovePlayer(t *testing.T) {
	game := createMockGame()
	game.AddPlayer("Alice")

	if err := game.RemovePlayer("Alice"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if Contains(game.Players(), "Alice") {
		t.Errorf("Alice should not be in players")
	}

	if err := game.RemovePlayer("Alice"); err == nil {
		t.Errorf("Expected error for removing non-existent player, got nil")
	}

	game.AddPlayer("Bob")
	game.Start()
	if err := game.RemovePlayer("Bob"); err == nil {
		t.Errorf("Expected error for removing player after game started, got nil")
	}
}

func TestStartGame(t *testing.T) {
	game := createMockGame()

	err := game.Start()
	if err == nil {
		t.Errorf("Expected error because of 0 players, got nil")
	}

	err = game.AddPlayer("Jack")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = game.Start()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = game.Start()
	if err == nil {
		t.Errorf("Expected error for starting already started game, got nil")
	}
}

func TestStartNextRound(t *testing.T) {
	game := createMockGame()

	err := game.AddPlayer("Jack")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = game.Start()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = game.FinishRoundEarly()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = game.StartNextRound()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = game.FinishRoundEarly()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = game.StartNextRound()
	if err != (ErrNoMoreQuestions{}) {
		t.Errorf("Expected no more questions error, got: %v", err)
	}
}

func TestFinishGame(t *testing.T) {
	game := createMockGame()
	game.Start()

	if err := game.Finish(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := game.Finish(); err == nil {
		t.Errorf("Expected error for finishing already finished game, got nil")
	}
}

func TestIsFinished(t *testing.T) {
	game := createMockGame()
	game.Start()
	game.Finish()

	if !game.IsFinished() {
		t.Errorf("Game should be finished")
	}
}

func TestGetPlayers(t *testing.T) {
	game := createMockGame()
	game.AddPlayer("Alice")
	game.AddPlayer("Bob")

	players := game.Players()
	if len(players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(players))
	}

	expectedPlayers := map[Username]bool{
		"Alice": true,
		"Bob":   true,
	}

	for _, player := range players {
		if !expectedPlayers[player] {
			t.Errorf("Unexpected player: %s", player)
		}
	}
}
