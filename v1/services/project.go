package services

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type Project struct {
	ProjectName string    `json:"projectName" db:"project_name"`
	Id          string    `json:"projectId,omitempty" db:"project_id"`
	CreatedAt   time.Time `json:"createdAt,omitempty" db:"created_at"`
}

type ProjectDetail struct {
	Name      string          `json:"projectName" db:"name"`
	CreatedAt time.Time       `json:"createdAt" db:"created_at"`
	Users     json.RawMessage `json:"users" db:"user_data"`
	Services  json.RawMessage `json:"services" db:"service_data"`
	Token     string          `json:"publicToken" db:"token"`
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
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			// unique project name voilation
			if pgErr.Code == "23505" {
				message := "A project with that name already exists."
				return "", &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}

		log.Printf("Error adding project in database: %v\n", err)
		return "", err
	}

	return clientToken, nil
}

func (p *Project) GetAllProjects() (*[]Project, error) {
	rows, err := db.Query(ctx, dbqueries.GetAllProjects)
	if err != nil {
		log.Printf("Error fetching projects from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	projects, err := pgx.CollectRows(rows, pgx.RowToStructByName[Project])
	if err != nil {
		log.Printf("Error reading rows: %v", err)
		return nil, err
	}

	return &projects, err
}

func (p *Project) GetProjectById() (*ProjectDetail, error) {
	args := dbqueries.GetProjectDetailsByIdArgs(p.Id)
	rows, err := db.Query(ctx, dbqueries.GetProjectDetailsById, args)
	if err != nil {
		log.Printf("Error fetching project details from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	project, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ProjectDetail])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message := "Project with the provided ID does not exist."
			return nil, &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
		}
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	return &project, err
}

func (p *Project) DeleteProjectById() error {
	args := dbqueries.DeleteProjectByIdArgs(p.Id)
	_, err := db.Exec(ctx, dbqueries.DeleteProjectById, args)
	if err != nil {
		log.Printf("Error deleting project from db: %v\n", err)
		return err
	}

	return nil
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
				message := "Invalid project id or user id."
				return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}

		log.Printf("Error adding user to project: %v\n", err)
		return err
	}

	return nil
}

func (pu *ProjectUserMap) DeleteUserProjectMap(projectId string) error {
	args := dbqueries.DeleteUserFromProjectArgs(pu.UserId, projectId)
	_, err := db.Exec(ctx, dbqueries.DeleteUserFromProject, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid project id or user id."
				return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}

		log.Printf("Error removing user from project: %v\n", err)
		return err
	}

	return nil
}
