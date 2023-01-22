-- name: CreateSecret :one
INSERT INTO secrets (
  owner,
  kind,
  name,
  value,
  created,
  modified
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetSecret :one
SELECT * FROM secrets
WHERE owner = $1 AND kind = $2 AND name = $3
LIMIT $1;

-- name: GetSecretsByUser :many
SELECT * FROM secrets
WHERE owner = $1;

-- name: GetSecretsByKind :many
SELECT * FROM secrets
WHERE owner = $1 AND kind = $2;

-- name: MarkSecretDeleted :exec
UPDATE secrets
SET deleted = true
WHERE owner = $1 AND kind = $2 AND name = $3;

-- name: DeleteSecret :exec
DELETE FROM secrets
WHERE owner = $1 AND kind = $2 AND name = $3;
