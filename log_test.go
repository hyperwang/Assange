package Assange

import (
	"testing"
)

func TestGetLogger_1(t *testing.T) {
	log := GetLogger("test", WARNING)
	log.Debug("debug")
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")
	log.Fatal("fatal")
}
