package scoreboard

import (
	"github.com/turbosquid/imago/work"
	"log"
	"os"
	"path/filepath"
	"time"
)

const TIMEOUT = 30
const LIFETIME = 300

// Public work status request type.
type WorkStatusRequest struct {
	Id       string         // Id of interest
	Chan     chan work.Work // Response channel
	LongPoll bool           // If false return status immediately, else wait until job is complete
}

// Private scoreboard entry
type workEntry struct {
	Work         *work.Work
	UpdatedAt    time.Time
	LongPollChan *chan work.Work
}

func (w *workEntry) touch() {
	w.UpdatedAt = time.Now()
}

// Scoreboard
type Scoreboard struct {
	updateWorkChannel chan work.Work
	GetWorkChannel    chan WorkStatusRequest
	workDir           string
}

// Scoreboard Ctor
func New(workDir string) (scoreboard *Scoreboard) {
	scoreboard = new(Scoreboard)
	scoreboard.updateWorkChannel = make(chan work.Work)
	scoreboard.GetWorkChannel = make(chan WorkStatusRequest)
	scoreboard.workDir = workDir
	go scoreboard.worker()
	return scoreboard
}

// Call to update a work. Makes a copy of the work before putting it on the channel
func (scoreboard *Scoreboard) UpdateWork(work *work.Work) {
	work_copy := *work
	scoreboard.updateWorkChannel <- work_copy
}

// respond with work info on channel. Do not block if channel isn;t ready
func notifyWorkStatus(c chan work.Work, w work.Work) {
	select {
	case c <- w:
	default:
	}
}

// Private scoreboard worker
func (scoreboard *Scoreboard) worker() {
	workerMap := make(map[string]*workEntry)

	for {
		select {
		case w := <-scoreboard.updateWorkChannel: // Update work status
			if mapentry, ok := workerMap[w.Id]; ok {
				mapentry.touch()
				// We should only get here on status changes, but just to be sure, be sure status us changing
				old_status := mapentry.Work.Status
				mapentry.Work = &w
				if old_status != mapentry.Work.Status && mapentry.Work.IsComplete() && mapentry.LongPollChan != nil {
					go notifyWorkStatus(*mapentry.LongPollChan, *mapentry.Work)
					mapentry.LongPollChan = nil
				}

			} else {
				workerMap[w.Id] = &(workEntry{&w, time.Now(), nil})
			}
		case work_check := <-scoreboard.GetWorkChannel: /// Request work status
			var work_data work.Work
			if workerMap[work_check.Id] != nil {
				work_data = *(workerMap[work_check.Id].Work)
				if work_check.LongPoll && !work_data.IsComplete() {
					workerMap[work_check.Id].LongPollChan = &work_check.Chan
				} else {
					go notifyWorkStatus(work_check.Chan, work_data)
				}
			} else {
				go notifyWorkStatus(work_check.Chan, work_data)
			}

		case <-time.After(time.Second * TIMEOUT):
			t := time.Now()
			for key, value := range workerMap {
				if value.Work.IsComplete() && t.Sub(value.UpdatedAt) > (LIFETIME*time.Second) {
					delete(workerMap, key)
					go func(w work.Work) {
						p := filepath.Join(scoreboard.workDir, w.Id)
						log.Println("Removing stale job: ", w.Id)
						os.RemoveAll(p)
					}(*value.Work)
				}
			}
		}
	}
}
