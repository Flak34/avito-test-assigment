package db

import (
	"avito-test-assigment/banner_app/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewDb(ctx context.Context, db config.DB) (*Database, error) {
	pool, err := pgxpool.Connect(ctx, generateDsn(db))
	if err != nil {
		return nil, err
	}
	return newDatabase(pool), nil
}

func generateDsn(db config.DB) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", db.Host, db.Port, db.User, db.Password, db.DbName)
}
