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

const GetContactUsItems = `
	SELECT id, created_at, data FROM contact_us
	WHERE
	project_id = @projectId
	AND 
	created_at < @createdAt
	ORDER BY created_at DESC
	LIMIT @limit
`

func GetContactUsItemsArgs(projectId, createdAt string, limit int) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
		"createdAt": createdAt,
		"limit":     limit,
	}
}

const DeleteContactUsById = `
	DELETE FROM contact_us
	WHERE
	id = @contactId
`

func DeleteContactUsByIdArgs(contactId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"contactId": contactId,
	}
}
