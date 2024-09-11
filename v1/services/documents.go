package services

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
	"log"
	"net/http"
	"os"
	"time"
)

type DocumentCover struct {
	Id        string    `json:"id" db:"cover_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

func (d *DocumentCover) PostDocumentCoverByProjectId(projectId, userId string) error {
	if d.Name == "" {
		return &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: "Invalid document cover details",
		}
	}

	args := dbqueries.PostDocumentCoverByProjectIdArgs(projectId, d.Name, userId)
	_, err := db.Exec(ctx, dbqueries.PostDocumentCoverByProjectId, args)
	if err != nil {
		log.Printf("Error adding document cover in database: %v\n", err)
		return err
	}

	return nil
}

func deleteDocumentsFromDocumentCover(coverId, projectId string) {
	mediaPrefix := fmt.Sprintf("services/documents/%v/%v/", projectId, coverId)
	if val := os.Getenv("ENV"); val == "dev" {
		mediaPrefix = "dev/" + mediaPrefix
	}
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)

		opts := minio.ListObjectsOptions{
			Recursive: true,
			Prefix:    mediaPrefix,
		}
		// List all objects from a bucket-name with a matching prefix.
		for object := range spaceStorage.ListObjects(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), opts) {
			if object.Err != nil {
				log.Printf("error listing object: %v\n", object.Err)
			} else {
				objectsCh <- object
			}
		}
	}()

	opts := minio.RemoveObjectsOptions{}

	for rErr := range spaceStorage.RemoveObjects(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), objectsCh, opts) {
		fmt.Println("Error detected during deletion: ", rErr)
	}
}

func (d *DocumentCover) DeleteDocumentCoverById(projectId string) error {
	args := dbqueries.DeleteDocumentCoverByIdArgs(d.Id)
	_, err := db.Exec(ctx, dbqueries.DeleteDocumentCoverBytId, args)
	if err != nil {
		log.Printf("Error deleting document cover: %v\n", err)
		return err
	}

	// delete everything in that document cover
	go deleteDocumentsFromDocumentCover(d.Id, projectId)

	return nil
}

func (d *DocumentCover) PatchDocumentCoverById() error {
	args := dbqueries.PatchDocumentCoverByIdArgs(d.Id, d.Name)
	_, err := db.Exec(ctx, dbqueries.PatchDocumentCoverById, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid document cover to update."
				return &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
			}
		}

		log.Printf("Error updating document cover data: %v\n", err)
		return err
	}
	return nil
}

func (d *DocumentCover) GetDocumentCoverByProjectId(projectId, cursor string) (*[]DocumentCover, error) {
	args := dbqueries.GetDocumentCoverByProjectIdArgs(projectId, cursor)
	rows, err := db.Query(ctx, dbqueries.GetDocumentCoverByProjectId, args)

	if err != nil {
		log.Printf("Error fetching document cover from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	documentCovers, err := pgx.CollectRows(rows, pgx.RowToStructByName[DocumentCover])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	return &documentCovers, nil
}
