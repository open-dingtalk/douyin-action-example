package conf

import (
	"github.com/chzealot/gobase/logger"
	"os"
	"strings"
)

var AppConfig logger.Config

var IsDebugMode = false

func init() {
	AppConfig = logger.Config{
		AppName:   "douyin-action-example",
		DebugMode: logger.DebugModeFromEnv,
	}

	debug := strings.ToLower(os.Getenv("DEBUG"))
	if debug == "true" || debug == "on" || debug == "enable" || debug == "1" {
		IsDebugMode = true
	}

}
