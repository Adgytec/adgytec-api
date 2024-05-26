package services

import (
	"context"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
)

var db *pgxpool.Pool
var ctx context.Context = context.Background()
var spaceStorage *minio.Client
var firebaseClient *auth.Client

func SetExternalConnection(pool *pgxpool.Pool, storage *minio.Client, client *auth.Client) {
	db = pool
	spaceStorage = storage
	firebaseClient = client
}
