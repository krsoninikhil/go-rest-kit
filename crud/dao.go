package crud

import (
	"context"
	"errors"

	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/krsoninikhil/go-rest-kit/pgdb"
	"gorm.io/gorm"
)

type Model interface {
	IsDeleted() bool
	ResourceName() string
	PK() int
	Joins() []string
}

type ModelWithCreator interface {
	SetCreatedBy(userID int)
	CreatedByID() int
}

type Dao[M Model] struct {
	*pgdb.PGDB
}

func (db *Dao[M]) Create(ctx context.Context, m M) (*M, error) {
	if err := db.DB(ctx).Create(&m).Error; errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, apperrors.NewConflictError(m.ResourceName(), err)
	} else if err != nil {
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
	q := db.DB(ctx).Model(m)
	for _, joins := range m.Joins() {
		q = q.Joins(joins)
	}

	tableName := db.TableName(m)
	if err := q.Where(tableName+".id = ?", id).First(&m).Error; err != nil {
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

func (db *Dao[M]) List(ctx context.Context, page pgdb.Page, creatorID int) (res []M, total int64, err error) {
	var m M
	q := db.DB(ctx).Model(m)

	if mc, ok := any(&m).(ModelWithCreator); ok {
		mc.SetCreatedBy(creatorID)
		q = q.Where(mc)
	} else if nm, ok := any(m).(NestedModel[M]); ok {
		nm = any(nm.SetParentID(creatorID)).(NestedModel[M])
		q = q.Where(nm)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, total, apperrors.NewServerError(err)
	}

	tableName := q.Statement.Table
	q = q.Scopes(pgdb.Paginate(page, tableName+".id"))
	for _, joins := range m.Joins() {
		q.Preload(joins)
	}

	if err := q.Find(&res).Error; err != nil {
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
