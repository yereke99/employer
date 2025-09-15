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

// --- minimal mock of service.EmployeeService ---

type mockService struct {
	CreateFn func(ctx context.Context, e *domain.Employee) error
	GetFn    func(ctx context.Context, id int) (*domain.Employee, error)
	GetAllFn func(ctx context.Context) ([]*domain.Employee, error)
	UpdateFn func(ctx context.Context, e *domain.Employee) error
	DeleteFn func(ctx context.Context, id int) error
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

func newRouter(svc *mockService) *mux.Router {
	log := zap.NewNop()
	h := handler.NewEmployeeHandler(svc, log)
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	return r
}

// --- tests ---

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
