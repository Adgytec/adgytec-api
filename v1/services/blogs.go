package services

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/rohan031/adgytec-api/v1/custom"
)

type BlogMedia struct {
	URL   string   `json:"image,omitempty"`
	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths,omitempty"`
}

func (bm *BlogMedia) UploadMedia(r *http.Request, projectId, blogId string) error {
	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error retriving file: %v\n", err)
		return err
	}
	defer file.Close()

	contentType := header.Header.Get("Content-type")
	if !strings.HasPrefix(contentType, "image/") {
		return (&custom.MalformedRequest{
			Status:  http.StatusUnsupportedMediaType,
			Message: http.StatusText(http.StatusUnsupportedMediaType),
		})
	}

	img, format, err := image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v\n", err)
		return err
	}

	buf := new(bytes.Buffer)
	err = handleImage(img, buf, format)
	if err != nil {
		return err
	}

	objectName := fmt.Sprintf("services/blogs/%v/%v/%v.%v",
		projectId,
		blogId,
		generateRandomString(),
		format,
	)
	if val := os.Getenv("ENV"); val == "dev" {
		objectName = "dev/" + objectName
	}

	bm.Path = objectName

	_, err = spaceStorage.PutObject(ctx,
		os.Getenv("SPACE_STORAGE_BUCKET_NAME"),
		objectName,
		buf,
		int64(buf.Len()),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		log.Printf("failed to upload media: %v", err)
		return err
	}

	reqParams := make(url.Values)
	presignedUrl, err := spaceStorage.PresignedGetObject(ctx,
		os.Getenv("SPACE_STORAGE_BUCKET_NAME"),
		objectName,
		expires,
		reqParams,
	)
	if err != nil {
		log.Printf("error generating presigned url for the image: %v\n", err)
		return nil
	}

	bm.URL = presignedUrl.String()
	return nil
}

func (bm *BlogMedia) DeleteMedia() error {
	if len(bm.Paths) == 0 {
		return nil
	}

	objectChan := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectChan)
		for _, path := range bm.Paths {
			objectChan <- minio.ObjectInfo{Key: path}
		}
	}()

	e := spaceStorage.RemoveObjects(ctx,
		os.Getenv("SPACE_STORAGE_BUCKET_NAME"),
		objectChan,
		minio.RemoveObjectsOptions{},
	)

	isErr := false
	for err := range e {
		log.Printf("Error deleting objects in space storage, %v\n", err)
		isErr = true
	}

	if isErr {
		return errors.New("error deleting image from space storage")
	}

	return nil
}
