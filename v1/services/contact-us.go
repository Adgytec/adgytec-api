package services

import (
	"encoding/json"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
	"log"
)

type ContactUs struct {
	Id        string          `json:"id" db:"id"`
	ProjectId string          `json:"projectId" db:"project_id"`
	Data      json.RawMessage `json:"data" db:"data"`
}

func (c *ContactUs) PostContactUs(data map[string]interface{}) error {
	args := dbqueries.CreateContactUsItemArgs(c.ProjectId, data)

	_, err := db.Exec(ctx, dbqueries.CreateContactUsItem, args)
	if err != nil {
		log.Printf("Error adding contact us record to database: %v\n", err)
	}

	return err
}
