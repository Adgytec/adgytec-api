package dbqueries

import "github.com/jackc/pgx/v5"

const PostCategoryByProjectId = `
	INSERT INTO category (parent_id, project_id, category_name)
	VALUES (@parentId, @projectId, @categoryName)
	RETURNING category_id
`

func PostCategoryByProjectIdArgs(parentId, projectId, categoryName string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"parentId":     parentId,
		"projectId":    projectId,
		"categoryName": categoryName,
	}
}

const PatchCategoryById = `
	UPDATE category
	SET category_name = @categoryName
	WHERE category_id = @categoryId
`

func PatchCategoryByIdArgs(categoryName, categoryId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"categoryName": categoryName,
		"categoryId":   categoryId,
	}
}

const GetCategoryByProjectId = `
	WITH RECURSIVE elt AS (
		SELECT 
			p.category_id AS category,
			CASE 
				WHEN ARRAY_AGG(c.category_id) = ARRAY[NULL::uuid] THEN NULL 
				ELSE ARRAY_AGG(c.category_id) 
			END AS sub_categories,
			jsonb_build_object(
				'categoryId', p.category_id,
				'categoryName', p.category_name,
				'subCategories', ARRAY_REMOVE(
					ARRAY_AGG(
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
		WHERE p.project_id = @projectId
		GROUP BY 
			p.category_id, p.category_name
	),
	list (category, json_tree, sub_categories, rank, path) AS (
		SELECT c.category, c.json_tree, c.sub_categories, 1, '{}'::TEXT[]
		FROM elt AS c
		LEFT JOIN elt AS f
			ON ARRAY[c.category] <@ f.sub_categories
		WHERE f.category IS NULL
		UNION ALL
		SELECT c.category
			, c.json_tree
			, c.sub_categories
			, f.rank + 1
			, f.path || ARRAY['subCategories',(ARRAY_POSITION(f.sub_categories, c.category)-1)::TEXT]
		FROM list AS f
		INNER JOIN elt AS c
			ON ARRAY[c.category] <@ f.sub_categories
	)
	SELECT 
	COALESCE(JSONB_SET_AGG(NULL::JSONB, path, json_tree, TRUE ORDER BY rank ASC), '[]'::JSONB) AS categories
	FROM list
`

func GetCategoryByProjectIdArgs(projectId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"projectId": projectId,
	}
}

const DeleteCategoryById = `
	DELETE FROM category 
	WHERE category_id = @categoryId
`

func DeleteCategoryByIdArgs(categoryId string) pgx.NamedArgs {
	return pgx.NamedArgs{
		"categoryId": categoryId,
	}
}
