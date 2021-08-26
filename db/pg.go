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
