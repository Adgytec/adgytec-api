package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type News struct {
	Title string    `json:"title" db:"title"`
	Link  string    `json:"link" db:"link"`
	Text  string    `json:"text" db:"text"`
	Image string    `json:"image" db:"image"`
	Id    string    `json:"id,omitempty" db:"news_id"`
	Date  time.Time `json:"createdAt,omitempty" db:"created_at"`
}

func uploadImageToCloudStorage(objectName string, buf *bytes.Buffer, contentType string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	_, err := spaceStorage.PutObject(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectName, buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("failed to upload file: %v", err)
	}
	errChan <- err
}

func addNewsToDatabase(n *News, projectId string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	args := dbqueries.CreateNewsItemArgs(n.Title, n.Link, n.Text, n.Image, projectId)
	_, err := db.Exec(ctx, dbqueries.CreateNewsItem, args)
	if err != nil {
		log.Printf("Error updating user in database: %v\n", err)
	}
	errChan <- err
}

func (n *News) CreateNewsItem(r *http.Request, projectId string) error {
	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error retriving file: %v\n ", err)
		return err
	}
	defer file.Close()

	contentType := handler.Header.Get("Content-type")
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

	objectName := fmt.Sprintf("services/news/%v-%v.%v", n.Title, generateRandomString(), format)
	n.Image = objectName
	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)

	wg.Add(2)

	go uploadImageToCloudStorage(objectName, buf, contentType, wg, errChan)
	go addNewsToDatabase(n, projectId, wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

type IndexedValue struct {
	Index int
	Url   string
}

func generatePresignedUrl(objectName string, ind int, wg *sync.WaitGroup, urlChan chan IndexedValue) {
	defer wg.Done()

	expires := time.Second * 60 * 60
	reqParams := make(url.Values)
	presignedURL, err := spaceStorage.PresignedGetObject(context.Background(), os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectName, expires, reqParams)
	if err != nil {
		log.Fatalln(err)
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

func (n *News) GetAllNewsByProjectId(projectId string) (*[]News, error) {
	args := dbqueries.GetAllNewsByProjectIdArgs(projectId)
	rows, err := db.Query(ctx, dbqueries.GetAllNewsByProjectId, args)
	if err != nil {
		log.Printf("Error fetching news from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	news, err := pgx.CollectRows(rows, pgx.RowToStructByName[News])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	wg := new(sync.WaitGroup)
	urlChan := make(chan IndexedValue, len(news))

	for ind, item := range news {
		wg.Add(1)

		img := item.Image
		go generatePresignedUrl(img, ind, wg, urlChan)
	}

	wg.Wait()
	close(urlChan)

	for url := range urlChan {
		ind := url.Index
		news[ind].Image = url.Url
	}

	return &news, nil
}
