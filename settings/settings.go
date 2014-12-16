package settings

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const NUM_WORKERS = 4
const QUEUE_SIZE = 1000

type Settings struct {
	QueueSize  int    `yaml:"queue_size"`
	NumWorkers int    `yaml:"num_workers"`
	AwsKey     string `yaml:"aws_key"`
	AwsSecret  string `yaml:"aws_secret"`
	WorkDir    string `yaml:"work_dir"`
	ImPath     string `yaml:"im_path"`
}

func LoadSettings(fn string) *Settings {
	if _, err := os.Stat(fn); err != nil {
		panic("Unable to open config file " + fn + ": " + err.Error())
	}
	dat, _ := ioutil.ReadFile(fn)
	settings := Settings{NUM_WORKERS, QUEUE_SIZE, "changeme", "changeme", ".", ""}
	err := yaml.Unmarshal(dat, &settings)
	if err != nil {
		panic("Unable to parse YAML: " + err.Error())
	}
	return &settings
}

func (s *Settings) SafeCopy() *Settings {
	safe_setting := *s
	safe_setting.AwsSecret = "CENSORED"
	return &safe_setting
}
