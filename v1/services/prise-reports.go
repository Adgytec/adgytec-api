package services

import (
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rohan031/adgytec-api/v1/dbqueries"
)

type PriseReports struct {
	Id        string          `json:"id" db:"id"`
	Data      json.RawMessage `json:"data" db:"data"`
	CreatedAt time.Time       `json:"createdAt" db:"created_at"`
	Region    string          `json:"region" db:"region"`
}

func (c *PriseReports) PostPriseReports(projectId, region string, data map[string]interface{}) error {
	args := dbqueries.CreatePriseReportItemArgs(projectId, region, data)

	_, err := db.Exec(ctx, dbqueries.CreatePriseReportItem, args)
	if err != nil {
		log.Printf("Error adding prise report record to database: %v\n", err)
	}

	return err
}

func (c *PriseReports) GetPriseReports(projectId, region string) (*[]PriseReports, error) {
	args := dbqueries.GetPriseReportItemsArgs(projectId, region)
	rows, err := db.Query(ctx, dbqueries.GetPriseReportItems, args)

	if err != nil {
		log.Printf("Error fetching prise report items from db: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[PriseReports])
	if err != nil {
		log.Printf("Error reading rows: %v\n", err)
		return nil, err
	}

	return &items, nil
}

func (c *PriseReports) UpdatePriseReportById(data map[string]interface{}) error {
	args := dbqueries.UpdatePriseReportByIdArgs(c.Id, data)

	_, err := db.Exec(ctx, dbqueries.UpdatePriseReportById, args)
	if err != nil {
		log.Printf("Error updating prise report record from db: %v\n", err)
		return err
	}

	return nil
}

func (c *PriseReports) DeletePriseReportById() error {
	args := dbqueries.DeletePriseReportByIdArgs(c.Id)

	_, err := db.Exec(ctx, dbqueries.DeletePriseReportById, args)
	if err != nil {
		log.Printf("Error deleting prise report record from db: %v\n", err)
		return err
	}

	return nil
}
