// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/wuhan005/oblivion/internal/dbutil"
)

var _ PodsStore = (*pods)(nil)

// Pods is the default instance of the PodsStore.
var Pods PodsStore

// PodsStore is the persistent interface for pods.
type PodsStore interface {
	Create(ctx context.Context, opts CreatePodOptions) error
	Get(ctx context.Context, opts GetPodsOptions) ([]*Pod, error)
	GetByID(ctx context.Context, id uint) (*Pod, error)
	Delete(ctx context.Context, id uint) error
}

// NewPodsStore returns a PodsStore instance with the given database connection.
func NewPodsStore(db *gorm.DB) PodsStore {
	return &pods{DB: db}
}

type Pod struct {
	gorm.Model

	UserID  uint   `uniqueIndex:pod_user_image_unique_idx, where:deleted_at IS NULL`
	User    *User  `gorm:"-"`
	ImageID uint   `uniqueIndex:pod_user_image_unique_idx, where:deleted_at IS NULL`
	Image   *Image `gorm:"-"`

	Address   string
	ExpiredAt time.Time
}

type pods struct {
	*gorm.DB
}

type CreatePodOptions struct {
	UserID    uint
	ImageID   uint
	Address   string
	ExpiredAt time.Time
}

var ErrDuplicatePod = errors.New("duplicate pod")

func (db *pods) Create(ctx context.Context, opts CreatePodOptions) error {
	if err := db.WithContext(ctx).Create(&Pod{
		UserID:    opts.UserID,
		ImageID:   opts.ImageID,
		Address:   opts.Address,
		ExpiredAt: opts.ExpiredAt,
	}).Error; err != nil {
		if dbutil.IsUniqueViolation(err, "pod_user_image_unique_idx") {
			return ErrDuplicateUser
		}
		return err
	}
	return nil
}

type GetPodsOptions struct {
	UserID  uint
	ImageID uint
}

func (db *pods) Get(ctx context.Context, opts GetPodsOptions) ([]*Pod, error) {
	var pods []*Pod
	if err := db.WithContext(ctx).Where(&Pod{
		UserID:  opts.UserID,
		ImageID: opts.ImageID,
	}).Find(&pods).Error; err != nil {
		return nil, err
	}
	return pods, nil
}

var ErrPodsNotFound = errors.New("pods dose not exist")

func (db *pods) GetByID(ctx context.Context, id uint) (*Pod, error) {
	var pod Pod
	if err := db.WithContext(ctx).First(&pod, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPodsNotFound
		}
		return nil, err
	}
	return &pod, nil
}

func (db *pods) Delete(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&Pod{}, id).Error
}
