package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool
var ctx context.Context = context.Background()

func SetDatabasePool(pool *pgxpool.Pool) {
	db = pool
}
