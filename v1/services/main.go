package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"image"
	"image/jpeg"
	"image/png"
	"io"
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
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rwcarlsen/goexif/exif"
	// "golang.org/x/image/webp"
)

var db *pgxpool.Pool
var ctx context.Context = context.Background()
var spaceStorage *minio.Client
var firebaseClient *auth.Client

var expires time.Duration = time.Second * 60 * 60 // 1hr
var week time.Duration = 604800 * time.Second

const webp = "image/webp"
const gif = "image/gif"
const svg = "image/svg+xml"

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

func reverseOrientation(img image.Image, o string) *image.NRGBA {
	switch o {
	case "1":
		return imaging.Clone(img)
	case "2":
		return imaging.FlipV(img)
	case "3":
		return imaging.Rotate180(img)
	case "4":
		return imaging.Rotate180(imaging.FlipV(img))
	case "5":
		return imaging.Rotate270(imaging.FlipV(img))
	case "6":
		return imaging.Rotate270(img)
	case "7":
		return imaging.Rotate90(imaging.FlipV(img))
	case "8":
		return imaging.Rotate90(img)
	}
	log.Printf("unknown orientation %s, expect 1-8", o)
	return imaging.Clone(img)
}

func handleImage(img image.Image, buf *bytes.Buffer, format string, file multipart.File) error {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		log.Printf("failed to seek file: %v\n", err)
	}

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		// resizedImg := resize.Thumbnail(1920, 1080, img, resize.Lanczos3)
		x, err := exif.Decode(file)
		if err != nil {
			log.Printf("failed reading exif data in %s\n", err.Error())
		}
		if x != nil && err == nil {
			orient, _ := x.Get(exif.Orientation)
			if orient != nil {
				img = reverseOrientation(img, orient.String())
			}
		}
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 80})
		if err != nil {
			log.Printf("failed to encode JPEG image: %v", err)
			return err
		}
	case "png":
		// resizedImg := resize.Thumbnail(1920, 1080, img, resize.Lanczos3)
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err := encoder.Encode(buf, img)
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

func uploadImageToCloudStorage(objectName string, buf io.Reader, size int64, contentType string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	_, err := spaceStorage.PutObject(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectName, buf, size, minio.PutObjectOptions{ContentType: contentType})
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

func deleteFromCloudStorage(objectName string) error {
	err := spaceStorage.RemoveObject(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("Error deleting image from space storage: %v\n", err)
		return err
	}

	return nil
}

// return type
// file to upload, file format, file content type, size of file, error if any
func handleRequestImage(file multipart.File, header *multipart.FileHeader) (io.Reader, string, string, int64, error) {
	contentType, err := isImageFile(header)
	if err != nil {
		return nil, "", "", 0, err
	}

	var format string
	var img image.Image
	buf := new(bytes.Buffer)

	if contentType == webp || contentType == svg || contentType == gif {
		switch contentType {
		case webp:
			format = "webp"
		case svg:
			format = "svg"
		case gif:
			format = "gif"
		}

		return file, format, contentType, header.Size, nil
	}

	img, format, err = image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v\n", err)
		return nil, "", "", 0, err
	}

	err = handleImage(img, buf, format, file)
	if err != nil {
		return nil, "", "", 0, err
	}

	return buf, format, contentType, int64(buf.Len()), nil
}
