package services

import (
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
	"log"
	"time"
)

type ContactUs struct {
	Id        string          `json:"id" db:"id"`
	Data      json.RawMessage `json:"data" db:"data"`
	CreatedAt time.Time       `json:"createdAt" db:"created_at"`
}

func (c *ContactUs) PostContactUs(projectId string, data map[string]interface{}) error {
	args := dbqueries.CreateContactUsItemArgs(projectId, data)

	_, err := db.Exec(ctx, dbqueries.CreateContactUsItem, args)
	if err != nil {
		log.Printf("Error adding contact us record to database: %v\n", err)
	}

	return err
}

func (c *ContactUs) GetContactUs(projectId, cursor string) (*[]ContactUs, error) {
	args := dbqueries.GetContactUsItemsArgs(projectId, cursor)
	rows, err := db.Query(ctx, dbqueries.GetContactUsItems, args)

	if err != nil {
		log.Printf("Error fetching contact us items from db: %v\n", err)
		return nil, err
	}

	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[ContactUs])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	return &items, nil
}

func (c *ContactUs) DeleteContactUsById() error {
	args := dbqueries.DeleteContactUsByIdArgs(c.Id)

	_, err := db.Exec(ctx, dbqueries.DeleteContactUsById, args)
	if err != nil {
		log.Printf("Error deleting contact us record from db: %v\n", err)
		return err
	}

	return nil
}
