package bunex

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsErrorUniqueViolation(err error) bool {
	sqlErr := &pgconn.PgError{}
	// Ref: https://www.postgresql.org/docs/current/errcodes-appendix.html
	return errors.As(err, &sqlErr) && (sqlErr.Code == "23505")
}

func IsErrorColumnNotExist(err error) bool {
	sqlErr := &pgconn.PgError{}
	// Ref: https://www.postgresql.org/docs/current/errcodes-appendix.html
	return errors.As(err, &sqlErr) && (sqlErr.Code == "42703")
}
