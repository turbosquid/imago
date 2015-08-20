package settings

import (
	"log"
	"testing"
)

func TestLoad(t *testing.T) {
	settings := LoadSettings("settings.test.yml")
	log.Printf("Loaded settings:\n")
	log.Printf("%v\n", settings)
}
