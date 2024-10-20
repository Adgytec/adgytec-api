package dbqueries

import "github.com/jackc/pgx/v5"

const CreateContactUsItem = `
	INSERT INTO contact_us 
	(project_id, data)
	VALUES
	(@project_id, @data)
`

func CreateContactUsItemArgs(projectID string, data map[string]interface{}) pgx.NamedArgs {
	return pgx.NamedArgs{
		"project_id": projectID,
		"data":       data,
	}
}
