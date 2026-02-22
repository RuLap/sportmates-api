package postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

var db *pgxpool.Pool

func InitDB(connStr string) (*Storage, error) {
	var err error
	db, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Database() *pgxpool.Pool {
	return s.db
}
