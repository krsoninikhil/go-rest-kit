package auth

import (
	"context"

	"github.com/krsoninikhil/go-rest-kit/pgdb"
)

type UserModel interface {
	SetPhone(string) UserModel
	PK() int
}

type userDao[U UserModel] struct {
	*pgdb.PGDB
}

func NewUserDao[U UserModel](db *pgdb.PGDB) *userDao[U] {
	return &userDao[U]{db}
}

func (d *userDao[U]) Create(ctx context.Context, phone string) (int, error) {
	var user U
	user = user.SetPhone(phone).(U)
	if err := d.PGDB.DB(ctx).Create(&user).Error; err != nil {
		return 0, err
	}
	return user.PK(), nil
}
