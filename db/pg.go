package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(connectionStr string) (*Postgres, error) {
	pgDB, err := sqlx.Connect("postgres", connectionStr)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		db: pgDB,
	}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) AddChat(u *Chat) error {
	_, err := p.db.NamedExec("INSERT OR IGNORE INTO Chat(chat_id, user_id, storage_name) VALUES(:chat_id, :user_id, :storage_name)", u)
	return err
}

func (p *Postgres) RemoveChat(chatID int) error {
	_, err := p.db.Exec("DELETE FROM Chat WHERE chat_id = $1", chatID)
	return err
}

func (p *Postgres) GetChatsByStorage(storageName string) ([]*Chat, error) {
	var chats []*Chat
	if err := p.db.Select(chats, "SELECT * FROM Chat WHERE storage_name = $1", storageName); err != nil {
		return nil, err
	}
	return chats, nil
}
