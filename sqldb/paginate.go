package sqldb

import "gorm.io/gorm"

const DefaultPageLimit = 25
const MaxPageLimit = 100

type Page struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
	After int `form:"after"`
}

func NewPage(page, after, limit int) Page {
	return Page{
		Page:  page,
		Limit: limit,
		After: after,
	}
}

func (p Page) Offset() int {
	currentPage := p.Page
	if currentPage > 0 {
		currentPage -= 1
	}
	return currentPage * p.Limit
}

func pageLimit(limit int) int {
	if limit == 0 || limit > MaxPageLimit {
		return DefaultPageLimit
	}
	return limit
}

// Paginate applies ascending cursor pagination on afterField.
// When after is set, returns rows where afterField > after ordered ascending.
// next_after should be the last item id/value from the returned page.
func Paginate(page Page, afterField string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		limit := pageLimit(page.Limit)
		if page.After > 0 {
			db = db.Where(afterField+" > ?", page.After)
			db = db.Order(afterField + " ASC")
		} else if page.Page > 0 {
			db = db.Offset(page.Offset())
		}
		return db.Limit(limit)
	}
}

// PaginateNewestFirst applies descending cursor pagination on afterField.
// First page is ordered newest-first. When after is set, returns rows where
// afterField < after (older items when afterField is a monotonic id/timestamp).
// next_after should still be the last item id/value from the returned page.
func PaginateNewestFirst(page Page, afterField string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		limit := pageLimit(page.Limit)
		if page.After > 0 {
			db = db.Where(afterField+" < ?", page.After)
			db = db.Order(afterField + " DESC")
		} else if page.Page > 0 {
			db = db.Order(afterField + " DESC").Offset(page.Offset())
		} else {
			db = db.Order(afterField + " DESC")
		}
		return db.Limit(limit)
	}
}
