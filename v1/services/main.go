package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	mathRand "math/rand/v2"
	"net/http"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/nfnt/resize"
	"github.com/rohan031/adgytec-api/v1/custom"
)

var db *pgxpool.Pool
var ctx context.Context = context.Background()
var spaceStorage *minio.Client
var firebaseClient *auth.Client

var expires time.Duration = time.Second * 60 * 60 // 1hr

func SetExternalConnection(pool *pgxpool.Pool, storage *minio.Client, client *auth.Client) {
	db = pool
	spaceStorage = storage
	firebaseClient = client
}

func generateSecureToken() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		log.Printf("Error generating token: %v\n", err)

		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateRandomString() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[mathRand.IntN(len(charset))]
	}
	return string(b)

}

func handleImage(img image.Image, buf *bytes.Buffer, format string) error {
	resizedImg := resize.Thumbnail(1920, 1080, img, resize.Lanczos3)
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err := jpeg.Encode(buf, resizedImg, &jpeg.Options{Quality: 80})
		if err != nil {
			log.Printf("failed to encode JPEG image: %v", err)
			return err
		}
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err := encoder.Encode(buf, resizedImg)
		if err != nil {
			log.Printf("failed to encode PNG image: %v", err)
			return err
		}
	default:
		log.Printf("unsupported image format: %s", format)
		message := "unsupported image format"
		return (&custom.MalformedRequest{
			Status: http.StatusUnsupportedMediaType, Message: message,
		})
	}

	return nil

}
