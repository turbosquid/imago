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

func Convert(settings *settings.Settings, scoreboard *scoreboard.Scoreboard, s3 *s3.S3Connection, work *work.Work) (err error) {

	defer func() {
		if err != nil {
			work.Status = "error"
			scoreboard.UpdateWork(work)
		}
	}()

	local_infile := filepath.Join(settings.WorkDir, work.Id, "infile", filepath.Base(work.Infile))
	local_outfile := filepath.Join(settings.WorkDir, work.Id, "outfile", filepath.Base(work.Outfile))
	os.MkdirAll(filepath.Dir(local_infile), 0755)
	os.MkdirAll(filepath.Dir(local_outfile), 0755)
	work.Status = "downloading"
	scoreboard.UpdateWork(work)
	err = s3.DownloadFile(work.Infile, local_infile)
	if err != nil {
		return err
	}
	work.Status = "converting"
	scoreboard.UpdateWork(work)
	args := make([]string, len(work.Operations)*2)
	args = append(args, settings.ImPath)
	args = append(args, local_infile)
	for _, op := range work.Operations {
		ops := strings.Split("-"+op, " ")
		for _, o := range ops {
			args = append(args, o)
		}
	}
	args = append(args, local_outfile)
	cmd := exec.Command("sh", "-c", strings.Join(args, " "))
	var output []byte
	output, err = cmd.CombinedOutput()
	work.Output = string(output)

	if err != nil {
		return err
	}
	work.Status = "uploading"
	scoreboard.UpdateWork(work)

	err = s3.UploadFile(local_outfile, work.Outfile, work.Mimetype)
	if err != nil {
		return err
	}
	work.Status = "done"
	scoreboard.UpdateWork(work)
	return err
}
