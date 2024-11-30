package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func Connect() (*sql.DB, error) {
	return sql.Open("pgx", "host=localhost user=postgres password=7dgvJVDJvh254aqOpfd dbname=postgres sslmode=disable")
}
