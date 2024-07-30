package lobbies

import (
	"time"

	"github.com/erykksc/kwikquiz/internal/quiz"
)

func GetExamples() []*Lobby {
	return []*Lobby{
		Example1234Lobby(),
		Example1235Lobby(),
		ExampleLobbyOnReadingView(),
		ExampleLobbyOnAnswerView(),
	}
}

func Example1234Lobby() *Lobby {
	options := NewLobbyOptions()
	options.Pin = "1234"
	options.Quiz = quiz.ExampleQuizGeography
	lobby := createLobby(options)
	return lobby
}

// &Example1234Lobby,

func Example1235Lobby() *Lobby {
	options := NewLobbyOptions()
	options.Pin = "1235"
	options.Quiz = quiz.ExampleQuizGeography
	lobby := createLobby(options)

	err := lobby.AddPlayer("Jack")
	if err != nil {
		panic(err)
	}
	return lobby
}

func ExampleLobbyOnReadingView() *Lobby {
	options := NewLobbyOptions()
	options.Pin = "2345"
	options.Quiz = quiz.ExampleQuizGeography
	options.ReadingTime = 999 * time.Second
	lobby := createLobby(options)

	err := lobby.AddPlayer("Jack")
	if err != nil {
		panic(err)
	}

	err = lobby.Start()
	if err != nil {
		panic(err)
	}

	return lobby
}

func ExampleLobbyOnAnswerView() *Lobby {
	options := NewLobbyOptions()
	options.Pin = "2346"
	options.Quiz = quiz.ExampleQuizGeography
	options.ReadingTime = 0
	options.AnswerTime = 999 * time.Second
	lobby := createLobby(options)

	err := lobby.AddPlayer("Jack")
	if err != nil {
		panic(err)
	}

	err = lobby.Start()
	if err != nil {
		panic(err)
	}

	return lobby
}
