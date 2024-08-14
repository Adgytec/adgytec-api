package dbqueries

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// create project
const CreateProject = `
	WITH inserted_row AS (
		INSERT INTO project (project_name, cover_image, project_id)
		VALUES (@projectName, @coverImage, @projectId)
		RETURNING project_id
	), insert_category as (
		INSERT INTO category (category_id, project_id, category_name)
		VALUES (@projectId, @projectId, 'default')
	)
	INSERT INTO client_token (token, project_id)
	SELECT @clientToken, project_id
	FROM inserted_row;
`

func CreateProjectArgs(projectName, coverImage, projectId, clientToken string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectName": projectName,
		"coverImage":  coverImage,
		"projectId":   projectId,
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
const AddUserToProject = `
	INSERT INTO user_to_project (user_id, project_id)
	VALUES (@userId, @projectId)
`

func AddUserToProjectArgs(userId, projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"userId":    userId,
		"projectId": projectId,
	}
}

// get project id  by token
const GetProjectIdByClientToken = `
	SELECT project_id FROM client_token
	WHERE token=@clientToken
`

func GetProjectIdByClientTokenArgs(clientToken string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"clientToken": clientToken,
	}
}

// auth to check if the user has rights to perform action in that project
const GetProjectIdByUserIdAndProjectId = `
	SELECT project_id FROM user_to_project 
	WHERE user_id=@userId AND project_id=@projectId
`

func GetProjectIdByUserIdAndProjectIdArgs(userId, projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"userId":    userId,
		"projectId": projectId,
	}
}

// to check if project exists or not
const GetProjectNameById = `
	SELECT project_name FROM project 
	WHERE project_id=@projectId
`

const GetProjectById = `
	SELECT * FROM project 
	WHERE project_id=@projectId
`

func GetProjectByIdArgs(projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
	}
}

// get all project
const GetAllProjects = `
	SELECT project_id, project_name, created_at, cover_image FROM project
`

// get project by id
const GetProjectDetailsById = `
	SELECT
		p.project_name as name,
		p.created_at,
		p.cover_image,
		coalesce(ud.user_data, '[]'::json) AS user_data,
		coalesce(s.service_data, '[]'::json) as service_data,
		c.token
	FROM project p
	INNER JOIN (
		SELECT json_agg(json_build_object('userId', up.user_id, 'name', u.name, 'email', u.email)) AS user_data
		FROM user_to_project up
		INNER JOIN users u 
		ON up.user_id = u.user_id
		WHERE up.project_id = @projectId
	) ud ON 1=1 
	INNER JOIN (
		SELECt json_agg(json_build_object('serviceId', sp.service_id, 'serviceName', s.service_name, 'icon', s.icon)) AS service_data
		FROM services s 
		INNER JOIN project_to_service sp 
		ON sp.service_id = s.service_id
		WHERE sp.project_id=@projectId
	) s ON 1=1
	INNER JOIN client_token c 
	ON p.project_id = c.project_id
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
	RETURNING cover_image
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
	SELECT service_name, service_id, icon FROM services
`

// get project by user id
const GetProjectByUserId = `
	SELECT p.project_name, p.project_id, p.created_at, p.cover_image
	FROM project p
	INNER JOIN user_to_project up
	ON p.project_id = up.project_id
	WHERE up.user_id = @userId
`

func GetProjectByUserIdArgs(userId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"userId": userId,
	}
}

// get services by project id
const GetMetadataByProjectId = `
	with recursive elt as (
	SELECT 
    p.category_id AS category,
    CASE 
        WHEN array_agg(c.category_id) = ARRAY[NULL::uuid] THEN NULL 
        ELSE array_agg(c.category_id) 
    END AS sub_categories,
    jsonb_build_object(
        'categoryId', p.category_id,
        'categoryName', p.category_name,
        'subCategories', array_remove(
            array_agg(
                CASE
                    WHEN c.category_id IS NOT NULL AND c.category_name IS NOT NULL THEN
                        jsonb_build_object(
                            'categoryId', c.category_id,
                            'categoryName', c.category_name
                        )
                    ELSE NULL
                END
            ), 
            NULL
        )
    ) AS json_tree
FROM 
    category AS p
LEFT JOIN 
    category AS c ON p.category_id = c.parent_id
where p.project_id = @projectId
GROUP BY 
    p.category_id, p.category_name

),
list (category, json_tree, sub_categories, rank, path) AS
(
SELECT c.category, c.json_tree, c.sub_categories, 1, '{}' :: text[]
  FROM elt AS c
  LEFT JOIN elt AS f
    ON array[c.category] <@ f.sub_categories
 WHERE f.category IS NULL
UNION ALL
SELECT c.category
     , c.json_tree
     , c.sub_categories
     , f.rank + 1
     , f.path || array['subCategories',(array_position(f.sub_categories, c.category)-1) :: text]
  FROM list AS f
 INNER JOIN elt AS c
    ON array[c.category] <@ f.sub_categories
)
SELECT 
p.project_name, coalesce(sd.service_data, '[]'::jsonb) AS services_data, coalesce(jsonb_set_agg(NULL :: jsonb, path, l.json_tree, true ORDER BY rank ASC), '[]'::jsonb) as categories_data
FROM project p
inner join list l on 1 =1
inner join (
	SELECT jsonb_agg(jsonb_build_object('id', s.service_id, 'name', s.service_name, 'icon', s.icon)) AS service_data
	FROM services s
	INNER JOIN project_to_service ps
	ON s.service_id = ps.service_id
	WHERE ps.project_id = @projectId
) sd on 1=1
WHERE p.project_id = @projectId
group by p.project_name, sd.service_data


`

func GetMetadataByProjectIdArgs(projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
	}
}
