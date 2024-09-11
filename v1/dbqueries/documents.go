package dbqueries

import "github.com/jackc/pgx/v5"

const PostDocumentCoverByProjectId = `
	INSERT INTO document_cover (project_id, name, user_id)
	VALUES
	(@projectId, @name, @userId);
`

func PostDocumentCoverByProjectIdArgs(projectId, name, userId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
		"name":      name,
		"userId":    userId,
	}
}

const GetDocumentCoverByProjectId = `
	SELECT cover_id, name, created_at
	FROM document_cover
	WHERE 
	projectId = @projectId
	AND 
	created_at < @createdAt
	ORDER BY created_at DESC
	LIMIT 20
`

func GetDocumentCoverByProjectIdArgs(projectId, createdAt string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
		"createdAt": createdAt,
	}
}

const DeleteDocumentCoverBytId = `
	Delete FROM document_cover
	where
	cover_id = @coverId
`

func DeleteDocumentCoverByIdArgs(coverId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"coverId": coverId,
	}
}

const PatchDocumentCoverById = `
	UPDATE document_cover
	SET name = @name
	Where cover_id = @coverId
`

func PatchDocumentCoverByIdArgs(coverId, name string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"coverId": coverId,
		"name":    name,
	}
}
