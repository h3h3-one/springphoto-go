package contact

import "database/sql"

type Contact struct {
	db *sql.DB
}

func New(storage *sql.DB) *Contact {
	return &Contact{db: storage}
}
