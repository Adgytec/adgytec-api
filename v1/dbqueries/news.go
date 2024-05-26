package dbqueries

import "github.com/jackc/pgx/v5"

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
LIMIT 5`

func GetAllNewsByProjectIdArgs(projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
	}
}
