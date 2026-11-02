package httpserver

import (
	"all-in-one/config"
	"testing"
)

func TestMain(t *testing.T) {
	StartGin(config.Conf.Server)
}
