package dbqueries

import "github.com/jackc/pgx/v5"

const CreatePriseReportItem = `
	INSERT INTO prise-reports 
	(project_id, data, region)
	VALUES
	(@project_id, @data, @region)
`

func CreatePriseReportItemArgs(projectID, region string, data map[string]interface{}) pgx.NamedArgs {
	return pgx.NamedArgs{
		"project_id": projectID,
		"data":       data,
		"region":     region,
	}
}

const GetPriseReportItems = `
	SELECT id, created_at, data, region FROM prise-reports 
	WHERE
	project_id = @projectId
	And region = @region
`

func GetPriseReportItemsArgs(projectId, region string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
		"region":    region,
	}
}

const UpdatePriseReportById = `
	Update prise-reports 
	set data=@data
	WHERE
	id = @reportId
`

func UpdatePriseReportByIdArgs(reportId string, data map[string]interface{}) pgx.NamedArgs {
	return pgx.NamedArgs{
		"reportId": reportId,
		"data":     data,
	}
}

const DeletePriseReportById = `
	DELETE FROM prise-reports 
	WHERE
	id = @reportId
`

func DeletePriseReportByIdArgs(reportId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"reportId": reportId,
	}
}
