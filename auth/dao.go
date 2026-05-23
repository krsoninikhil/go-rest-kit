package auth

import (
	"context"
	"errors"

	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/krsoninikhil/go-rest-kit/sqldb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserModel interface {
	SetPhone(string) UserModel
	SetSignupInfo(SigupInfo) UserModel
	SetOAuthInfo(OAuthUserInfo) UserModel
	SetEmail(string) UserModel
	SetName(string) UserModel
	SetPasswordHash(string) UserModel
	GetPasswordHash() string
	PK() int
	ResourceName() string
}

type userDao[U UserModel] struct {
	*sqldb.PGDB
}

func NewUserDao[U UserModel](db *sqldb.PGDB) *userDao[U] {
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

func (d *userDao[U]) GetByEmail(ctx context.Context, email string) (int, error) {
	var user U
	err := d.PGDB.DB(ctx).Where("LOWER(email) = LOWER(?)", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.NewNotFoundError(user.ResourceName())
		}
		return 0, apperrors.NewServerError(err)
	}
	return user.PK(), nil
}

func (d *userDao[U]) UpsertByEmail(ctx context.Context, oauthInfo OAuthUserInfo) (int, error) {
	userID, err := d.GetByEmail(ctx, oauthInfo.Email)
	if err == nil {
		return userID, nil
	}
	if _, ok := err.(apperrors.NotFoundError); !ok {
		return 0, err
	}

	var user U
	user = user.SetOAuthInfo(oauthInfo).(U)
	err = d.PGDB.DB(ctx).Clauses(
		clause.Returning{Columns: []clause.Column{{Name: "id"}}},
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "email"}},
			DoNothing: true,
		},
	).Create(&user).Error
	if err != nil {
		return 0, err
	}
	// When conflict happened, Create did nothing and user.ID is still 0 — fetch existing user by email
	if user.PK() == 0 {
		return d.GetByEmail(ctx, oauthInfo.Email)
	}
	return user.PK(), nil
}

func (d *userDao[U]) GetByEmailWithPassword(ctx context.Context, email string) (int, string, error) {
	var user U
	err := d.PGDB.DB(ctx).Where("LOWER(email) = LOWER(?)", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, "", apperrors.NewNotFoundError(user.ResourceName())
		}
		return 0, "", apperrors.NewServerError(err)
	}
	return user.PK(), user.GetPasswordHash(), nil
}

func (d *userDao[U]) CreateWithEmailAndPassword(ctx context.Context, email, name, passwordHash, locale string) (int, error) {
	var user U
	user = user.SetEmail(email).(U)
	user = user.SetName(name).(U)
	user = user.SetPasswordHash(passwordHash).(U)
	user = user.SetSignupInfo(SigupInfo{Email: email, Locale: locale}).(U)
	if err := d.PGDB.DB(ctx).Create(&user).Error; err != nil {
		return 0, err
	}
	return user.PK(), nil
}

func (d *userDao[U]) GetByUsername(ctx context.Context, username string) (int, error) {
	var user U
	err := d.PGDB.DB(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.NewNotFoundError(user.ResourceName())
		}
		return 0, apperrors.NewServerError(err)
	}
	return user.PK(), nil
}
