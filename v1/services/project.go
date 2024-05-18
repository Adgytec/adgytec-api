package services

import (
	"log"

	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type Project struct {
	ProjectName string `json:"projectName" db:"project_name"`
}

func (p *Project) CreateProject() error {
	args := dbqueries.CreateProjectArgs(p.ProjectName)
	_, err := db.Exec(ctx, dbqueries.CreateProject, args)
	if err != nil {
		log.Printf("Error adding project in database: %v\n", err)
		return err
	}

	return nil
}
