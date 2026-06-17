package request

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/sqldb"
)

func TestBindGet_doesNotReuseStaleQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var (
		mu   sync.Mutex
		seen []sqldb.Page
	)

	router := gin.New()
	router.GET("/items", BindGet(func(c *gin.Context, p struct {
		sqldb.Page
	}) (*struct{}, error) {
		mu.Lock()
		seen = append(seen, p.Page)
		mu.Unlock()
		return &struct{}{}, nil
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/items?after=99&page=2&limit=10", nil)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request status=%d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("second request status=%d", rec2.Code)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 2 {
		t.Fatalf("seen len=%d want 2", len(seen))
	}
	if seen[0].After != 99 || seen[0].Page != 2 || seen[0].Limit != 10 {
		t.Fatalf("first bind: %+v", seen[0])
	}
	if seen[1].After != 0 || seen[1].Page != 0 || seen[1].Limit != 0 {
		t.Fatalf("stale query leak: %+v", seen[1])
	}
}

func TestBindDelete_doesNotReuseStaleURIParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var (
		mu   sync.Mutex
		seen []int
	)

	router := gin.New()
	router.DELETE("/items/:id", BindDelete(func(c *gin.Context, p struct {
		ID int `uri:"id"`
	}) error {
		mu.Lock()
		seen = append(seen, p.ID)
		mu.Unlock()
		return nil
	}))

	for _, id := range []int{7, 3} {
		req := httptest.NewRequest(http.MethodDelete, "/items/"+strconv.Itoa(id), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
			t.Fatalf("delete id=%d status=%d", id, rec.Code)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 2 || seen[0] != 7 || seen[1] != 3 {
		t.Fatalf("uri bind leak: %v", seen)
	}
}
