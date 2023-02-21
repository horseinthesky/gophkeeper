-- name: CreateUser :one
INSERT INTO users (
  name,
  passhash
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE name = $1
LIMIT 1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE name = $1;
