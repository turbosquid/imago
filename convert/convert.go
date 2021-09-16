package convert

import (
	"github.com/turbosquid/imago/s3"
	"github.com/turbosquid/imago/scoreboard"
	"github.com/turbosquid/imago/settings"
	"github.com/turbosquid/imago/shellwords"
	"github.com/turbosquid/imago/work"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func doAction(settings *settings.Settings, scoreboard *scoreboard.Scoreboard, work *work.Work, action *work.Action) (err error) {
	creds := settings.Credentials[action.Credential]
	s3 := s3.New(creds.Key, creds.Secret, "us-east-1")
	defer func() {
		if err != nil {
			action.Status = "error"
			action.Error = err.Error()
			work.Status = "error"
		}
	}()

	local_infile := filepath.Join(settings.WorkDir, work.Id, action.Infile)
	local_outfile := filepath.Join(settings.WorkDir, work.Id, action.Outfile)

	os.MkdirAll(filepath.Dir(local_infile), 0755)
	os.MkdirAll(filepath.Dir(local_outfile), 0755)
	action.Status = "downloading"
	scoreboard.UpdateWork(work)
	err = s3.DownloadFile(action.Infile, local_infile)
	if err != nil {
		return err
	}
	action.Status = "converting"
	scoreboard.UpdateWork(work)

	args := make([]string, len(action.Operations)*2)
	args = append(args, settings.ImPath)
	args = append(args, shellwords.Escape(local_infile))
	for _, op := range action.Operations {
		ops := strings.Split("-"+op, " ")
		for _, o := range ops {
			args = append(args, o)
		}
	}
	args = append(args, shellwords.Escape(local_outfile))
	cmd := exec.Command("sh", "-c", strings.Join(args, " "))
	var output []byte
	output, err = cmd.CombinedOutput()
	action.Output = string(output)

	if err != nil {
		return err
	}
	action.Status = "uploading"
	scoreboard.UpdateWork(work)

	err = s3.UploadFile(local_outfile, action.Outfile, action.Mimetype)
	if err != nil {
		return err
	}
	action.Status = "done"
	scoreboard.UpdateWork(work)
	return err
}

func Convert(settings *settings.Settings, scoreboard *scoreboard.Scoreboard, work *work.Work) (err error) {

	work.Status = "running"
	scoreboard.UpdateWork(work)

	for idx, _ := range work.Actions {
		a := &work.Actions[idx]
		err = doAction(settings, scoreboard, work, a)
		if err != nil {
			break
		}
	}
	if err == nil {
		work.Status = "done"
	}
	scoreboard.UpdateWork(work)
	return err
}
