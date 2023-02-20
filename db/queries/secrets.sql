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
LIMIT 1;

-- name: GetSecretsByUser :many
SELECT * FROM secrets
WHERE owner = $1
ORDER BY modified DESC;

-- name: GetSecretsByKind :many
SELECT * FROM secrets
WHERE owner = $1 AND kind = $2
ORDER BY modified DESC;

-- name: UpdateSecret :one
UPDATE secrets
  set value = $4,
  created = $5,
  modified = $6
WHERE owner = $1 AND kind = $2 AND name = $3
RETURNING *;

-- name: MarkSecretDeleted :exec
UPDATE secrets
SET deleted = true
WHERE owner = $1 AND kind = $2 AND name = $3;

-- name: DeleteSecret :exec
DELETE FROM secrets
WHERE owner = $1 AND kind = $2 AND name = $3;

-- name: CleanSecrets :many
DELETE FROM secrets
WHERE deleted = true
RETURNING *;
