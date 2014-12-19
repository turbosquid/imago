package work

import (
	"github.com/twinj/uuid"
)

type Action struct {
	Status     string
	Infile     string
	Outfile    string
	Mimetype   string
	Operations []string
	Output     string
	Error      string
}

type Work struct {
	Id      string
	Status  string
	Actions []Action
}

func (w *Work) Initialize() {
	u := uuid.NewV4()
	uuid.SwitchFormat(uuid.Clean)
	w.Id = u.String()
	w.Status = "queued"
}

func (w *Work) IsComplete() bool {
	return (w.Status == "done" || w.Status == "error")
}
