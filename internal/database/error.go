package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

func IsUniqueViolation(err error) bool     { return hasCode(err, pgUniqueViolation) }
func IsForeignKeyViolation(err error) bool { return hasCode(err, pgForeignKeyViolation) }

func hasCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}
