package auth

import (
	"context"
	"errors"

	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/krsoninikhil/go-rest-kit/pgdb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserModel interface {
	SetPhone(string) UserModel
	SetSignupInfo(SigupInfo) UserModel
	PK() int
	ResourceName() string
}

type userDao[U UserModel] struct {
	*pgdb.PGDB
}

func NewUserDao[U UserModel](db *pgdb.PGDB) *userDao[U] {
	return &userDao[U]{db}
}

func (d *userDao[U]) Create(ctx context.Context, u SigupInfo) (int, error) {
	var user U
	user = user.SetPhone(u.Phone).(U)
	user = user.SetSignupInfo(u).(U)
	if err := d.PGDB.DB(ctx).Create(&user).Error; err != nil {
		return 0, err
	}
	return user.PK(), nil
}

// Upsert doesn't return the id if the user already exists
func (d *userDao[U]) Upsert(ctx context.Context, phone string) (int, error) {
	var user U
	user = user.SetPhone(phone).(U)
	err := d.PGDB.DB(ctx).Clauses(
		clause.Returning{Columns: []clause.Column{{Name: "id"}}},
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "phone"}},
			DoNothing: true,
		},
	).Create(&user).Error
	if err != nil {
		return 0, err
	}
	return user.PK(), nil
}

func (d *userDao[U]) GetByPhone(ctx context.Context, phone string) (int, error) {
	var user U
	err := d.PGDB.DB(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.NewNotFoundError(user.ResourceName())
		}
		return 0, apperrors.NewServerError(err)
	}
	return user.PK(), nil
}
