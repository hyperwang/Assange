package logging

import (
	"github.com/op/go-logging"
	"os"
)

const (
	CRITICAL int = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

func GetLogger(module string, level int) *logging.Logger {
	log := logging.MustGetLogger(module)
	logging.SetLevel(logging.Level(level), module)
	return log
}

func init() {
	var format = logging.MustStringFormatter(
		"%{color}%{time:2006-01-02 15:04:05.999Z-07:00} %{shortfile} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
	)
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)
}
