package services

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type Album struct {
	Id        string    `json:"id" db:"album_id"`
	Name      string    `json:"name" db:"name"`
	Cover     string    `json:"cover" db:"cover"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

func addAlbumToDatabase(a *Album, userId, projectId string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	args := dbqueries.PostAlbumByProjectIdArgs(a.Id, projectId, userId, a.Name, a.Cover)
	_, err := db.Exec(ctx, dbqueries.PostAlbumByProjectId, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23502" {
				message := "Some required values are empty."
				err = &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				errChan <- err
				return
			}

			if pgErr.Code == "23503" {
				message := "Invalid user or project."
				err = &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				errChan <- err
				return
			}
		}

		log.Printf("Error adding album in database: %v\n", err)
	}

	errChan <- err
}

func (a *Album) CreateAlbum(r *http.Request, projectId, userId string) error {
	file, header, err := r.FormFile("cover")
	if err != nil {
		log.Printf("Error retriving file: %v\n ", err)
		return err
	}
	defer file.Close()

	contentType, err := isImageFile(header)
	if err != nil {
		return err
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

	albumId := GenerateUUID().String()
	objectName := fmt.Sprintf("services/gallery/%v/%v/%v.%v", projectId, albumId, generateRandomString(), format)

	if val := os.Getenv("ENV"); val == "dev" {
		objectName = "dev/" + objectName
	}
	a.Cover = objectName
	a.Id = albumId

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)

	wg.Add(2)
	go uploadImageToCloudStorage(objectName, buf, contentType, wg, errChan)
	go addAlbumToDatabase(a, userId, projectId, wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			go deleteFromCloudStorage(objectName)
			// func to delete from db
			return err
		}
	}

	return nil
}
