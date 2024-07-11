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
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/google/uuid"
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
var week time.Duration = 604800 * time.Second

type IndexedValue struct {
	Index int
	Url   string
}

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

func isImageFile(header *multipart.FileHeader) (string, error) {
	contentType := header.Header.Get("Content-type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", (&custom.MalformedRequest{
			Status:  http.StatusUnsupportedMediaType,
			Message: http.StatusText(http.StatusUnsupportedMediaType),
		})
	}

	return contentType, nil
}

func handleImage(img image.Image, buf *bytes.Buffer, format string) error {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		resizedImg := resize.Thumbnail(1920, 1080, img, resize.Lanczos3)
		err := jpeg.Encode(buf, resizedImg, &jpeg.Options{Quality: 80})
		if err != nil {
			log.Printf("failed to encode JPEG image: %v", err)
			return err
		}
	case "png":
		resizedImg := resize.Thumbnail(1920, 1080, img, resize.Lanczos3)
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

func uploadImageToCloudStorage(objectName string, buf *bytes.Buffer, contentType string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	_, err := spaceStorage.PutObject(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectName, buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("failed to upload image: %v", err)
	}
	errChan <- err
}

func generatePresignedUrl(objectName string, ind int, expires time.Duration, wg *sync.WaitGroup, urlChan chan IndexedValue) {
	defer wg.Done()

	reqParams := make(url.Values)
	presignedURL, err := spaceStorage.PresignedGetObject(ctx,
		os.Getenv("SPACE_STORAGE_BUCKET_NAME"),
		objectName,
		expires,
		reqParams,
	)
	if err != nil {
		log.Printf("error generating presigned url for the image: %v\n", err)
		urlChan <- IndexedValue{
			Index: ind,
			Url:   "",
		}
		return
	}

	urlChan <- IndexedValue{
		Index: ind,
		Url:   presignedURL.String(),
	}
}

func GenerateUUID() uuid.UUID {
	return uuid.New()
}
