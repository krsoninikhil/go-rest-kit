package pgdb

import "gorm.io/gorm"

type CursorPage struct {
	After  uint `form:"after"`
	Before uint `form:"before"`
	Limit  int  `form:"limit"`
}

const DefaultPageLimit = 25
const DefaultPageID = 0

func CursorPaginate(page CursorPage, tableName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		limit, id := CursorLimitID(page)
		if page.After == 0 && page.Before != 0 {
			return db.Where(tableName+`.id < ?`, id).Limit(limit)
		}
		return db.Where(tableName+`.id > ?`, id).Limit(limit)
	}
}

func CursorLimitID(page CursorPage) (int, uint) {
	var id uint
	if page.Limit == 0 {
		page.Limit = DefaultPageLimit
	}
	if page.After == 0 && page.Before == 0 {
		id = DefaultPageID
	}
	if page.After == 0 {
		id = page.Before
	} else {
		id = page.After
	}
	return page.Limit, id
}
