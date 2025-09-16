package handler_test

import (
	"bytes"
	"context"
	"employer/internal/domain"
	"employer/internal/handler"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type mockService struct {
	CreateFn func(ctx context.Context, e *domain.Employee) error
	GetFn    func(ctx context.Context, id int) (*domain.Employee, error)
	GetAllFn func(ctx context.Context) ([]*domain.Employee, error)
	UpdateFn func(ctx context.Context, e *domain.Employee) error
	DeleteFn func(ctx context.Context, id int) error
	SearchFn func(ctx context.Context, query string) ([]*domain.Employee, error) // Added
}

func (m *mockService) CreateEmployee(ctx context.Context, e *domain.Employee) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, e)
	}
	return nil
}

func (m *mockService) GetEmployee(ctx context.Context, id int) (*domain.Employee, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) GetAllEmployees(ctx context.Context) ([]*domain.Employee, error) {
	if m.GetAllFn != nil {
		return m.GetAllFn(ctx)
	}
	return nil, nil
}

func (m *mockService) UpdateEmployee(ctx context.Context, e *domain.Employee) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, e)
	}
	return nil
}

func (m *mockService) DeleteEmployee(ctx context.Context, id int) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

// Added SearchEmployees method
func (m *mockService) SearchEmployees(ctx context.Context, query string) ([]*domain.Employee, error) {
	if m.SearchFn != nil {
		return m.SearchFn(ctx, query)
	}
	return []*domain.Employee{}, nil
}

func newRouter(svc *mockService) *mux.Router {
	log := zap.NewNop()
	h := handler.NewEmployeeHandler(svc, log)
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	return r
}

// --- existing tests ---

func TestCreateEmployee_Success(t *testing.T) {
	svc := &mockService{
		CreateFn: func(ctx context.Context, e *domain.Employee) error {
			e.ID = 101
			return nil
		},
	}
	r := newRouter(svc)

	body := `{"name":"Alice","phone":"+77010000000","city":"Almaty"}`
	req := httptest.NewRequest(http.MethodPost, "/api/employees", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, rr.Code)
	}
	var resp domain.EmployeeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != 101 || resp.Name != "Alice" || resp.Phone != "+77010000000" || resp.City != "Almaty" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestGetEmployee_Success(t *testing.T) {
	svc := &mockService{
		GetFn: func(ctx context.Context, id int) (*domain.Employee, error) {
			return &domain.Employee{ID: id, Name: "Bob", Phone: "123", City: "Astana"}, nil
		},
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/employees/7", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}
	var resp domain.EmployeeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != 7 || resp.Name != "Bob" || resp.City != "Astana" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestGetAllEmployees_Success(t *testing.T) {
	svc := &mockService{
		GetAllFn: func(ctx context.Context) ([]*domain.Employee, error) {
			return []*domain.Employee{
				{ID: 1, Name: "A", Phone: "1", City: "X"},
				{ID: 2, Name: "B", Phone: "2", City: "Y"},
			}, nil
		},
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/employees", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}
	var list []domain.EmployeeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 2 || list[0].ID != 1 || list[1].ID != 2 {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestUpdateEmployee_Success(t *testing.T) {
	svc := &mockService{
		UpdateFn: func(ctx context.Context, e *domain.Employee) error { return nil },
	}
	r := newRouter(svc)

	body := `{"name":"Neo","phone":"777","city":"Matrix"}`
	req := httptest.NewRequest(http.MethodPut, "/api/employees/10", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}
	var resp domain.EmployeeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != 10 || resp.Name != "Neo" || resp.Phone != "777" || resp.City != "Matrix" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestDeleteEmployee_Success(t *testing.T) {
	svc := &mockService{
		DeleteFn: func(ctx context.Context, id int) error { return nil },
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/employees/12", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, rr.Code)
	}
}

// --- new search tests ---

func TestSearchEmployees_Success(t *testing.T) {
	svc := &mockService{
		SearchFn: func(ctx context.Context, query string) ([]*domain.Employee, error) {
			// Mock search results based on query
			if query == "john" {
				return []*domain.Employee{
					{ID: 1, Name: "John Doe", Phone: "+77777777777", City: "Almaty"},
					{ID: 2, Name: "John Smith", Phone: "+77777777778", City: "Astana"},
				}, nil
			}
			if query == "777" {
				return []*domain.Employee{
					{ID: 1, Name: "John Doe", Phone: "+77777777777", City: "Almaty"},
				}, nil
			}
			return []*domain.Employee{}, nil
		},
	}
	r := newRouter(svc)

	// Test search by name
	req := httptest.NewRequest(http.MethodGet, "/api/employees/search?q=john", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var results []domain.EmployeeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &results); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Name != "John Doe" || results[1].Name != "John Smith" {
		t.Fatalf("unexpected search results: %+v", results)
	}
}

func TestSearchEmployees_ByPhone(t *testing.T) {
	svc := &mockService{
		SearchFn: func(ctx context.Context, query string) ([]*domain.Employee, error) {
			if query == "777" {
				return []*domain.Employee{
					{ID: 1, Name: "John Doe", Phone: "+77777777777", City: "Almaty"},
				}, nil
			}
			return []*domain.Employee{}, nil
		},
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/employees/search?q=777", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var results []domain.EmployeeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &results); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Phone != "+77777777777" {
		t.Fatalf("unexpected phone in search result: %s", results[0].Phone)
	}
}

func TestSearchEmployees_EmptyQuery(t *testing.T) {
	svc := &mockService{}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/employees/search", nil) // No query parameter
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var errResp domain.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if errResp.Error == "" {
		t.Fatalf("expected error message, got empty string")
	}
}

func TestSearchEmployees_NoResults(t *testing.T) {
	svc := &mockService{
		SearchFn: func(ctx context.Context, query string) ([]*domain.Employee, error) {
			return []*domain.Employee{}, nil // No results
		},
	}
	r := newRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/employees/search?q=nonexistent", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var results []domain.EmployeeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &results); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

// Benchmark test for search performance
func BenchmarkSearchEmployees(b *testing.B) {
	svc := &mockService{
		SearchFn: func(ctx context.Context, query string) ([]*domain.Employee, error) {
			// Simulate realistic search results
			results := make([]*domain.Employee, 50)
			for i := range results {
				results[i] = &domain.Employee{
					ID:    i + 1,
					Name:  "Employee " + string(rune(i)),
					Phone: "+77777777777",
					City:  "Almaty",
				}
			}
			return results, nil
		},
	}
	r := newRouter(svc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/employees/search?q=test", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status code: %d", rr.Code)
		}
	}
}
