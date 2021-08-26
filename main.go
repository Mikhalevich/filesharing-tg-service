package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/server"
	"github.com/sirupsen/logrus"
)

type params struct {
	ServiceName        string
	DBConnectionString string
	TGBotToken         string
}

func loadParams() (*params, error) {
	var p params
	p.ServiceName = os.Getenv("FS_SERVICE_NAME")
	if p.ServiceName == "" {
		p.ServiceName = "tg.service"
	}

	// p.DBConnectionString = os.Getenv("FS_DB_CONNECTION_STRING")
	// if p.DBConnectionString == "" {
	// 	return nil, errors.New("databse connection string is missing, please specify FS_DB_CONNECTION_STRING environment variable")
	// }

	p.TGBotToken = os.Getenv("FS_TG_BOT_TOKEN")
	if p.TGBotToken == "" {
		return nil, errors.New("telegram bot token is missing, please specify FS_TG_BOT_TOKEN environment variable")
	}

	return &p, nil
}

func makeLoggerWrapper(logger *logrus.Logger) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			logger.Infof("processing %s", req.Method())
			start := time.Now()
			defer logger.Infof("end processing %s, time = %v", req.Method(), time.Now().Sub(start))
			err := fn(ctx, req, rsp)
			if err != nil {
				logger.Errorln(err)
			}
			return err
		}
	}
}

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	p, err := loadParams()
	if err != nil {
		logger.Errorln(fmt.Errorf("unable to load params: %W", err))
		return
	}

	logger.Infof("running auth service with params: %v\n", p)

	srv := micro.NewService(
		micro.Name(p.ServiceName),
		micro.WrapHandler(makeLoggerWrapper(logger)),
	)

	srv.Init()

	ts, err := NewTelegramService(p.TGBotToken)
	if err != nil {
		logger.Errorln(err)
		return
	}
	micro.RegisterSubscriber("filesharing.file.event", srv.Server(), ts.StoreEvent, server.SubscriberQueue("filesharing.tg.service.queue"))

	err = srv.Run()
	if err != nil {
		logger.Errorln(err)
		return
	}
}
