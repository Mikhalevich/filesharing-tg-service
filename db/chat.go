package db

type Chat struct {
	ChatID      int    `db:"chat_id"`
	UserID      int    `db:"user_id"`
	StorageName string `db:"storage_name"`
}
