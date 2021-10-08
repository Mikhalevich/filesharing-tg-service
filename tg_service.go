package main

import (
	"context"
	"fmt"

	"github.com/Mikhalevich/filesharing-tg-service/db"
	"github.com/Mikhalevich/filesharing/proto/event"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type storager interface {
	AddChat(u *db.Chat) error
	RemoveChat(chatID int) error
	GetChatsByStorage(storageName string) ([]*db.Chat, error)
}

type logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type TelegramService struct {
	storage     storager
	l           logger
	storageName string
	bot         *tgbotapi.BotAPI
}

func NewTelegramService(token string, storage storager, l logger, storageName string) (*TelegramService, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	//bot.Debug = true

	return &TelegramService{
		storage:     storage,
		l:           l,
		storageName: storageName,
		bot:         bot,
	}, nil
}

func (ts *TelegramService) RunTGBot() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := ts.bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() && update.Message.Text == "/start" {
			if err := ts.processAddChat(update.Message.Chat.ID, update.Message.From.ID, update.Message.MessageID); err != nil {
				ts.l.Errorf("run tg bot error: %v\n", err)
				continue
			}
		}
	}

	return nil
}

func (ts *TelegramService) processAddChat(chatID int64, userID int, messageID int) error {
	if err := ts.storage.AddChat(&db.Chat{
		ChatID:      int(chatID),
		UserID:      userID,
		StorageName: ts.storageName,
	}); err != nil {
		return fmt.Errorf("add chat error: %v\n", err)
	}

	msg := tgbotapi.NewMessage(chatID, "you now subscribed to storage events")
	msg.ReplyToMessageID = messageID
	if _, err := ts.bot.Send(msg); err != nil {
		return fmt.Errorf("send message error: %v\n", err)
	}

	return nil
}

func (ts *TelegramService) StoreEvent(ctx context.Context, req *event.FileEvent) error {
	if req.UserName != ts.storageName {
		ts.l.Infof("tg bot[%s] skiping event for %s\n", ts.storageName, req.UserName)
		return nil
	}

	var msgText string
	switch req.Action {
	case event.Action_Add:
		msgText = fmt.Sprintf("file \"%s\" added to storage \"%s\"", req.FileName, req.UserName)
	case event.Action_Remove:
		msgText = fmt.Sprintf("file \"%s\" removed from storage \"%s\"", req.FileName, req.UserName)
	}

	chats, err := ts.storage.GetChatsByStorage(req.UserName)
	if err != nil {
		return err
	}

	for _, c := range chats {
		msg := tgbotapi.NewMessage(int64(c.ChatID), msgText)
		if _, err := ts.bot.Send(msg); err != nil {
			return err
		}
	}
	return nil
}
