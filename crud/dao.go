package crud

import (
	"context"
	"errors"

	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/krsoninikhil/go-rest-kit/pgdb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Model interface {
	IsDeleted() bool
	ResourceName() string
	PK() int
}

type Dao[M Model] struct {
	*pgdb.PGDB
}

func (db *Dao[M]) Create(ctx context.Context, m M) (*M, error) {
	if err := db.DB(ctx).Create(&m).Error; err != nil {
		return nil, apperrors.NewServerError(err)
	}
	return &m, nil
}

func (db *Dao[M]) Update(ctx context.Context, id int, m M) (*M, error) {
	if err := db.DB(ctx).Model(&m).Where("id = ?", id).Updates(m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError(m.ResourceName())
		}
		return nil, apperrors.NewServerError(err)
	}
	return &m, nil
}

func (db *Dao[M]) Get(ctx context.Context, id int) (*M, error) {
	var m M
	if err := db.DB(ctx).Where("id = ?", id).Preload(clause.Associations).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError(m.ResourceName())
		}
		return nil, apperrors.NewServerError(err)
	}
	return &m, nil
}

func (db *Dao[M]) Delete(ctx context.Context, id int) error {
	var m M
	res := db.DB(ctx).Delete(&m, id)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError(m.ResourceName())
		}
		return apperrors.NewServerError(res.Error)
	} else if res.RowsAffected == 0 {
		return apperrors.NewNotFoundError(m.ResourceName())
	}
	return nil
}

func (db *Dao[M]) List(ctx context.Context, after int, limit int) (res []M, total int64, err error) {
	var m M
	q := db.DB(ctx).Model(m)
	if err := q.Count(&total).Error; err != nil {
		return nil, total, apperrors.NewServerError(err)
	}

	q.Where("id > ?", after).Order("id DESC")
	if limit > 0 {
		q.Limit(limit)
	}
	if err := q.Preload(clause.Associations).Scan(&res).Error; err != nil {
		return nil, total, apperrors.NewServerError(err)
	}

	return
}

func (db *Dao[M]) BulkCreate(ctx context.Context, m []M) error {
	if err := db.DB(ctx).Create(&m).Error; err != nil {
		return apperrors.NewServerError(err)
	}
	return nil
}
