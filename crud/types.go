package crud

import (
	"context"

	"github.com/krsoninikhil/go-rest-kit/pgdb"
)

type Service[M any] interface {
	Get(ctx context.Context, id int) (*M, error)
	Create(ctx context.Context, m M) (*M, error)
	Update(ctx context.Context, id int, m M) (*M, error)
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, after int, limit int, creatorID int) (res []M, total int64, err error)
	BulkCreate(ctx context.Context, m []M) error
}

type PageItem interface {
	ItemID() int
}

type DaoI[M any] Service[M]

type (
	ListResponse[M any] struct {
		Items     []M   `json:"items"`
		Total     int64 `json:"total"`
		NextAfter int   `json:"next_after,omitempty"`
	}

	ResourceParam struct {
		ID int `uri:"parentID" binding:"required"`
	}

	ListParam struct {
		After int `form:"after"`
		Limit int `form:"limit"`
		Page  int `form:"page"`
		// CreatedAfter  *time.Time `form:"created_after"`
		// CreatedBefore *time.Time `form:"created_before"`
	}
)

func (p ListParam) QueryPage() pgdb.Page {
	return pgdb.Page{
		After: p.After,
		Limit: p.Limit,
		Page:  p.Page,
	}
}
