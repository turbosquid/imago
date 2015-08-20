package work

import (
	"github.com/twinj/uuid"
)

type Action struct {
	Status     string   `json:"status"`
	Credential string   `json:"credential"`
	Infile     string   `json:"infile"`
	Outfile    string   `json:"outfile"`
	Mimetype   string   `json:"mimetype"`
	Operations []string `json:"operations"`
	Output     string   `json:"output"`
	Error      string   `json:"error"`
}

type Work struct {
	Id      string   `json:"id"`
	Status  string   `json:"status"`
	Actions []Action `json:"actions"`
}

func (w *Work) Initialize() {
	u := uuid.NewV4()
	uuid.SwitchFormat(uuid.Clean)
	w.Id = u.String()
	w.Status = "queued"
	for i, a := range w.Actions {
		if a.Credential == "" {
			w.Actions[i].Credential = "default"
		}
	}
}

func (w *Work) IsComplete() bool {
	return (w.Status == "done" || w.Status == "error")
}
