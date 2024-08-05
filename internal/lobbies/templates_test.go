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

	err := ChooseUsernameView.Execute(io.Discard, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWaitingRoomView(t *testing.T) {
	data := ViewData{
		Lobby: Example1234Lobby(),
		User:  &ExampleUser,
	}

	err := WaitingRoomView.Execute(io.Discard, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestQuestionView(t *testing.T) {
	data := ViewData{
		Lobby: ExampleLobbyOnReadingView(),
		User:  &ExampleUser,
	}

	err := QuestionView.Execute(io.Discard, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAnswerView(t *testing.T) {
	data := ViewData{
		Lobby: ExampleLobbyOnAnswerView(),
		User:  &ExampleUser,
	}

	err := AnswerView.Execute(io.Discard, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestAnswerViewHost(t *testing.T) {
	lobby := ExampleLobbyOnAnswerView()
	lobby.Host = &ExampleUser
	data := ViewData{
		Lobby: lobby,
		User:  &ExampleUser,
	}

	err := AnswerView.Execute(io.Discard, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestOnFinishView(t *testing.T) {
	lobby := ExampleLobbyOnAnswerView()
	if err := lobby.Finish(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	data := OnFinishData{
		PastGameID: 120,
		ViewData: ViewData{
			Lobby: ExampleLobbyOnAnswerView(),
			User:  &ExampleUser,
		},
	}

	err := onFinishView.Execute(io.Discard, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
