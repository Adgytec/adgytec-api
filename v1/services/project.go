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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type Project struct {
	ProjectName string    `json:"projectName" db:"project_name"`
	Id          string    `json:"projectId,omitempty" db:"project_id"`
	CreatedAt   time.Time `json:"createdAt,omitempty" db:"created_at"`
	Cover       string    `json:"cover" db:"cover_image"`
}

type ProjectDetail struct {
	Name      string          `json:"projectName" db:"name"`
	CreatedAt time.Time       `json:"createdAt" db:"created_at"`
	Users     json.RawMessage `json:"users" db:"user_data"`
	Services  json.RawMessage `json:"services" db:"service_data"`
	Token     string          `json:"publicToken" db:"token"`
	Cover     string          `json:"cover" db:"cover_image"`
}

type ProjectUserMap struct {
	UserId string `json:"userId"`
}

type ProjectServiceMap struct {
	Services []string `json:"services"`
}

type ServicesDetails struct {
	Name string `json:"serviceName" db:"service_name"`
	Id   string `json:"serviceId" db:"service_id"`
	Icon string `json:"icon" db:"icon"`
}

type MetaDataByProject struct {
	Name       string          `json:"projectName" db:"project_name"`
	Services   json.RawMessage `json:"services" db:"services_data"`
	Categories json.RawMessage `json:"categories" db:"categories_data"`
}

type ProjectImage struct {
	Cover string `db:"cover_image"`
}

func addProjectToDatabase(p *Project, clientToken string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	args := dbqueries.CreateProjectArgs(p.ProjectName, p.Cover, p.Id, clientToken)
	_, err := db.Exec(ctx, dbqueries.CreateProject, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			// unique project name voilation
			if pgErr.Code == "23505" {
				message := "A project with that name already exists."
				err = &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
				errChan <- err
				return
			}
		}

		log.Printf("Error adding project in database: %v\n", err)
	}

	errChan <- err
}

// admin only
func (p *Project) CreateProject(r *http.Request) error {
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

	projectId := GenerateUUID().String()

	objectName := fmt.Sprintf("projects/%v/cover.%v", projectId, format)

	if val := os.Getenv("ENV"); val == "dev" {
		objectName = "dev/" + objectName
	}

	clientToken, err := generateSecureToken()
	if err != nil {
		return err
	}

	p.Cover = objectName
	p.Id = projectId

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)

	wg.Add(2)

	if contentType == webp {
		go uploadImageToCloudStorage(objectName, file, header.Size, contentType, wg, errChan)
	} else {
		go uploadImageToCloudStorage(objectName, buf, int64(buf.Len()), contentType, wg, errChan)
	}
	go addProjectToDatabase(p, clientToken, wg, errChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			go deleteFromCloudStorage(objectName)
			go p.DeleteProjectById()
			return err
		}
	}

	return nil
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

	wg := new(sync.WaitGroup)
	urlChan := make(chan IndexedValue, len(projects))

	for ind, item := range projects {
		wg.Add(1)

		img := item.Cover
		go generatePresignedUrl(img, ind, expires, wg, urlChan)
	}

	wg.Wait()
	close(urlChan)

	for url := range urlChan {
		ind := url.Index

		projects[ind].Cover = url.Url
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

	wg := new(sync.WaitGroup)
	urlChan := make(chan IndexedValue, 1)

	wg.Add(1)

	img := project.Cover
	go generatePresignedUrl(img, 1, expires, wg, urlChan)

	wg.Wait()
	close(urlChan)

	for url := range urlChan {
		project.Cover = url.Url
	}

	return &project, err
}

func (p *Project) DeleteProjectById() error {
	args := dbqueries.DeleteProjectByIdArgs(p.Id)
	rows, err := db.Query(ctx, dbqueries.DeleteProjectById, args)
	if err != nil {
		log.Printf("Error deleting project from db: %v\n", err)
		return err
	}
	defer rows.Close()

	project, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ProjectImage])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message := "project not found"
			return &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}

		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// foreign key violation code 23503
			if pgErr.Code == "23503" {
				message := "You need to delete all the services data inorder to delete the project."
				return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}

			}
		}

		log.Printf("Error reading rows: %v\n", err)
		return err
	}

	// delete from space storage
	// err = spaceStorage.RemoveObject(ctx, os.Getenv("SPACE_STORAGE_BUCKET_NAME"), project.Cover, minio.RemoveObjectOptions{})
	// if err != nil {
	// 	log.Printf("Error deleting image from space storage: %v\n", err)
	// 	// return err
	// }
	go deleteFromCloudStorage(project.Cover)

	return nil
}

func (p *Project) GetAllServices() (*[]ServicesDetails, error) {
	rows, err := db.Query(ctx, dbqueries.GetAllServices)
	if err != nil {
		log.Printf("Error fetching services from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	services, err := pgx.CollectRows(rows, pgx.RowToStructByName[ServicesDetails])
	if err != nil {
		log.Printf("Error reading rows: %v", err)
		return nil, err
	}

	return &services, err
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

			// composite key violation
			if pgErr.Code == "23505" {
				message := "The selected services are already included in this project."
				return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}

		log.Printf("Error adding services to project: %v\n", err)
		return err
	}

	return nil
}

func (ps *ProjectServiceMap) DeleteProjectServiceMap(projectId string) error {
	args := dbqueries.DeleteServiceFromProjectArgs(ps.Services[0], projectId)
	_, err := db.Exec(ctx, dbqueries.DeleteServiceFromProject, args)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid project id or service id."
				return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}

		log.Printf("Error removing service from project: %v\n", err)
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

			// composite key violation
			if pgErr.Code == "23505" {
				message := "It looks like this user is already associated with the project."
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

// admin and user
func (p *Project) GetProjectsByUserId(userId string) (*[]Project, error) {
	args := dbqueries.GetProjectByUserIdArgs(userId)
	rows, err := db.Query(ctx, dbqueries.GetProjectByUserId, args)
	if err != nil {
		log.Printf("Error fetching projects from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	projects, err := pgx.CollectRows(rows, pgx.RowToStructByName[Project])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	wg := new(sync.WaitGroup)
	urlChan := make(chan IndexedValue, len(projects))

	for ind, item := range projects {
		wg.Add(1)

		img := item.Cover
		go generatePresignedUrl(img, ind, expires, wg, urlChan)
	}

	wg.Wait()
	close(urlChan)

	for url := range urlChan {
		ind := url.Index

		projects[ind].Cover = url.Url
	}

	return &projects, err
}

func (p *Project) GetMetadataByProjectId() (*MetaDataByProject, error) {
	args := dbqueries.GetMetadataByProjectIdArgs(p.Id)
	rows, err := db.Query(ctx, dbqueries.GetMetadataByProjectId, args)
	if err != nil {
		log.Printf("Error fetching project details from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	project, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[MetaDataByProject])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			message := "Project with the provided ID does not exist."
			return nil, &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				message := "Invalid project id"
				return nil, &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			}
		}

		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	return &project, err
}
