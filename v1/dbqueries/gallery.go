package dbqueries

import (
	"github.com/jackc/pgx/v5"
)

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

// photos
const PostPhotoByAlbumId = `
	INSERT INTO photos (photo_id, album_id, path, user_id)
	VALUES
	(@photoId, @albumId, @path, @userId)
`

func PostPhotoByAlbumIdArgs(photoId, albumId, path, userId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"photoId": photoId,
		"albumId": albumId,
		"path":    path,
		"userId":  userId,
	}
}

const GetPhotosByAlbumId = `
	SELECT photo_id, path, created_at
	FROM photos
	WHERE
	album_id = @albumId
	AND created_at < @createdAt
	ORDER BY created_at DESC
	LIMIT 20
`

func GetPhotosByAlbumIdArgs(albumId, createdAt string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"albumId":   albumId,
		"createdAt": createdAt,
	}
}

const DeletePhotosById = `
	DELETE FROM photos
	WHERE 
	photo_id = ANY(@photoIds)
	RETURNING path
`

func DeletePhotosByIdArgs(photoIds []string) pgx.NamedArgs {

	return pgx.NamedArgs{
		"photoIds": photoIds,
	}
}
