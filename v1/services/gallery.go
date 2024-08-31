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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type Album struct {
	Id        string    `json:"id" db:"album_id"`
	Name      string    `json:"name" db:"name"`
	Cover     string    `json:"cover" db:"cover"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type Photos struct {
	Id        string    `json:"id" db:"photo_id"`
	Path      string    `json:"image" db:"path"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type PhotosPath struct {
	Path string `db:"path"`
}

type PhotoDelete struct {
	Id []string
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

	var format string
	var img image.Image
	buf := new(bytes.Buffer)

	if contentType == webp {
		log.Println("webp image")
		format = "webp"
	} else {
		img, format, err = image.Decode(file)
		if err != nil {
			log.Printf("Error decoding image: %v\n", err)
			return err
		}

		err = handleImage(img, buf, format)
		if err != nil {
			return err
		}
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
	if contentType == webp {
		go uploadImageToCloudStorage(objectName, file, header.Size, contentType, wg, errChan)
	} else {
		go uploadImageToCloudStorage(objectName, buf, int64(buf.Len()), contentType, wg, errChan)
	}
	go addAlbumToDatabase(a, userId, projectId, wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			go deleteFromCloudStorage(objectName)
			go a.DeleteAlbumById(projectId)
			return err
		}
	}

	return nil
}

func deleteImagesFromAlbum(albumId, projectId string) {
	objectsCh := make(chan minio.ObjectInfo)
	objectName := fmt.Sprintf("services/gallery/%v/%v/", projectId, albumId)
	if val := os.Getenv("ENV"); val == "dev" {
		objectName = "dev/" + objectName
	}

	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range spaceStorage.ListObjects(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), minio.ListObjectsOptions{
			Prefix:    objectName,
			Recursive: true,
		}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			objectsCh <- object
		}
	}()

	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	for rErr := range spaceStorage.RemoveObjects(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectsCh, opts) {
		fmt.Println("Error detected during deletion: ", rErr)
	}

}

func (a *Album) DeleteAlbumById(projectId string) error {
	args := dbqueries.DeleteAlbumByIdArgs(a.Id)
	_, err := db.Exec(ctx, dbqueries.DeleteAlbumById, args)
	if err != nil {
		log.Printf("Error deleting album from db: %v\n", err)
		return err
	}

	// delete everything in that album
	go deleteImagesFromAlbum(a.Id, projectId)

	return nil
}

func (a *Album) PatchAlbumMetadataById() error {
	args := dbqueries.PatchAlbumMetadataByIdArgs(a.Id, a.Name)
	_, err := db.Exec(ctx, dbqueries.PatchAlbumMetadataById, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid album to update."
				return &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
			}
		}

		log.Printf("Error updating blog data: %v\n", err)
		return err
	}
	return nil
}

func handleAlbumCoverDatabase(cover, albumId string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	args := dbqueries.PatchAlbumCoverByIdArgs(albumId, cover)
	rows, err := db.Query(ctx, dbqueries.PatchAlbumCoverById, args)
	if err != nil {
		log.Printf("error updating cover image in db: %v\n", err)
		errChan <- err
		return
	}
	defer rows.Close()

	prevPath, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[struct {
		Image string `db:"image"`
	}])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message := "album with the following id doesn't exist"
			errChan <- &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid album id."
				errChan <- &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				return
			}
		}

		log.Printf("Error reading rows: %v\n", err)
		errChan <- nil
		return
	}

	go deleteFromCloudStorage(prevPath.Image)

	errChan <- nil
}

func (a *Album) PatchAlbumCoverById(r *http.Request, projectId string) error {
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

	var format string
	var img image.Image
	buf := new(bytes.Buffer)

	if contentType == webp {
		log.Println("webp image")
		format = "webp"
	} else {
		img, format, err = image.Decode(file)
		if err != nil {
			log.Printf("Error decoding image: %v\n", err)
			return err
		}

		err = handleImage(img, buf, format)
		if err != nil {
			return err
		}
	}

	objectName := fmt.Sprintf("services/gallery/%v/%v/%v.%v", projectId, a.Id, generateRandomString(), format)

	if val := os.Getenv("ENV"); val == "dev" {
		objectName = "dev/" + objectName
	}
	a.Cover = objectName

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)

	wg.Add(2)

	if contentType == webp {
		go uploadImageToCloudStorage(objectName, file, header.Size, contentType, wg, errChan)
	} else {
		go uploadImageToCloudStorage(objectName, buf, int64(buf.Len()), contentType, wg, errChan)
	}
	go handleAlbumCoverDatabase(objectName, a.Id, wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Album) GetAlbumsByProjectId(projectId, cursor string) (*[]Album, error) {
	args := dbqueries.GetAlbumsByProjectIdArgs(projectId, cursor)
	rows, err := db.Query(ctx, dbqueries.GetAlbumsByProjectId, args)

	if err != nil {
		log.Printf("Error fetching albums from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	albums, err := pgx.CollectRows(rows, pgx.RowToStructByName[Album])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	wg := new(sync.WaitGroup)
	urlChan := make(chan IndexedValue, len(albums))

	for ind, item := range albums {
		wg.Add(1)

		img := item.Cover
		go generatePresignedUrl(img, ind, expires, wg, urlChan)
	}

	wg.Wait()
	close(urlChan)

	for url := range urlChan {
		ind := url.Index
		albums[ind].Cover = url.Url
	}

	return &albums, nil
}

// photos

func addPhotoToDatabase(p *Photos, userId, albumId string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	args := dbqueries.PostPhotoByAlbumIdArgs(p.Id, albumId, p.Path, userId)
	_, err := db.Exec(ctx, dbqueries.PostPhotoByAlbumId, args)
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
				message := "Invalid user or album."
				err = &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				errChan <- err
				return
			}
		}

		log.Printf("Error adding photo in database: %v\n", err)
	}

	errChan <- err

}

func (p *Photos) PostPhotoByAlbumId(r *http.Request, projectId, albumId, userId string) (string, error) {
	photoId := GenerateUUID().String()

	file, header, err := r.FormFile("photo")
	if err != nil {
		log.Printf("Error retriving file: %v\n ", err)
		return "", err
	}
	defer file.Close()

	contentType, err := isImageFile(header)
	if err != nil {
		return "", err
	}

	var format string
	var img image.Image
	buf := new(bytes.Buffer)

	if contentType == webp {
		log.Println("webp image")
		format = "webp"
	} else {
		img, format, err = image.Decode(file)
		if err != nil {
			log.Printf("Error decoding image: %v\n", err)
			return "", err
		}

		err = handleImage(img, buf, format)
		if err != nil {
			return "", err
		}
	}

	objectName := fmt.Sprintf("services/gallery/%v/%v/photos/%v.%v", projectId, albumId, photoId, format)

	if val := os.Getenv("ENV"); val == "dev" {
		objectName = "dev/" + objectName
	}
	p.Path = objectName
	p.Id = photoId

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)

	wg.Add(2)
	if contentType == webp {
		go uploadImageToCloudStorage(objectName, file, header.Size, contentType, wg, errChan)
	} else {
		go uploadImageToCloudStorage(objectName, buf, int64(buf.Len()), contentType, wg, errChan)
	}
	go addPhotoToDatabase(p, userId, albumId, wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			go deleteFromCloudStorage(objectName)
			go p.DeletePhotoById([]string{p.Id})
			return "", err
		}
	}

	return photoId, nil
}

func (p *Photos) DeletePhotoById(photoId []string) error {
	args := dbqueries.DeletePhotosByIdArgs(photoId)
	rows, err := db.Query(ctx, dbqueries.DeletePhotosById, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid photo ids."
				return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}
		log.Printf("Error deleting photos from db: %v\n", err)
		return err
	}
	defer rows.Close()

	photos, err := pgx.CollectRows(rows, pgx.RowToStructByName[PhotosPath])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return err
	}

	if len(photos) == 0 {
		return &custom.MalformedRequest{Status: http.StatusNotFound, Message: "Photos not found"}
	}

	objectChan := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectChan)
		for _, img := range photos {
			objectChan <- minio.ObjectInfo{Key: img.Path}
		}
	}()
	e := spaceStorage.RemoveObjects(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectChan, minio.RemoveObjectsOptions{})

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

func (p *Photos) GetPhotosByAlbumId(albumId, cursor string) (*[]Photos, error) {
	args := dbqueries.GetPhotosByAlbumIdArgs(albumId, cursor)
	rows, err := db.Query(ctx, dbqueries.GetPhotosByAlbumId, args)

	if err != nil {
		log.Printf("Error fetching photos from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	photos, err := pgx.CollectRows(rows, pgx.RowToStructByName[Photos])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	wg := new(sync.WaitGroup)
	urlChan := make(chan IndexedValue, len(photos))

	for ind, item := range photos {
		wg.Add(1)

		img := item.Path
		go generatePresignedUrl(img, ind, expires, wg, urlChan)
	}

	wg.Wait()
	close(urlChan)

	for url := range urlChan {
		ind := url.Index
		photos[ind].Path = url.Url
	}

	return &photos, nil
}
