// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"os"
	"time"

	"github.com/pkg/errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/wuhan005/oblivion/internal/dbutil"
)

// Init initializes the database.
func Init() (*gorm.DB, error) {
	dsn := os.ExpandEnv("postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=$POSTGRES_SSLMODE")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return dbutil.Now()
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "open connection")
	}

	// Migrate databases.
	if db.AutoMigrate(&Image{}, &Pod{}, &User{}) != nil {
		return nil, errors.Wrap(err, "auto migrate")
	}

	Images = NewImagesStore(db)
	Pods = NewPodsStore(db)
	Users = NewUsersStore(db)

	return db, nil
}
