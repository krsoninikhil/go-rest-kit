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
	PK() int
	ResourceName() string
}

// UserWithPassword is an optional capability required only for password auth flows.
// Models that don't support password auth don't need to implement it.
type UserWithPassword interface {
	SetEmail(string) UserModel
	SetName(string) UserModel
	SetPasswordHash(string) UserModel
	GetPasswordHash() string
}

type UserWithUsername interface {
	SetUsername(string) UserModel
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
	passwordModel, ok := any(user).(UserWithPassword)
	if !ok {
		return 0, "", apperrors.NewServerError(errors.New("user model does not support password authentication"))
	}
	return user.PK(), passwordModel.GetPasswordHash(), nil
}

func (d *userDao[U]) GetByID(ctx context.Context, userID int) (string, string, error) {
	var user U
	err := d.PGDB.DB(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", apperrors.NewNotFoundError(user.ResourceName())
		}
		return "", "", apperrors.NewServerError(err)
	}
	passwordModel, ok := any(user).(UserWithPassword)
	if !ok {
		return "", "", apperrors.NewServerError(errors.New("user model does not support password authentication"))
	}
	var lookup struct {
		Email string `gorm:"column:email"`
	}
	if err := d.PGDB.DB(ctx).Model(&user).Select("email").Where("id = ?", userID).Take(&lookup).Error; err != nil {
		return "", "", apperrors.NewServerError(err)
	}
	return lookup.Email, passwordModel.GetPasswordHash(), nil
}

func (d *userDao[U]) CreateWithEmailAndPassword(ctx context.Context, email, name, passwordHash, username, locale string) (int, error) {
	var user U
	passwordModel, ok := any(user).(UserWithPassword)
	if !ok {
		return 0, apperrors.NewServerError(errors.New("user model does not support password authentication"))
	}
	user = passwordModel.SetEmail(email).(U)
	if name != "" {
		user = passwordModel.SetName(name).(U)
	}
	user = passwordModel.SetPasswordHash(passwordHash).(U)
	user = user.SetSignupInfo(SigupInfo{Email: email, Locale: locale}).(U)

	usernameModel, ok := any(user).(UserWithUsername)
	if ok && username != "" {
		user = usernameModel.SetUsername(username).(U)
	}

	if err := d.PGDB.DB(ctx).Create(&user).Error; err != nil {
		return 0, err
	}
	return user.PK(), nil
}

func (d *userDao[U]) SupportsUsername() bool {
	var user U
	_, ok := any(user).(UserWithUsername)
	return ok
}

func (d *userDao[U]) SupportsPassword() bool {
	var user U
	_, ok := any(user).(UserWithPassword)
	return ok
}

func (d *userDao[U]) SetPasswordHashByUserID(ctx context.Context, userID int, passwordHash string) error {
	var user U
	passwordModel, ok := any(user).(UserWithPassword)
	if !ok {
		return apperrors.NewServerError(errors.New("user model does not support password authentication"))
	}
	updateUser := passwordModel.SetPasswordHash(passwordHash).(U)
	if err := d.PGDB.DB(ctx).Model(&user).Where("id = ?", userID).Updates(&updateUser).Error; err != nil {
		return apperrors.NewServerError(err)
	}
	return nil
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
