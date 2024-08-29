package dbqueries

import "github.com/jackc/pgx/v5"

const PostAlbumByProjectId = `
	INSERT INTO album (album_id, project_id, name, cover, user_id)
	VALUES
	(@albumId, @projectId, @name, @cover, @userId);
`

func PostAlbumByProjectIdArgs(albumId, projectId, userId, name, cover string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"albumId":   albumId,
		"projectId": projectId,
		"name":      name,
		"cover":     cover,
		"userId":    userId,
	}
}

const GetAlbumsByProjectId = `
	SELECT album_id, name, cover, created_at 
	FROM album
	WHERE 
	project_id = @projectId
	AND
	created_at < @createdAt
	ORDER BY created_at DESC
	LIMIT 20
`

func GetAlbumsByProjectIdArgs(projectId, createdAt string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
		"createdAt": createdAt,
	}
}

const DeleteAlbumById = `
	DELETE FROM album
	WHERE
	album_id = @albumId
`

func DeleteAlbumByIdArgs(albumId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"albumId": albumId,
	}
}

const PatchAlbumMetadataById = `
	UPDATE album
	SET name = @name
	WHERE
	album_id = @albumId
`

func PatchAlbumMetadataByIdArgs(albumId, name string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"albumId": albumId,
		"name":    name,
	}
}

const PatchAlbumCoverById = `
	WITH cover AS (
		SELECT cover as image
		FROM album 
		WHERE album_id = @albumId
	)
	UPDATE album
	SET cover  = @cover
	WHERE album_id = @albumId
	RETURNING (
		SELECT image FROM cover
	)
`

func PatchAlbumCoverByIdArgs(albumId, cover string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"albumId": albumId,
		"cover":   cover,
	}
}
