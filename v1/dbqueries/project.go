package dbqueries

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// create project
// const CreateProject = `INSERT INTO project (project_name)
//
//	VALUES (@projectName)`
const CreateProject = `WITH inserted_row AS (
    INSERT INTO project (project_name)
    VALUES (@projectName)
    RETURNING project_id
)
INSERT INTO client_token (token, project_id)
SELECT @clientToken, project_id
FROM inserted_row;`

func CreateProjectArgs(projectName, clientToken string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectName": projectName,
		"clientToken": clientToken,
	}
}

// add services to project
func AddServicesToProject(projectId string, services []string) string {
	query := "INSERT INTO project_to_service (project_id, service_id) VALUES "
	var values []string

	for _, id := range services {
		value := fmt.Sprintf("('%v', '%v')", projectId, id)
		values = append(values, value)
	}

	query += strings.Join(values, ", ")

	return query
}

// add a user to project
const AddUserToProject = `INSERT INTO user_to_project (user_id, project_id)
	VALUES (@userId, @projectId)
`

func AddUserToProjectArgs(userId, projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"userId":    userId,
		"projectId": projectId,
	}
}

// get project id  by token
const GetProjectIdByClientToken = `SELECT project_id FROM client_token WHERE token=@clientToken`

func GetProjectIdByClientTokenArgs(clientToken string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"clientToken": clientToken,
	}
}

// auth to check if the user has rights to perform action in that project
const GetProjectIdByUserIdAndProjectId = `SELECT project_id FROM user_to_project WHERE user_id=@userId AND project_id=@projectId`

func GetProjectIdByUserIdAndProjectIdArgs(userId, projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"userId":    userId,
		"projectId": projectId,
	}
}

// to check if project exists or not
const GetProjectById = `SELECT * FROM project WHERE project_id=@projectId`

func GetProjectByIdArgs(projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
	}
}

// get all project
const GetAllProjects = `SELECT * FROM project`

// get project by id
const GetProjectDetailsById = `
SELECT
  p.project_name as name,
  p.created_at,
  coalesce(ud.user_data, '[]'::json) AS user_data,
  coalesce(s.service_data, '[]'::json) as service_data,
  c.token
FROM project p
INNER JOIN (
  SELECT json_agg(json_build_object('id', up.user_id, 'name', u.name, 'email', u.email)) AS user_data
  FROM user_to_project up
  INNER JOIN users u ON up.user_id = u.user_id
  WHERE up.project_id = @projectId
) ud ON 1=1 
INNER JOIN (
	SELECt json_agg(json_build_object('id', sp.service_id, 'name', s.service_name)) AS service_data
	FROM services s 
	INNER JOIN project_to_service sp ON sp.service_id = s.service_id
	WHERE sp.project_id=@projectId
) s ON 1=1
INNER JOIN client_token c ON p.project_id = c.project_id
WHERE p.project_id = @projectId;
`

func GetProjectDetailsByIdArgs(projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
	}
}

// delete project
const DeleteProjectById = `
DELETE FROM project 
WHERE project_id = @projectId
`

func DeleteProjectByIdArgs(projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
	}
}

// remove user project map
const DeleteUserFromProject = `
DELETE FROM user_to_project 
WHERE user_id = @userId AND project_id = @projectId
`

func DeleteUserFromProjectArgs(userId, projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"userId":    userId,
		"projectId": projectId,
	}
}

// remove service project map
const DeleteServiceFromProject = `
DELETE FROM project_to_service 
WHERE service_id = @serviceId AND project_id = @projectId
`

func DeleteServiceFromProjectArgs(serviceId, projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"serviceId": serviceId,
		"projectId": projectId,
	}
}

// get all services
const GetAllServices = `
SELECT service_name, service_id FROM services
`
