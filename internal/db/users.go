// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"

	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
	"gorm.io/gorm"

	"github.com/wuhan005/oblivion/internal/dbutil"
)

var _ UsersStore = (*users)(nil)

// Users is the default instance of the UsersStore.
var Users UsersStore

// UsersStore is the persistent interface for users.
type UsersStore interface {
	// Create creates a new user.
	Create(ctx context.Context, opts CreateUserOptions) error
	// BatchCreate creates new users in batch.
	BatchCreate(ctx context.Context, opts BatchCreateOptions) error
	// GetByID returns a user by its ID.
	GetByID(ctx context.Context, id uint) (*User, error)
	// GetByToken returns a user by its token.
	GetByToken(ctx context.Context, token string) (*User, error)
	// GetByDomain returns a user by its domain.
	GetByDomain(ctx context.Context, domain string) (*User, error)
	// Delete deletes a user by its ID.
	Delete(ctx context.Context, id uint) error
}

// NewUsersStore returns a UsersStore instance with the given database connection.
func NewUsersStore(db *gorm.DB) UsersStore {
	return &users{DB: db}
}

type User struct {
	gorm.Model

	Token  string `uniqueIndex:"user_token_unique_idx, where:deleted_at IS NULL"`
	Domain string `uniqueIndex:"user_domain_unique_idx, where:deleted_at IS NULL"`
}

type users struct {
	*gorm.DB
}

type CreateUserOptions struct {
	Token string
}

var ErrDuplicateUser = errors.New("duplicate user")

func (db *users) Create(ctx context.Context, opts CreateUserOptions) error {
	if err := db.WithContext(ctx).Create(&User{
		Token:  opts.Token,
		Domain: randstr.String(8),
	}).Error; err != nil {
		if dbutil.IsUniqueViolation(err, "user_token_unique_idx") {
			return ErrDuplicateUser
		}
		return err
	}
	return nil
}

type BatchCreateOptions struct {
	Tokens []string
}

func (db *users) BatchCreate(ctx context.Context, opts BatchCreateOptions) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, token := range opts.Tokens {
			if err := tx.Create(&User{
				Token:  token,
				Domain: randstr.String(8),
			}).Error; err != nil {
				if dbutil.IsUniqueViolation(err, "user_token_unique_idx") {
					return ErrDuplicateUser
				}
				return err
			}
		}
		return nil
	})
}

var ErrUserNotFound = errors.New("user dose not exist")

func (db *users) GetByID(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (db *users) GetByToken(ctx context.Context, token string) (*User, error) {
	var user User
	if err := db.WithContext(ctx).Where("token = ?", token).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (db *users) GetByDomain(ctx context.Context, domain string) (*User, error) {
	var user User
	if err := db.WithContext(ctx).Where("domain = ?", domain).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (db *users) Delete(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&User{}, id).Error
}
