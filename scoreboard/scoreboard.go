package scoreboard

import (
	"github.com/mowings/imago/work"
	"log"
	"time"
)

const TIMEOUT = 30

type LongPollChanMessage struct {
	Id           string
	LongPollChan *chan work.Work
}

type WorkStatusRequest struct {
	Id       string
	Chan     chan work.Work
	LongPoll bool
}

type Scoreboard struct {
	updateWorkChannel   chan work.Work
	GetWorkChannel      chan WorkStatusRequest
	LongPollChanRequest chan LongPollChanMessage
}

func New() (scoreboard *Scoreboard) {
	scoreboard = new(Scoreboard)
	scoreboard.updateWorkChannel = make(chan work.Work)
	scoreboard.GetWorkChannel = make(chan WorkStatusRequest)
	go scoreboard.worker()
	return scoreboard
}

type WorkEntry struct {
	Work         *work.Work
	UpdatedAt    time.Time
	LongPollChan *chan work.Work
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

	for {
		select {
		case w := <-scoreboard.updateWorkChannel:
			if mapentry, ok := workerMap[w.Id]; ok {
				mapentry.touch()
				// We should only get here on status changes, but just to be sure, be sure status us changing
				old_status := mapentry.Work.Status
				mapentry.Work = &w
				if old_status != mapentry.Work.Status {
					if mapentry.Work.Status == "error" || mapentry.Work.Status == "done" {
						log.Println("Reporting work completion on long poll channel for ", w.Id)
						*mapentry.LongPollChan <- *mapentry.Work
					}
				}

			} else {
				c := make(chan work.Work, 1)
				workerMap[w.Id] = &(WorkEntry{&w, time.Now(), &c})
			}
			log.Println("Updated worker", w.Id, w.Status)
		case work_check := <-scoreboard.GetWorkChannel:
			var work_data work.Work
			if workerMap[work_check.Id] != nil {
				work_data = *(workerMap[work_check.Id].Work)
			}
			work_check.Chan <- work_data

		case <-time.After(time.Second * TIMEOUT):
		}
	}
}
