// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package db

import (
	"time"
)

type Secret struct {
	ID       int64
	Owner    string
	Kind     int32
	Name     string
	Value    []byte
	Created  time.Time
	Modified time.Time
	Deleted  bool
}

type User struct {
	ID       int32
	Name     string
	Passhash string
}
