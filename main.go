package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Mikhalevich/filesharing-tg-service/db"
	"github.com/Mikhalevich/filesharing/service"
	"github.com/Mikhalevich/repeater"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/server"

	_ "github.com/asim/go-micro/plugins/broker/nats/v3"
)

type params struct {
	ServiceName        string
	DBConnectionString string
	TGBotToken         string
	StorageName        string
}

func loadParams() (*params, error) {
	var p params
	p.ServiceName = os.Getenv("FS_SERVICE_NAME")
	if p.ServiceName == "" {
		p.ServiceName = "tg.service"
	}

	p.DBConnectionString = os.Getenv("FS_DB_CONNECTION_STRING")
	if p.DBConnectionString == "" {
		return nil, errors.New("databse connection string is missing, please specify FS_DB_CONNECTION_STRING environment variable")
	}

	p.TGBotToken = os.Getenv("FS_TG_BOT_TOKEN")
	if p.TGBotToken == "" {
		return nil, errors.New("telegram bot token is missing, please specify FS_TG_BOT_TOKEN environment variable")
	}

	p.StorageName = os.Getenv("FS_STORAGE_NAME")
	if p.StorageName == "" {
		return nil, errors.New("storage name is missing, please specify FS_STORAGE_NAME environment variable")
	}

	return &p, nil
}

func main() {
	p, err := loadParams()
	if err != nil {
		fmt.Printf("unable to load params: %v\n", err)
		return
	}

	srv := service.New(p.ServiceName)

	srv.Logger().Infof("running tg service with params: %v\n", p)

	var storage *db.Postgres
	if err := repeater.Do(
		func() error {
			storage, err = db.NewPostgres(p.DBConnectionString)
			return err
		},
		repeater.WithTimeout(time.Second*1),
		repeater.WithLogger(srv.Logger()),
		repeater.WithLogMessage("try to connect to database"),
	); err != nil {
		srv.Logger().Errorf("unable to connect to database: %v\n", err)
		return
	}
	defer storage.Close()

	srv.RegisterHandler(func(ms micro.Service, sv service.Servicer) error {
		ts, err := NewTelegramService(p.TGBotToken, storage, sv.Logger(), p.StorageName)
		if err != nil {
			return fmt.Errorf("unable to create tg service: %v", err)
		}
		micro.RegisterSubscriber("filesharing.file.event", ms.Server(), ts.StoreEvent, server.SubscriberQueue("filesharing.tg.service.queue"))

		go func() {
			if err := ts.RunTGBot(); err != nil {
				sv.Logger().Errorf("unable to run tg bot: %v", err)
			}
		}()
		return nil
	})

	srv.Run()
}
