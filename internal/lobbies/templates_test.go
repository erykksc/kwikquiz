package lobbies

import (
	"io"
	"testing"
)

func TestChooseUsernameView(t *testing.T) {
	data := ViewData{
		Lobby: Example1234Lobby(),
		User:  &ExampleUser,
	}
	w := io.Discard

	err := ChooseUsernameView.Execute(w, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWaitingRoomView(t *testing.T) {
	data := ViewData{
		Lobby: Example1234Lobby(),
		User:  &ExampleUser,
	}
	w := io.Discard

	err := WaitingRoomView.Execute(w, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestQuestionView(t *testing.T) {
	data := ViewData{
		Lobby: ExampleLobbyOnReadingView(),
		User:  &ExampleUser,
	}

	w := io.Discard
	err := QuestionView.Execute(w, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAnswerView(t *testing.T) {
	data := ViewData{
		Lobby: ExampleLobbyOnAnswerView(),
		User:  &ExampleUser,
	}
	w := io.Discard

	err := AnswerView.Execute(w, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestOnFinishView(t *testing.T) {
	data := OnFinishData{
		PastGameID: 120,
		ViewData: ViewData{
			Lobby: ExampleLobbyOnAnswerView(),
			User:  &ExampleUser,
		},
	}

	w := io.Discard
	err := onFinishView.Execute(w, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
