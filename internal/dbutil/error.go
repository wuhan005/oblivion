// Copyright 2021 E99p1ant. All rights reserved.

package dbutil

import (
	"github.com/jackc/pgconn"
)

func IsUniqueViolation(err error, constraint string) bool {
	// NOTE: How to check if error type is DUPLICATE KEY in GORM.
	// https://github.com/go-gorm/gorm/issues/4037
	pgError, ok := err.(*pgconn.PgError)
	return ok && pgError.Code == "23505" && pgError.ConstraintName == constraint
}
