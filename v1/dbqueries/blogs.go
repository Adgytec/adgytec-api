package dbqueries

import "github.com/jackc/pgx/v5"

const CreateBlogItem = `
	INSERT INTO blogs 
	(blog_id, user_id, project_id, title, cover_image, short_text, content, author)
	VALUES 
	(@blogId, @userId, @projectId, @title, @cover, @summary, @content, @author)
`

func CreateBlogItemArgs(
	blogId,
	userId,
	projectId,
	title,
	cover,
	summary,
	content,
	author string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"blogId":    blogId,
		"userId":    userId,
		"projectId": projectId,
		"title":     title,
		"cover":     cover,
		"summary":   summary,
		"content":   content,
		"author":    author,
	}
}
