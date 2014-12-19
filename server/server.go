package server

import (
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/mowings/imago/convert"
	"github.com/mowings/imago/s3"
	"github.com/mowings/imago/scoreboard"
	"github.com/mowings/imago/settings"
	"github.com/mowings/imago/work"
	"log"
	"net/http"
	"os"
	"runtime"
)

const SETTINGS_FILE = "settings.yml"

type WorkChan chan work.Work

type Server struct {
	ServerSettings *settings.Settings
	Scoreboard     *scoreboard.Scoreboard
	martini        *martini.ClassicMartini
}

func (server *Server) error_response(r render.Render, s string) {
	r.JSON(400, map[string]interface{}{"status": "error", "message": s})
}

func (server *Server) worker(id int, workchan WorkChan) {
	s3conn := s3.New(server.ServerSettings.AwsKey, server.ServerSettings.AwsSecret, "us-east-1")
	var _ = s3conn
	for {
		w := <-workchan
		log.Printf("Worker got: %+v\n", w)
		err := convert.Convert(server.ServerSettings, server.Scoreboard, s3conn, &w)
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
		log.Printf("Worker %d processed %s\n", id, w.Id)
	}
}

func (server *Server) startWorkers(workchan WorkChan) {
	for i := 1; i <= server.ServerSettings.NumWorkers; i++ {
		go server.worker(i, workchan)
	}
}

func (server *Server) addNewWork(req *http.Request, r render.Render, c WorkChan) {
	decoder := json.NewDecoder(req.Body)
	var w work.Work
	err := decoder.Decode(&w)
	if err != nil {
		server.error_response(r, "JSON Parse: "+err.Error())
	} else {
		w.Initialize()
		server.Scoreboard.UpdateWork(&w)
		c <- w
		r.JSON(200, map[string]interface{}{"status": "ok", "id": w.Id})
	}
}

func (server *Server) queueLength(r render.Render, c WorkChan) {
	r.JSON(200, map[string]interface{}{"status": "ok", "length": len(c)})
}

func (server *Server) getWorkById(id string) *work.Work {
	c := make(chan work.Work)
	w := scoreboard.WorkStatusRequest{Id: id, Chan: c, LongPoll: false}
	server.Scoreboard.GetWorkChannel <- w
	s := <-c
	if s.Status == "" {
		return nil
	} else {
		return &s
	}
}

func (server *Server) longPollGetWork(params martini.Params, r render.Render) {
	w := server.getWorkById(params["id"])
	if w == nil {
		r.JSON(404, "Not found")
	} else {
		if w.Status != "done" && w.Status != "error" {
			lpmsg := scoreboard.LongPollChanMessage{w.Id, nil}
			server.Scoreboard.LongPollChanRequest <- lpmsg
			lpresponse := <-server.Scoreboard.LongPollChanRequest
			log.Println("Waiting on work completion")
			r := <-*lpresponse.LongPollChan
			w = &r
		}

		r.JSON(200, w)
	}

}

func (server *Server) getWork(params martini.Params, r render.Render) {
	w := server.getWorkById(params["id"])
	if w == nil {
		r.JSON(404, "Not found")
	} else {
		r.JSON(200, w)
	}
}

func (server *Server) Run() {
	server.martini.Run()
}

func New() (server *Server) {
	server = new(Server)
	server.ServerSettings = settings.LoadSettings(SETTINGS_FILE)
	log.Println("Cleaning up working directory", server.ServerSettings.WorkDir, "...")
	os.RemoveAll(server.ServerSettings.WorkDir)

	log.Printf("MAXPROCS is: %d\n", runtime.GOMAXPROCS(0))
	log.Printf("Settings: %+v\n", *(server.ServerSettings.SafeCopy()))
	server.Scoreboard = scoreboard.New()
	var workQueue = make(chan work.Work, server.ServerSettings.QueueSize)

	server.martini = martini.Classic()
	server.martini.Use(render.Renderer())
	server.martini.Get("/", func() string {
		return "Image manipulation service. API Only."
	})
	server.martini.Post("/api/v1/work", func(req *http.Request, r render.Render) {
		server.addNewWork(req, r, workQueue)
	})
	server.martini.Get("/api/v1/queue_length", func(r render.Render) {
		server.queueLength(r, workQueue)
	})
	server.martini.Get("/api/v1/work/:id", func(params martini.Params, r render.Render) {
		server.getWork(params, r)
	})
	server.martini.Get("/api/v1/work/poll/:id", func(params martini.Params, r render.Render) {
		server.longPollGetWork(params, r)
	})

	server.startWorkers(workQueue)
	return server
}
