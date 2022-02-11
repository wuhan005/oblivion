// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/wuhan005/oblivion/internal/dbutil"
)

var _ ImagesStore = (*images)(nil)

// Images is the default instance of the ImagesStore.
var Images ImagesStore

// ImagesStore is the persistent interface for images.
type ImagesStore interface {
	Create(ctx context.Context, opts CreateImageOptions) error
	GetByID(ctx context.Context, id uint) (*Image, error)
	Update(ctx context.Context, id uint, opts UpdateImageOptions) error
	Delete(ctx context.Context, id uint) error
}

// NewImagesStore returns a ImagesStore instance with the given database connection.
func NewImagesStore(db *gorm.DB) ImagesStore {
	return &images{DB: db}
}

type Image struct {
	gorm.Model

	UID       string
	Name      string `uniqueIndex:image_name_unique_idx, where:deleted_at IS NULL`
	Namespace string
}

type images struct {
	*gorm.DB
}

type CreateImageOptions struct {
	Name      string
	Namespace string
}

var ErrDuplicateImage = errors.New("duplicate image")

func (db *images) Create(ctx context.Context, opts CreateImageOptions) error {
	if err := db.WithContext(ctx).Create(&Image{
		UID:       uuid.New().String(),
		Name:      opts.Name,
		Namespace: opts.Namespace,
	}).Error; err != nil {
		if dbutil.IsUniqueViolation(err, "image_name_unique_idx") {
			return ErrDuplicateUser
		}
		return err
	}
	return nil
}

var ErrImageNotFound = errors.New("image does not exist")

func (db *images) GetByID(ctx context.Context, id uint) (*Image, error) {
	var image Image
	if err := db.WithContext(ctx).First(&image, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrImageNotFound
		}
		return nil, err
	}
	return &image, nil
}

type UpdateImageOptions struct {
	Name      string
	Namespace string
}

func (db *images) Update(ctx context.Context, id uint, opts UpdateImageOptions) error {
	var image Image
	if err := db.WithContext(ctx).First(&image, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrImageNotFound
		}
		return err
	}
	return db.WithContext(ctx).Where("id = ?", id).Updates(&Image{
		Name:      opts.Name,
		Namespace: opts.Namespace,
	}).Error
}

func (db *images) Delete(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&Image{}, id).Error
}
