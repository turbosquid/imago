package convert

import (
	"os"
	"os/exec"
	"path/filepath"
	"server/s3"
	"server/settings"
	"server/work"
	"strings"
)

func Convert(settings *settings.Settings, s3 *s3.S3Connection, work *work.Work) (err error) {
	local_infile := filepath.Join(settings.WorkDir, work.Id, "infile", filepath.Base(work.Infile))
	local_outfile := filepath.Join(settings.WorkDir, work.Id, "outfile", filepath.Base(work.Outfile))
	os.MkdirAll(filepath.Dir(local_infile), 0755)
	os.MkdirAll(filepath.Dir(local_outfile), 0755)
	err = s3.DownloadFile(work.Infile, local_infile)
	if err != nil {
		return err
	}
	args := make([]string, len(work.Operations)+2)
	args = append(args, "/usr/local/bin/convert")
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
	err = s3.UploadFile(local_outfile, work.Outfile, work.Mimetype)
	if err != nil {
		return err
	}
	return err
}
