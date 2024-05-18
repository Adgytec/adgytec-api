package dbqueries

import "github.com/jackc/pgx/v5"

const CreateProject = `INSERT INTO project (project_name)
						VALUES (@projectName)`

func CreateProjectArgs(projectName string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectName": projectName,
	}
}
