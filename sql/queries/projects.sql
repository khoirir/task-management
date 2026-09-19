-- name: CreateProject :one
INSERT INTO projects (
    name, description, owner_id
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects
WHERE id = $1 LIMIT 1;

-- name: ListProjectsByOwner :many
SELECT * FROM projects
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    updated_at = NOW()
WHERE id = $1 AND owner_id = $4
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1 AND owner_id = $2;
