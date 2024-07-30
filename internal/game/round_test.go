package game

import (
	"testing"
	"time"
)

// Mock MyQuestion struct
type MyQuestion struct{}

func (q MyQuestion) IsAnswerCorrect(index int) bool {
	return index == 1 // Mock implementation
}

func (q MyQuestion) IsAnswerValid(index int) bool {
	return index > -1
}

func TestCreateRound(t *testing.T) {
	players := []Username{"Alice", "Bob"}
	question := MyQuestion{}
	settings := RoundSettings{
		ReadingTime: 5 * time.Second,
		AnswerTime:  10 * time.Second,
	}

	round := CreateRound(players, question, settings)

	if len(round.players) != len(players) {
		t.Errorf("Expected %d players, got %d", len(players), len(round.players))
	}

	for _, player := range players {
		if _, exists := round.players[player]; !exists {
			t.Errorf("Player %s not found in round players", player)
		}
	}
}

func TestStart(t *testing.T) {
	players := []Username{"Alice", "Bob"}
	question := MyQuestion{}
	settings := RoundSettings{
		ReadingTime: 2 * time.Second,
		AnswerTime:  3 * time.Second,
	}

	round := CreateRound(players, question, settings)

	if err := round.Start(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := round.Start(); err == nil {
		t.Errorf("Expected error for starting already started round, got nil")
	}
}

func TestFinishEarly(t *testing.T) {
	players := []Username{"Alice", "Bob"}
	question := MyQuestion{}
	settings := RoundSettings{
		ReadingTime: 5 * time.Second,
		AnswerTime:  10 * time.Second,
	}

	round := CreateRound(players, question, settings)
	if err := round.FinishEarly(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test finishing an already finished round
	if err := round.FinishEarly(); err == nil {
		t.Errorf("Expected error for finishing already finished round, got nil")
	}
}

func TestSubmitAnswer(t *testing.T) {
	players := []Username{"Alice", "Bob"}
	question := MyQuestion{}
	settings := RoundSettings{
		ReadingTime: 500 * time.Millisecond,
		AnswerTime:  2 * time.Second,
	}

	round := CreateRound(players, question, settings)
	err := round.Start()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := round.SubmitAnswer("Alice", 1); err == nil {
		t.Errorf("Expected error for submitting answer during reading time, got nil")
	}

	time.Sleep(settings.ReadingTime)

	if err := round.SubmitAnswer("Alice", 1); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := round.SubmitAnswer("Alice", 2); err == nil {
		t.Errorf("Expected error for already submitted answer, got nil")
	}

	if err := round.SubmitAnswer("Unknown", 1); err == nil {
		t.Errorf("Expected error for unknown player, got nil")
	}
}

func TestGetResults(t *testing.T) {
	players := []Username{"Alice", "Bob"}
	question := MyQuestion{}
	settings := RoundSettings{
		ReadingTime: 0,
		AnswerTime:  100 * time.Millisecond,
	}

	round := CreateRound(players, question, settings)
	err := round.Start()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = round.SubmitAnswer("Alice", 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_, err = round.GetResults()
	if err == nil {
		t.Errorf("Expected error for getting results before round ends, got nil")
	}

	time.Sleep(settings.ReadingTime + settings.AnswerTime + 100*time.Millisecond)

	results, err := round.GetResults()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != len(players) {
		t.Errorf("Expected %d results, got %d", len(players), len(results))
	}

	if results["Alice"] == 0 {
		t.Errorf("Expected Alice to have points, got 0")
	}
}

func TestFinished(t *testing.T) {
	players := []Username{"Alice", "Bob"}
	question := MyQuestion{}
	settings := RoundSettings{
		ReadingTime: 100 * time.Millisecond,
		AnswerTime:  200 * time.Millisecond,
	}

	round := CreateRound(players, question, settings)

	startTime := time.Now()
	err := round.Start()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	<-round.Finished()

	if time.Since(startTime).Round(100*time.Millisecond) != settings.ReadingTime+settings.AnswerTime {
		t.Errorf("Expected round to finish after ReadingTime + AnswerTime, got %v", time.Since(startTime))
	}
}
