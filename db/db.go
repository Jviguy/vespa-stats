package db

import (
	"context"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect() {
	// Connect to the database
	dbconfig, err := pgxpool.ParseConfig("postgres://postgres@localhost:5432/r6stats-local?sslmode=prefer&pool_max_conns=10")
	if err != nil {
		// handle error
		panic(err)
	}
	dbconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}
	DB, err = pgxpool.NewWithConfig(context.Background(), dbconfig)
	if err != nil {
		// handle error
		panic(err)
	}
}
