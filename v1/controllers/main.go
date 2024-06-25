package controllers

import (
	"context"

	"github.com/google/uuid"
)

const mb = 1048576

var ctx = context.Background()

func generateUUID() uuid.UUID {
	return uuid.New()
}
