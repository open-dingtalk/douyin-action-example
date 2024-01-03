package main

import (
	"douyin-action-example/internal/actions"
	"douyin-action-example/internal/conf"
	"github.com/chzealot/gobase/logger"
)

func main() {
	if err := logger.InitWithConfig(conf.AppConfig); err != nil {
		panic(err)
	}
	logger.Infof("start DingTalk calendar ...")

	server := actions.NewHttpServer()
	if err := server.Run(":3021"); err != nil {
		panic(err)
	}
	select {}
}
