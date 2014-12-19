package convert

import (
	"github.com/mowings/imago/s3"
	"github.com/mowings/imago/scoreboard"
	"github.com/mowings/imago/settings"
	"github.com/mowings/imago/work"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func doAction(settings *settings.Settings, scoreboard *scoreboard.Scoreboard, s3 *s3.S3Connection, work *work.Work, action *work.Action) (err error) {
	defer func() {
		if err != nil {
			action.Status = "error"
			action.Error = err.Error()
			work.Status = "error"
		}
	}()

	local_infile := filepath.Join(settings.WorkDir, work.Id, "infile", filepath.Base(action.Infile))
	local_outfile := filepath.Join(settings.WorkDir, work.Id, "outfile", filepath.Base(action.Outfile))
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
	args = append(args, local_infile)
	for _, op := range action.Operations {
		ops := strings.Split("-"+op, " ")
		for _, o := range ops {
			args = append(args, o)
		}
	}
	args = append(args, local_outfile)
	cmd := exec.Command("sh", "-c", strings.Join(args, " "))
	var output []byte
	output, err = cmd.CombinedOutput()
	action.Output = string(output)

	if err != nil {
		return err
	}
	work.Status = "uploading"
	scoreboard.UpdateWork(work)

	err = s3.UploadFile(local_outfile, action.Outfile, action.Mimetype)
	if err != nil {
		return err
	}
	action.Status = "done"
	scoreboard.UpdateWork(work)
	return err
}

func Convert(settings *settings.Settings, scoreboard *scoreboard.Scoreboard, s3 *s3.S3Connection, work *work.Work) (err error) {

	work.Status = "running"
	scoreboard.UpdateWork(work)

	for idx, _ := range work.Actions {
		a := &work.Actions[idx]
		err = doAction(settings, scoreboard, s3, work, a)
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
