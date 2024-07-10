package pgdb

import "gorm.io/gorm"

const DefaultPageLimit = 25
const MaxPageLimit = 100

type Page struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
	After int `form:"after"`
}

func (p Page) Offset() int {
	currentPage := p.Page
	if currentPage > 0 {
		currentPage -= 1
	}
	return currentPage * p.Limit
}

func Paginate(page Page, afterField string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page.Limit == 0 || page.Limit > MaxPageLimit {
			page.Limit = DefaultPageLimit
		}
		if page.After > 0 {
			db = db.Where(afterField+" > ?", page.After)
			db = db.Order(afterField + " ASC")
		}
		if page.Page > 0 {
			db = db.Offset(page.Offset())
		}
		return db.Limit(page.Limit)
	}
}
