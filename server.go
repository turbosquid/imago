package main

import (
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"os"
	"runtime"
	"server/convert"
	"server/s3"
	"server/scoreboard"
	"server/settings"
	"server/work"
)

const SETTINGS_FILE = "settings.yml"

type WorkChan chan work.Work

var ServerSettings *settings.Settings

func error_response(r render.Render, s string) {
	r.JSON(400, map[string]interface{}{"status": "error", "message": s})
}

func worker(id int, workchan WorkChan) {
	s3conn := s3.New(ServerSettings.AwsKey, ServerSettings.AwsSecret, "us-east-1")
	var _ = s3conn
	for {
		w := <-workchan
		log.Printf("Worker got: %+v\n", w)
		err := convert.Convert(ServerSettings, s3conn, &w)
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
		log.Printf("Worker %d processed %s\n", id, w.Id)
	}
}

func startWorkers(workchan WorkChan) {
	for i := 1; i <= ServerSettings.NumWorkers; i++ {
		go worker(i, workchan)
	}
}

func addNewWork(req *http.Request, r render.Render, c WorkChan) {
	decoder := json.NewDecoder(req.Body)
	var w work.Work
	err := decoder.Decode(&w)
	if err != nil {
		error_response(r, "JSON Parse: "+err.Error())
	} else {
		w.Initialize()
		scoreboard.UpdateWork(&w)
		c <- w
		r.JSON(200, map[string]interface{}{"status": "ok", "id": w.Id})
	}
}

func queueLength(r render.Render, c WorkChan) {
	r.JSON(200, map[string]interface{}{"status": "ok", "length": len(c)})
}

func getWork(params martini.Params, r render.Render) {
	var w work.Work
	w.Id = params["id"]
	scoreboard.GetWorkChannel <- w
	s := <-scoreboard.GetWorkChannel
	if s.Status == "" {
		r.JSON(404, "Not found")
	} else {
		r.JSON(200, s)
	}
}

func main() {
	ServerSettings = settings.LoadSettings(SETTINGS_FILE)
	log.Println("Cleaning up working directory", ServerSettings.WorkDir, "...")
	os.RemoveAll(ServerSettings.WorkDir)

	log.Printf("MAXPROCS is: %d\n", runtime.GOMAXPROCS(0))
	log.Printf("Settings: %+v\n", *(ServerSettings.SafeCopy()))
	var workQueue = make(chan work.Work, ServerSettings.QueueSize)

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Get("/", func() string {
		return "Image manipulation service. API Only."
	})
	m.Post("/api/v1/work", func(req *http.Request, r render.Render) {
		addNewWork(req, r, workQueue)
	})
	m.Get("/api/v1/queue_length", func(r render.Render) {
		queueLength(r, workQueue)
	})
	m.Get("/api/v1/work/:id", getWork)

	startWorkers(workQueue)
	scoreboard.Start()
	m.Run()
}
