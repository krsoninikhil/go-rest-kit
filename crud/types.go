package crud

import "context"

type CrudService[M any] interface {
	Get(ctx context.Context, id int) (*M, error)
	Create(ctx context.Context, m M) (*M, error)
	Update(ctx context.Context, id int, m M) (*M, error)
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, after int, limit int) (res []M, total int64, err error)
}

type (
	ListResponse[M any] struct {
		Items []M   `json:"items"`
		Total int64 `json:"total"`
	}

	ResourceParam struct {
		ID int `uri:"parentID" binding:"required"`
	}

	ListParam struct {
		After int `form:"after"`
		Limit int `form:"limit"`
		// CreatedAfter  *time.Time `form:"created_after"`
		// CreatedBefore *time.Time `form:"created_before"`
	}
)
