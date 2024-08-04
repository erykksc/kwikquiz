package lobbies

import (
	"io"
	"testing"
)

var viewData = ViewData{
	Lobby: &Lobby{},
	User:  &User{},
}

func TestChooseUsernameView(t *testing.T) {
	w := io.Discard
	err := ChooseUsernameView.Execute(w, viewData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWaitingRoomView(t *testing.T) {
	w := io.Discard
	err := WaitingRoomView.Execute(w, viewData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestQuestionView(t *testing.T) {
	w := io.Discard
	err := QuestionView.Execute(w, viewData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAnswerView(t *testing.T) {
	w := io.Discard
	err := AnswerView.Execute(w, viewData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestOnFinishView(t *testing.T) {
	data := OnFinishData{
		PastGameID: 120,
		ViewData:   viewData,
	}
	w := io.Discard
	err := onFinishView.Execute(w, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
