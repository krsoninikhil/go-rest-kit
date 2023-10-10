package pgdb

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (d BaseModel) IsDeleted() bool { return d.DeletedAt.Valid }
func (d BaseModel) PK() int         { return d.ID }
