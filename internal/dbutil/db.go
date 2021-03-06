// Copyright 2021 E99p1ant. All rights reserved.

package dbutil

import (
	"database/sql"

	"gorm.io/gorm"
)

type Transactor interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) (err error)
}
