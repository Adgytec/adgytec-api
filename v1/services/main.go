package services

import "github.com/jackc/pgx/v5/pgxpool"

var db *pgxpool.Pool

func SetDatabasePool(pool *pgxpool.Pool) {
	db = pool
}
