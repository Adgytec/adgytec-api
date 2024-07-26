package services

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type Category struct {
	ParentId     string
	CategoryName string
}

type CategoryDetail struct {
	Categories json.RawMessage `json:"categories" db:"categories"`
}

func (c *Category) PostCategoryByProjectId(projectId string) error {
	if c.ParentId == "" || c.CategoryName == "" {
		return &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: "Invalid category details",
		}
	}

	args := dbqueries.PostCategoryByProjectIdArgs(c.ParentId, projectId, c.CategoryName)
	_, err := db.Exec(ctx, dbqueries.PostCategoryByProjectId, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			// foreign key violation code 23503
			if pgErr.Code == "23503" {
				var message string
				if strings.Contains(pgErr.Detail, "project_id") {
					message = "Project with the given id doesn't exist."
				} else if strings.Contains(pgErr.Detail, "parent_id") {
					message = "Invalid parent for the category."
				}

				return &custom.MalformedRequest{
					Status:  http.StatusBadRequest,
					Message: message,
				}
			}

			if pgErr.Code == "22P02" {
				message := "Invalid category details."
				return &custom.MalformedRequest{
					Status:  http.StatusBadRequest,
					Message: message,
				}
			}
		}

		log.Printf("Error creating new category: %v/n", err)
		return err
	}

	return nil
}

func (c *Category) PatchCategoryById(categoryId string) error {
	if c.CategoryName == "" {
		return &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: "Invalid category details",
		}
	}

	args := dbqueries.PatchCategoryByIdArgs(c.CategoryName, categoryId)
	_, err := db.Exec(ctx, dbqueries.PatchCategoryById, args)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid category details"
				return &custom.MalformedRequest{
					Status:  http.StatusBadRequest,
					Message: message,
				}
			}
		}

		log.Printf("Error updating category detail: %v/n", err)
		return err
	}

	return nil
}

func (c *Category) GetCategoryByProjectId(projectId string) (*CategoryDetail, error) {
	args := dbqueries.GetCategoryByProjectIdArgs(projectId)
	row, err := db.Query(ctx, dbqueries.GetCategoryByProjectId, args)
	if err != nil {
		log.Printf("Error fetching category details from db: %v\n", err)
		return nil, err
	}
	defer row.Close()

	categories, err := pgx.CollectOneRow(row, pgx.RowToStructByName[CategoryDetail])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message := "Project with the provided ID does not exist."
			return nil, &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid project id."
				return nil, &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}

		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	return &categories, err
}

func (c *Category) DeleteCategoryById(categoryId string) error {
	args := dbqueries.DeleteCategoryByIdArgs(categoryId)
	_, err := db.Exec(ctx, dbqueries.DeleteCategoryById, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid category id"
				return &custom.MalformedRequest{
					Status:  http.StatusBadRequest,
					Message: message,
				}
			}
		}

		log.Printf("Error deleting category from db: %v\n", err)
		return err
	}

	return nil
}
