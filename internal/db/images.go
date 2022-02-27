// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
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
	GetByUID(ctx context.Context, uid string) (*Image, error)
	Update(ctx context.Context, id uint, opts UpdateImageOptions) error
	Delete(ctx context.Context, id uint) error
}

// NewImagesStore returns a ImagesStore instance with the given database connection.
func NewImagesStore(db *gorm.DB) ImagesStore {
	return &images{DB: db}
}

type Image struct {
	gorm.Model

	UID        string
	Name       string `uniqueIndex:"image_name_unique_idx, where:deleted_at IS NULL"`
	Domain     string
	Port       int32
	Limitation datatypes.JSON `gorm:"type:jsonb"`
}

func (i *Image) GetLimitation() *ImageLimitation {
	var limitation ImageLimitation
	_ = json.Unmarshal(i.Limitation, &limitation)
	return &limitation
}

type ImageLimitation struct {
	LimitsCPU      string
	LimitsMemory   string
	RequestsCPU    string
	RequestsMemory string
}

type images struct {
	*gorm.DB
}

type CreateImageOptions struct {
	Name       string
	Domain     string
	Port       int32
	Limitation ImageLimitation
}

var ErrDuplicateImage = errors.New("duplicate image")

func (db *images) Create(ctx context.Context, opts CreateImageOptions) error {
	limitation, _ := json.Marshal(opts.Limitation)

	if err := db.WithContext(ctx).Create(&Image{
		UID:        uuid.New().String(),
		Name:       opts.Name,
		Domain:     opts.Domain,
		Port:       opts.Port,
		Limitation: limitation,
	}).Error; err != nil {
		if dbutil.IsUniqueViolation(err, "image_name_unique_idx") {
			return ErrDuplicateImage
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

func (db *images) GetByUID(ctx context.Context, uid string) (*Image, error) {
	var image Image
	if err := db.WithContext(ctx).Where("uid = ?", uid).First(&image).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrImageNotFound
		}
		return nil, err
	}
	return &image, nil
}

type UpdateImageOptions struct {
	Name       string
	Domain     string
	Port       int32
	Limitation ImageLimitation
}

func (db *images) Update(ctx context.Context, id uint, opts UpdateImageOptions) error {
	limitation, _ := json.Marshal(opts.Limitation)

	var image Image
	if err := db.WithContext(ctx).First(&image, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrImageNotFound
		}
		return err
	}
	return db.WithContext(ctx).Where("id = ?", id).Updates(&Image{
		Name:       opts.Name,
		Domain:     opts.Domain,
		Port:       opts.Port,
		Limitation: limitation,
	}).Error
}

func (db *images) Delete(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&Image{}, id).Error
}
