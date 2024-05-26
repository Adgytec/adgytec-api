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
