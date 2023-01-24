// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: users.sql

package db

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  name,
  passhash
) VALUES (
  $1, $2
)
RETURNING id, name, passhash
`

type CreateUserParams struct {
	Name     string
	Passhash string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.Name, arg.Passhash)
	var i User
	err := row.Scan(&i.ID, &i.Name, &i.Passhash)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users
WHERE name = $1
`

func (q *Queries) DeleteUser(ctx context.Context, name string) error {
	_, err := q.db.ExecContext(ctx, deleteUser, name)
	return err
}

const getUser = `-- name: GetUser :one
SELECT id, name, passhash FROM users
WHERE name = $1 AND passhash = $2
LIMIT 1
`

type GetUserParams struct {
	Name     string
	Passhash string
}

func (q *Queries) GetUser(ctx context.Context, arg GetUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, getUser, arg.Name, arg.Passhash)
	var i User
	err := row.Scan(&i.ID, &i.Name, &i.Passhash)
	return i, err
}
