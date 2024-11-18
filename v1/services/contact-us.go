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

func (c *ContactUs) GetContactUs(projectId, cursor string, limit int) (*[]ContactUs, *PageInfo, error) {
	args := dbqueries.GetContactUsItemsArgs(projectId, cursor, limit+1)
	rows, err := db.Query(ctx, dbqueries.GetContactUsItems, args)

	if err != nil {
		log.Printf("Error fetching contact us items from db: %v\n", err)
		return nil, nil, err
	}

	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[ContactUs])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, nil, err
	}

	var pageInfo PageInfo = PageInfo{
		NextPage: false,
		Cursor:   nil,
	}

	if len(items) > limit {
		items = items[:len(items)-1]
		pageInfo.NextPage = true
		pageInfo.Cursor = &items[len(items)-1].CreatedAt
	}

	return &items, &pageInfo, nil
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
