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
	"strconv"
	"time"
)

const SETTINGS_FILE = "settings.yml"

type WorkStatusResult int

const (
	GetOk WorkStatusResult = iota
	GetNotFound
	GetTimeout
)

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
		queue_length := len(c)
		c <- w
		r.JSON(200, map[string]interface{}{"status": "ok", "id": w.Id, "queue_length": queue_length})
	}
}

func (server *Server) queueLength(r render.Render, c WorkChan) {
	r.JSON(200, map[string]interface{}{"status": "ok", "length": len(c)})
}

func (server *Server) getWorkById(id string, timeout int) (*work.Work, WorkStatusResult) {
	c := make(chan work.Work)
	longpoll := false
	if timeout > 0 {
		longpoll = true
	}
	w := scoreboard.WorkStatusRequest{Id: id, Chan: c, LongPoll: longpoll}
	server.Scoreboard.GetWorkChannel <- w
	var s work.Work
	if longpoll {
		select {
		case s = <-c:
		case <-time.After(time.Duration(timeout) * time.Second):
			return &s, GetTimeout
		}
	} else {
		s = <-c
	}
	if s.Status == "" {
		return &s, GetNotFound
	} else {
		return &s, GetOk
	}
}

func (server *Server) getWork(params martini.Params, r render.Render, req *http.Request) {
	timeout := 0
	timeout, _ = strconv.Atoi(req.URL.Query().Get("timeout"))
	w, res := server.getWorkById(params["id"], timeout)
	switch res {
	case GetOk:
		r.JSON(200, w)
	case GetNotFound:
		r.JSON(404, "Not found")
	case GetTimeout:
		r.JSON(202, "Timed out checking work status")
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
	server.Scoreboard = scoreboard.New(server.ServerSettings.WorkDir)
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
	server.martini.Get("/api/v1/work/:id", func(params martini.Params, r render.Render, req *http.Request) {
		server.getWork(params, r, req)
	})

	server.startWorkers(workQueue)
	return server
}
