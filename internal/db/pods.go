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
	Create(ctx context.Context, opts CreatePodOptions) (*Pod, error)
	Get(ctx context.Context, opts GetPodsOptions) ([]*Pod, error)
	GetByID(ctx context.Context, id uint) (*Pod, error)
	GetExpired(ctx context.Context) ([]*Pod, error)
	Delete(ctx context.Context, id uint) error
}

// NewPodsStore returns a PodsStore instance with the given database connection.
func NewPodsStore(db *gorm.DB) PodsStore {
	return &pods{DB: db}
}

type Pod struct {
	gorm.Model

	UserID  uint   `uniqueIndex:"pod_user_image_unique_idx, where:deleted_at IS NULL" json:"-"`
	User    *User  `gorm:"-" json:"-"`
	ImageID uint   `uniqueIndex:"pod_user_image_unique_idx, where:deleted_at IS NULL" json:"-"`
	Image   *Image `gorm:"-" json:"-"`

	Name      string
	Address   string
	ExpiredAt time.Time
}

type pods struct {
	*gorm.DB
}

type CreatePodOptions struct {
	UserID    uint
	ImageID   uint
	Name      string
	Address   string
	ExpiredAt time.Time
}

var ErrDuplicatePod = errors.New("duplicate pod")

func (db *pods) Create(ctx context.Context, opts CreatePodOptions) (*Pod, error) {
	pod := &Pod{
		UserID:    opts.UserID,
		ImageID:   opts.ImageID,
		Name:      opts.Name,
		Address:   opts.Address,
		ExpiredAt: opts.ExpiredAt,
	}
	if err := db.WithContext(ctx).Create(pod).Error; err != nil {
		if dbutil.IsUniqueViolation(err, "pod_user_image_unique_idx") {
			return nil, ErrDuplicateUser
		}
		return nil, err
	}
	return pod, nil
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
	return db.loadAttributes(ctx, pods...)
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
	pods, err := db.loadAttributes(ctx, &pod)
	if err != nil {
		return nil, errors.Wrap(err, "load attributes")
	}
	return pods[0], nil
}

func (db *pods) GetExpired(ctx context.Context) ([]*Pod, error) {
	var pods []*Pod
	if err := db.WithContext(ctx).Model(&Pod{}).Where("pods.expired_at < CURRENT_DATE").Find(&pods).Error; err != nil {
		return nil, err
	}
	return db.loadAttributes(ctx, pods...)
}

func (db *pods) Delete(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&Pod{}, id).Error
}

func (db *pods) loadAttributes(ctx context.Context, pods ...*Pod) ([]*Pod, error) {
	userIDs := map[uint]struct{}{}
	imageIDs := map[uint]struct{}{}
	for _, pod := range pods {
		userIDs[pod.UserID] = struct{}{}
		imageIDs[pod.ImageID] = struct{}{}
	}

	// Get pods' users.
	usersStore := NewUsersStore(db.DB)
	userSets := map[uint]*User{}
	for userID := range userIDs {
		var err error
		userSets[userID], err = usersStore.GetByID(ctx, userID)
		if err != nil {
			return nil, errors.Wrap(err, "get users")
		}
	}

	// Get pods' images.
	imagesStore := NewImagesStore(db.DB)
	imageSets := map[uint]*Image{}
	for imageID := range imageIDs {
		var err error
		imageSets[imageID], err = imagesStore.GetByID(ctx, imageID)
		if err != nil {
			return nil, errors.Wrap(err, "get image")
		}
	}

	for _, pod := range pods {
		pod.User = userSets[pod.UserID]
		pod.Image = imageSets[pod.ImageID]
	}

	return pods, nil
}
