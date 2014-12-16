package scoreboard

import (
	"github.com/mowings/imago/work"
	"log"
	"time"
)

var updateWorkChannel = make(chan *work.Work)
var GetWorkChannel = make(chan work.Work)

const TIMEOUT = 30

type WorkEntry struct {
	Work      *work.Work
	UpdatedAt time.Time
}

func (w *WorkEntry) touch() {
	w.UpdatedAt = time.Now()
}

// Call to update a work. Makes a copy of the work before putting it on the channel
func UpdateWork(work *work.Work) {
	work_copy := *work
	updateWorkChannel <- &work_copy
}

// Private worker
func worker() {
	workerMap := make(map[string]*WorkEntry)
	var w *work.Work

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
		case w = <-updateWorkChannel:
			workerMap[w.Id] = &WorkEntry{w, time.Now()}
			log.Println("Updated worker", w.Id, w.Status)
		case work_check := <-GetWorkChannel:
			if workerMap[work_check.Id] != nil {
				work_check = *(workerMap[work_check.Id].Work)
			} else {
				work_check.Status = ""
			}
			GetWorkChannel <- work_check
		case <-timeout:
		}
	}
}

func Start() {
	go worker()
}
