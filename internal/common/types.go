package common

// TODO: This should be moved to a results/pastgames package
// type Game struct {
// 	StartedAt time.Time
// 	EndedAt   time.Time
// 	Quiz      *quiz.Quiz
// 	// Username -> Points
// 	Points map[string]int
// }

type HX_Headers struct {
	HxCurrentURL  string `json:"HX-Current-URL"`
	HxRequest     string `json:"HX-Request"`
	HxTarget      string `json:"HX-Target"`
	HxTrigger     string `json:"HX-Trigger"`
	HxTriggerName string `json:"HX-Trigger-Name"`
}
