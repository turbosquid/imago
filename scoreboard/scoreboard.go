package scoreboard

import (
	"github.com/mowings/imago/work"
	"log"
	"time"
)

const TIMEOUT = 30

type Scoreboard struct {
	updateWorkChannel chan work.Work
	GetWorkChannel    chan work.Work
}

func New() (scoreboard *Scoreboard) {
	scoreboard = new(Scoreboard)
	scoreboard.updateWorkChannel = make(chan work.Work)
	scoreboard.GetWorkChannel = make(chan work.Work)
	go scoreboard.worker()
	return scoreboard
}

type WorkEntry struct {
	Work      *work.Work
	UpdatedAt time.Time
}

func (w *WorkEntry) touch() {
	w.UpdatedAt = time.Now()
}

// Call to update a work. Makes a copy of the work before putting it on the channel
func (scoreboard *Scoreboard) UpdateWork(work *work.Work) {
	work_copy := *work
	scoreboard.updateWorkChannel <- work_copy
}

// Private worker
func (scoreboard *Scoreboard) worker() {
	workerMap := make(map[string]*WorkEntry)
	var w work.Work

	// Set up timeout goroutine
	timeout := make(chan bool, 1)
	go func() {
		for {
			time.Sleep(TIMEOUT * time.Second)
			timeout <- true
		}
	}()

	for {
		select {
		case w = <-scoreboard.updateWorkChannel:
			workerMap[w.Id] = &WorkEntry{&w, time.Now()}
			log.Println("Updated worker", w.Id, w.Status)
		case work_check := <-scoreboard.GetWorkChannel:
			if workerMap[work_check.Id] != nil {
				work_check = *(workerMap[work_check.Id].Work)
			} else {
				work_check.Status = ""
			}
			scoreboard.GetWorkChannel <- work_check
		case <-timeout:
		}
	}
}
