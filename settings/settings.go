package settings

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

const NUM_WORKERS = 4
const QUEUE_SIZE = 1000

type Credential struct {
	Key    string `yaml: key`
	Secret string `yaml: "secret"`
}

type Settings struct {
	QueueSize   int                   `yaml:"queue_size"`
	NumWorkers  int                   `yaml:"num_workers"`
	Credentials map[string]Credential `yaml: "credentials"`
	WorkDir     string                `yaml:"work_dir"`
	ImPath      string                `yaml:"im_path"`
}

func LoadSettings() *Settings {
	settingsPath := os.Getenv("IMAGO_SETTINGS_PATH")
	if settingsPath == "" {
		settingsPath = "./"
	}
	settingsFn := filepath.Join(settingsPath, "settings.yml")
	credsFn := filepath.Join(settingsPath, "credentials.yml")
	settings := Settings{QUEUE_SIZE, NUM_WORKERS, nil, ".", ""}
	loadSettings(settingsFn, &settings)
	loadSettings(credsFn, &settings)

	if settings.Credentials == nil {
		panic("No credentials provided.")
	}
	if _, ok := settings.Credentials["default"]; ok == false {
		panic("No default credential defined.")
	}
	return &settings
}

func loadSettings(fn string, settings *Settings) {
	if _, err := os.Stat(fn); err != nil {
		panic("Unable to open config file " + fn + ": " + err.Error())
	}
	dat, _ := ioutil.ReadFile(fn)
	err := yaml.Unmarshal(dat, settings)
	if err != nil {
		panic("Unable to parse YAML: " + err.Error())
	}
}

func (s *Settings) SafeCopy() *Settings {
	safe_setting := *s
	safe_setting.Credentials = make(map[string]Credential)
	for name, cred := range s.Credentials {
		safe_setting.Credentials[name] = Credential{Key: cred.Key, Secret: "CENSORED"}
	}
	return &safe_setting
}
