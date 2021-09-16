package main

import (
	"github.com/turbosquid/imago/server"
	"github.com/twinj/uuid"
)

func main() {
	uuid.SwitchFormat(uuid.FormatHex)
	s := server.New()
	s.Run()
}
