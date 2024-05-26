package services

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type Project struct {
	ProjectName string `json:"projectName" db:"project_name"`
}

type ProjectUserMap struct {
	UserId string `json:"userId"`
}

type ProjectServiceMap struct {
	Services []string `json:"services"`
}

func (p *Project) CreateProject() (string, error) {
	clientToken, err := generateSecureToken()
	if err != nil {
		return "", err
	}

	args := dbqueries.CreateProjectArgs(p.ProjectName, clientToken)
	_, err = db.Exec(ctx, dbqueries.CreateProject, args)
	if err != nil {
		log.Printf("Error adding project in database: %v\n", err)
		return "", err
	}

	return clientToken, nil
}

func (ps *ProjectServiceMap) CreateProjectServiceMap(projectId string) error {
	query := dbqueries.AddServicesToProject(projectId, ps.Services)
	_, err := db.Exec(ctx, query)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// foreign key violation code 23503
			if pgErr.Code == "23503" {
				if strings.Contains(pgErr.Detail, "project_id") {
					message := "Project id doesn't exist."
					return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				} else if strings.Contains(pgErr.Detail, "service_id") {
					message := "Requested service doesn't exist."
					return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				}
			}

			if pgErr.Code == "22P02" {
				message := "Invalid project id or service."
				return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}

		log.Printf("Error adding services to project: %v\n", err)
		return err
	}

	return nil
}

func (pu *ProjectUserMap) CreateUserProjectMap(projectId string) error {
	args := dbqueries.AddUserToProjectArgs(pu.UserId, projectId)
	_, err := db.Exec(ctx, dbqueries.AddUserToProject, args)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				if strings.Contains(pgErr.Detail, "project_id") {
					message := "Project id doesn't exist."
					return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				} else if strings.Contains(pgErr.Detail, "user_id") {
					message := "User doesn't exist."
					return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				}
			}

			if pgErr.Code == "22P02" {
				message := "Invalid project id."
				return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}

		log.Printf("Error adding services to project: %v\n", err)
		return err
	}

	return nil
}
