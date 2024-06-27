package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type FileMetaData struct {
	Path string `json:"path"`
}

type BlogMedia struct {
	Paths []string `json:"paths,omitempty"`
}

type Blog struct {
	Title     string    `json:"title" db:"title"`
	Summary   string    `json:"summary,omitempty" db:"short_text"`
	Content   string    `json:"content,omitempty" db:"content"`
	Author    string    `json:"author" db:"author"`
	Id        string    `json:"blogId" db:"blog_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Cover     string    `json:"cover" db:"cover_image"`
}

type BlogSummary struct {
	Title     string    `json:"title" db:"title"`
	Summary   string    `json:"summary,omitempty" db:"short_text"`
	Author    string    `json:"author" db:"author"`
	Id        string    `json:"blogId" db:"blog_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	Cover     string    `json:"cover" db:"cover_image"`
}

func (bm *BlogMedia) UploadMedia(r *http.Request) (error, bool) {
	metadataJSON := r.FormValue("metadata")
	var metadata []FileMetaData
	err := json.Unmarshal([]byte(metadataJSON), &metadata)
	if err != nil {
		return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: "Invalid file metadata."}, false
	}

	isSuccess := true
	for i, meta := range metadata {
		go func(index int, metadata FileMetaData) {
			file, header, err := r.FormFile(fmt.Sprintf("media_%d", index))
			if err != nil {
				log.Printf("error reteriving file: %v\n", err)
				isSuccess = false
				return
			}
			defer file.Close()

			contentType := header.Header.Get("Content-type")
			if !strings.HasPrefix(contentType, "image/") {
				isSuccess = false
				return
			}

			img, format, err := image.Decode(file)
			if err != nil {
				log.Printf("Error decoding image: %v\n", err)
				isSuccess = false
				return
			}

			buf := new(bytes.Buffer)
			err = handleImage(img, buf, format)
			if err != nil {
				isSuccess = false
				return
			}

			_, err = spaceStorage.PutObject(ctx,
				os.Getenv("SPACE_STORAGE_BUCKET_NAME"),
				metadata.Path,
				buf,
				int64(buf.Len()), minio.PutObjectOptions{ContentType: contentType})
			if err != nil {
				isSuccess = false
				return
			}

		}(i, meta)
	}

	return nil, isSuccess
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

func addBlogToDatabase(b *Blog, projectId, userId string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	args := dbqueries.CreateBlogItemArgs(b.Id, userId, projectId, b.Title,
		b.Cover, b.Summary, b.Content, b.Author)

	_, err := db.Exec(ctx, dbqueries.CreateBlogItem, args)
	if err != nil {
		log.Printf("Error adding blog item in database: %v\n", err)
	}
	errChan <- err
}

func (b *Blog) CreateBlog(r *http.Request, projectId, userId string) error {
	file, header, err := r.FormFile("cover")
	if err != nil {
		log.Printf("Error retriving file: %v\n ", err)
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

	objectName := fmt.Sprintf("services/blogs/%v/%v/%v.%v", projectId, b.Id, generateRandomString(), format)

	if val := os.Getenv("ENV"); val == "dev" {
		objectName = "dev/" + objectName
	}
	b.Cover = objectName

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)

	wg.Add(2)

	go uploadImageToCloudStorage(objectName, buf, contentType, wg, errChan)
	go addBlogToDatabase(b, projectId, userId, wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Blog) GetBlogsByProjectId(projectId string) (*[]BlogSummary, error) {
	args := dbqueries.GetBlogsByProjectIdArgs(projectId)
	rows, err := db.Query(ctx, dbqueries.GetBlogsByProjectId, args)

	if err != nil {
		log.Printf("Error fetching blogs from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	blogs, err := pgx.CollectRows(rows, pgx.RowToStructByName[BlogSummary])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	wg := new(sync.WaitGroup)
	urlChan := make(chan IndexedValue, len(blogs))

	for ind, item := range blogs {
		wg.Add(1)

		img := item.Cover
		go generatePresignedUrl(img, ind, expires, wg, urlChan)
	}

	wg.Wait()
	close(urlChan)

	for url := range urlChan {
		ind := url.Index
		blogs[ind].Cover = url.Url
	}

	return &blogs, nil
}
