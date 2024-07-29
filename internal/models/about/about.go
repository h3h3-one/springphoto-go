package about

import "database/sql"

type About struct {
	db *sql.DB
}

func New(storage *sql.DB) *About {
	return &About{db: storage}
}
