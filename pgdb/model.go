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
func (s BaseModel) Joins() []string { return []string{} }

type NamedModel interface {
	GetName() string
	PK() int
}

func ToNameModelMap[M NamedModel](models []M) map[string]M {
	var res = make(map[string]M, len(models))
	for _, model := range models {
		res[model.GetName()] = model
	}
	return res
}

func ToIDModelMap[M NamedModel](models []M) map[int]M {
	var res = make(map[int]M, len(models))
	for _, model := range models {
		res[model.PK()] = model
	}
	return res
}
