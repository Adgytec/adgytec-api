package controllers

import (
	"bytes"
	"image"
	"log"
	"net/http"
	"os"
	"strings"

	_ "image/gif"
	"image/jpeg"
	"image/png"

	"github.com/minio/minio-go/v7"
	"github.com/nfnt/resize"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/storage"
	"github.com/rohan031/adgytec-api/v1/custom"
)

func PostNews(w http.ResponseWriter, r *http.Request) {
	maxSize := 10 << 20

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxSize))
	err := r.ParseMultipartForm(int64(maxSize))
	if err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "http: request body too large") {
			messgage := "request body too large. Limit 10MB"
			helper.HandleError(w, &custom.MalformedRequest{Status: http.StatusRequestEntityTooLarge, Message: messgage})
			return
		}

		if strings.Contains(err.Error(), "mime: no media type") {
			helper.HandleError(w, &custom.MalformedRequest{
				Status:  http.StatusUnsupportedMediaType,
				Message: http.StatusText(http.StatusUnsupportedMediaType),
			})
			return
		}

		if strings.Contains(err.Error(), "request Content-Type isn't multipart/form-data") {
			helper.HandleError(w, &custom.MalformedRequest{
				Status:  http.StatusBadRequest,
				Message: "Request Content-Type isn't multipart/form-data",
			})
			return
		}

		log.Printf("Error parsing multipart form data: %v\n", err)
		helper.HandleError(w, err)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error retriving file: %v\n ", err)
		helper.HandleError(w, err)
		return
	}
	defer file.Close()

	contentType := handler.Header.Get("Content-type")
	if !strings.HasPrefix(contentType, "image/") {
		helper.HandleError(w, &custom.MalformedRequest{
			Status:  http.StatusUnsupportedMediaType,
			Message: http.StatusText(http.StatusUnsupportedMediaType),
		})
		return
	}

	img, format, err := image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v\n", err)
		helper.HandleError(w, err)
		return
	}

	resizedImg := resize.Thumbnail(1920, 1080, img, resize.Lanczos3)
	buf := new(bytes.Buffer)
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(buf, resizedImg, &jpeg.Options{Quality: 80})
		if err != nil {
			log.Printf("failed to encode JPEG image: %v", err)
			helper.HandleError(w, err)
			return
		}
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(buf, resizedImg)
		if err != nil {
			log.Printf("failed to encode PNG image: %v", err)
			helper.HandleError(w, err)
			return
		}
	default:
		log.Printf("unsupported image format: %s", format)
		message := "unsupported image format"
		helper.HandleError(w, &custom.MalformedRequest{
			Status: http.StatusUnsupportedMediaType, Message: message,
		})
		return
	}

	objectName := "services/news/image." + format

	_, err = storage.SpaceStorage.PutObject(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectName, buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("failed to upload file: %v", err)
		helper.HandleError(w, err)
		return
	}

	log.Printf("Successfully uploaded %s\n", objectName)

	// fmt.Println(img)
}
