package dbqueries

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// create news item
const CreateNewsItem = `INSERT INTO news (title, link, text, image, project_id)
VALUES (@title, @link, @text, @image, @projectId)`

func CreateNewsItemArgs(title, link, text, image, projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"title":     title,
		"link":      link,
		"text":      text,
		"image":     image,
		"projectId": projectId,
	}
}

// get all news
const GetAllNewsByProjectId = `SELECT news_id, title, link, text, image, created_at FROM news
WHERE project_id=@projectId
ORDER BY created_at DESC
LIMIT @limit`

func GetAllNewsByProjectIdArgs(projectId string, limit int) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
		"limit":     limit,
	}
}

// get image by news id
const GetNewsImageById = `SELECT image FROM news WHERE news_id=@newsId`

func GetNewsImageByIdArgs(newsId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"newsId": newsId,
	}
}

// delete news
const DeleteNewsById = `DELETE FROM news WHERE news_id=@newsId RETURNING image`

func DeleteNewsByIdArgs(newsId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"newsId": newsId,
	}
}

func DeleteNewsByProjectId(projectId string) string {
	query := fmt.Sprintf("DELETE FROM news WHERE project_id='%v' RETURNING image", projectId)
	return query
}

func DeleteMultipleNewsById(newsId []string) string {
	for i, val := range newsId {
		newsId[i] = "'" + val + "'"
	}
	newsIdString := strings.Join(newsId, ", ")

	query := fmt.Sprintf("DELETE FROM news WHERE news_id IN (%v) RETURNING image", newsIdString)

	return query
}
