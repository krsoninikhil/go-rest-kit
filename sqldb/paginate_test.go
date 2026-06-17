package sqldb

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type paginateRow struct {
	ID    int `gorm:"primaryKey"`
	Title string
}

func setupPaginateDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&paginateRow{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	for _, row := range []paginateRow{
		{ID: 1, Title: "one"},
		{ID: 2, Title: "two"},
		{ID: 3, Title: "three"},
	} {
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return db
}

func ids(rows []paginateRow) []int {
	out := make([]int, len(rows))
	for i, r := range rows {
		out[i] = r.ID
	}
	return out
}

func TestPaginateNewestFirst_firstPage(t *testing.T) {
	db := setupPaginateDB(t)
	var rows []paginateRow
	err := db.Scopes(PaginateNewestFirst(Page{}, "id")).Find(&rows).Error
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got := ids(rows); len(got) != 3 || got[0] != 3 || got[1] != 2 || got[2] != 1 {
		t.Fatalf("got ids %v want [3 2 1]", got)
	}
}

func TestPaginateNewestFirst_afterCursor(t *testing.T) {
	db := setupPaginateDB(t)
	var rows []paginateRow
	err := db.Scopes(PaginateNewestFirst(Page{After: 3}, "id")).Find(&rows).Error
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got := ids(rows); len(got) != 2 || got[0] != 2 || got[1] != 1 {
		t.Fatalf("got ids %v want [2 1]", got)
	}
}

func TestPaginate_ascAfterCursor(t *testing.T) {
	db := setupPaginateDB(t)
	var rows []paginateRow
	err := db.Scopes(Paginate(Page{After: 1}, "id")).Find(&rows).Error
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got := ids(rows); len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("got ids %v want [2 3]", got)
	}
}
