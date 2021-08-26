package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/Mikhalevich/filesharing/proto/event"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramService struct {
	m     sync.Mutex
	users map[int]int64
	bot   *tgbotapi.BotAPI
}

func NewTelegramService(token string) (*TelegramService, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	//bot.Debug = true

	return &TelegramService{
		users: make(map[int]int64),
		bot:   bot,
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

		ts.m.Lock()

		if _, ok := ts.users[update.Message.From.ID]; !ok {
			ts.users[update.Message.From.ID] = update.Message.Chat.ID

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "you now subscribed to storage events")
			msg.ReplyToMessageID = update.Message.MessageID
			ts.bot.Send(msg)
		}

		ts.m.Unlock()
	}

	return nil
}

func (ts *TelegramService) StoreEvent(ctx context.Context, req *event.FileEvent) error {
	msgText := ""
	switch req.Action {
	case event.Action_Add:
		msgText = fmt.Sprintf("file %s added to storage %s", req.FileName, req.UserName)
	case event.Action_Remove:
		msgText = fmt.Sprintf("file %s removed to storage %s", req.FileName, req.UserName)
	}

	ts.m.Lock()
	defer ts.m.Unlock()

	for _, chatID := range ts.users {
		msg := tgbotapi.NewMessage(chatID, msgText)
		ts.bot.Send(msg)
	}
	return nil
}
