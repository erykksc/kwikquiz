package common

type HX_Headers struct {
	HxCurrentURL  string `json:"HX-Current-URL"`
	HxRequest     string `json:"HX-Request"`
	HxTarget      string `json:"HX-Target"`
	HxTrigger     string `json:"HX-Trigger"`
	HxTriggerName string `json:"HX-Trigger-Name"`
}
