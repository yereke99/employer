package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"employer/internal/domain"
	"employer/internal/repository"

	"go.uber.org/zap"
)

// мок репозитория под интерфейс repository.EmployeeRepository
type mockRepo struct {
	CreateFn             func(ctx context.Context, e *domain.Employee) error
	GetByIDFn            func(ctx context.Context, id int) (*domain.Employee, error)
	GetAllFn             func(ctx context.Context) ([]*domain.Employee, error)
	UpdateFn             func(ctx context.Context, e *domain.Employee) error
	DeleteFn             func(ctx context.Context, id int) error
	GetByPhoneFn         func(ctx context.Context, phone string) (*domain.Employee, error)
	SearchEmployeesFn    func(ctx context.Context, searchQuery string) ([]*domain.Employee, error)
	GetEmployeesByCityFn func(ctx context.Context, city string) ([]*domain.Employee, error)
	GetEmployeeStatsFn   func(ctx context.Context) (*repository.EmployeeStats, error)
	CheckPhoneExistsFn   func(ctx context.Context, phone string, excludeID ...int) (bool, error)
}

func (m *mockRepo) Create(ctx context.Context, e *domain.Employee) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, e)
	}
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id int) (*domain.Employee, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockRepo) GetAll(ctx context.Context) ([]*domain.Employee, error) {
	if m.GetAllFn != nil {
		return m.GetAllFn(ctx)
	}
	return nil, nil
}

func (m *mockRepo) Update(ctx context.Context, e *domain.Employee) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, e)
	}
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id int) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *mockRepo) GetByPhone(ctx context.Context, phone string) (*domain.Employee, error) {
	if m.GetByPhoneFn != nil {
		return m.GetByPhoneFn(ctx, phone)
	}
	return nil, nil
}

func (m *mockRepo) SearchEmployees(ctx context.Context, searchQuery string) ([]*domain.Employee, error) {
	if m.SearchEmployeesFn != nil {
		return m.SearchEmployeesFn(ctx, searchQuery)
	}
	return []*domain.Employee{}, nil
}

func (m *mockRepo) GetEmployeesByCity(ctx context.Context, city string) ([]*domain.Employee, error) {
	if m.GetEmployeesByCityFn != nil {
		return m.GetEmployeesByCityFn(ctx, city)
	}
	return []*domain.Employee{}, nil
}

func (m *mockRepo) GetEmployeeStats(ctx context.Context) (*repository.EmployeeStats, error) {
	if m.GetEmployeeStatsFn != nil {
		return m.GetEmployeeStatsFn(ctx)
	}
	return &repository.EmployeeStats{}, nil
}

func (m *mockRepo) CheckPhoneExists(ctx context.Context, phone string, excludeID ...int) (bool, error) {
	if m.CheckPhoneExistsFn != nil {
		return m.CheckPhoneExistsFn(ctx, phone, excludeID...)
	}
	return false, nil
}

// Убедись, что тип удовлетворяет интерфейсу (компиляционная проверка)
var _ repository.EmployeeRepository = (*mockRepo)(nil)

// Существующие тесты
func TestCreateEmployee_Success(t *testing.T) {
	repo := &mockRepo{
		CreateFn: func(ctx context.Context, e *domain.Employee) error {
			e.ID = 42
			return nil
		},
	}
	svc := NewEmployeeService(repo, zap.NewNop())

	e := &domain.Employee{Name: "Alice", Phone: "+7701", City: "Almaty"}
	if err := svc.CreateEmployee(context.Background(), e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.ID != 42 {
		t.Fatalf("expected ID=42, got %d", e.ID)
	}
}

func TestCreateEmployee_ValidationError(t *testing.T) {
	repo := &mockRepo{}
	svc := NewEmployeeService(repo, zap.NewNop())

	// отсутствует phone -> должен вернуться ValidationError
	err := svc.CreateEmployee(context.Background(), &domain.Employee{
		Name: "Bob", City: "Astana",
	})
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("expected *ValidationError, got %T (%v)", err, err)
	}
}

func TestGetEmployee_RepoError(t *testing.T) {
	repo := &mockRepo{
		GetByIDFn: func(ctx context.Context, id int) (*domain.Employee, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewEmployeeService(repo, zap.NewNop())

	if _, err := svc.GetEmployee(context.Background(), 99); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// Новые тесты для поиска
func TestSearchEmployees_Success(t *testing.T) {
	repo := &mockRepo{
		SearchEmployeesFn: func(ctx context.Context, searchQuery string) ([]*domain.Employee, error) {
			if searchQuery == "john" {
				return []*domain.Employee{
					{ID: 1, Name: "John Doe", Phone: "+77777777777", City: "Almaty"},
					{ID: 2, Name: "John Smith", Phone: "+77777777778", City: "Astana"},
				}, nil
			}
			return []*domain.Employee{}, nil
		},
	}
	svc := NewEmployeeService(repo, zap.NewNop())

	results, err := svc.SearchEmployees(context.Background(), "john")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Name != "John Doe" || results[1].Name != "John Smith" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestSearchEmployees_EmptyQuery(t *testing.T) {
	repo := &mockRepo{}
	svc := NewEmployeeService(repo, zap.NewNop())

	// Based on the actual service behavior, empty query returns validation error
	_, err := svc.SearchEmployees(context.Background(), "")
	if err == nil {
		t.Fatalf("expected validation error for empty query, got nil")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("expected *ValidationError, got %T (%v)", err, err)
	}
}

func TestSearchEmployees_WhitespaceQuery(t *testing.T) {
	repo := &mockRepo{}
	svc := NewEmployeeService(repo, zap.NewNop())

	// Test whitespace-only query (should be treated as empty after trimming)
	_, err := svc.SearchEmployees(context.Background(), "   ")
	if err == nil {
		t.Fatalf("expected validation error for whitespace query, got nil")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("expected *ValidationError, got %T (%v)", err, err)
	}
}

func TestSearchEmployees_ShortQuery(t *testing.T) {
	repo := &mockRepo{}
	svc := NewEmployeeService(repo, zap.NewNop())

	_, err := svc.SearchEmployees(context.Background(), "a")
	if err == nil {
		t.Fatalf("expected validation error for short query, got nil")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("expected *ValidationError, got %T (%v)", err, err)
	}
}

func TestSearchEmployees_LongQuery(t *testing.T) {
	repo := &mockRepo{}
	svc := NewEmployeeService(repo, zap.NewNop())

	// Create a query longer than 100 characters
	longQuery := strings.Repeat("a", 101)

	_, err := svc.SearchEmployees(context.Background(), longQuery)
	if err == nil {
		t.Fatalf("expected validation error for long query, got nil")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("expected *ValidationError, got %T (%v)", err, err)
	}
}

func TestSearchEmployees_ValidQuery(t *testing.T) {
	repo := &mockRepo{
		SearchEmployeesFn: func(ctx context.Context, searchQuery string) ([]*domain.Employee, error) {
			return []*domain.Employee{
				{ID: 1, Name: "Test User", Phone: "+77777777777", City: "Almaty"},
			}, nil
		},
	}
	svc := NewEmployeeService(repo, zap.NewNop())

	// Test with 2-character query (minimum valid)
	results, err := svc.SearchEmployees(context.Background(), "te")
	if err != nil {
		t.Fatalf("unexpected error for valid query: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchEmployees_RepoError(t *testing.T) {
	repo := &mockRepo{
		SearchEmployeesFn: func(ctx context.Context, searchQuery string) ([]*domain.Employee, error) {
			return nil, errors.New("database connection failed")
		},
	}
	svc := NewEmployeeService(repo, zap.NewNop())

	_, err := svc.SearchEmployees(context.Background(), "test")
	if err == nil {
		t.Fatalf("expected repo error, got nil")
	}
	if _, ok := err.(*ValidationError); ok {
		t.Fatalf("expected repo error, got ValidationError")
	}
}

func TestSearchEmployees_ByPhone(t *testing.T) {
	repo := &mockRepo{
		SearchEmployeesFn: func(ctx context.Context, searchQuery string) ([]*domain.Employee, error) {
			if searchQuery == "777" {
				return []*domain.Employee{
					{ID: 1, Name: "Alice Johnson", Phone: "+77777777777", City: "Almaty"},
				}, nil
			}
			return []*domain.Employee{}, nil
		},
	}
	svc := NewEmployeeService(repo, zap.NewNop())

	results, err := svc.SearchEmployees(context.Background(), "777")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !strings.Contains(results[0].Phone, "777") {
		t.Fatalf("expected phone to contain 777, got %s", results[0].Phone)
	}
}

func TestSearchEmployees_ByCity(t *testing.T) {
	repo := &mockRepo{
		SearchEmployeesFn: func(ctx context.Context, searchQuery string) ([]*domain.Employee, error) {
			if searchQuery == "almaty" {
				return []*domain.Employee{
					{ID: 1, Name: "Alice Brown", Phone: "+77777777779", City: "Almaty"},
					{ID: 2, Name: "Bob Green", Phone: "+77777777780", City: "Almaty"},
				}, nil
			}
			return []*domain.Employee{}, nil
		},
	}
	svc := NewEmployeeService(repo, zap.NewNop())

	results, err := svc.SearchEmployees(context.Background(), "almaty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, result := range results {
		if result.City != "Almaty" {
			t.Fatalf("expected city Almaty, got %s", result.City)
		}
	}
}

func TestSearchEmployees_NoResults(t *testing.T) {
	repo := &mockRepo{
		SearchEmployeesFn: func(ctx context.Context, searchQuery string) ([]*domain.Employee, error) {
			return []*domain.Employee{}, nil // No results
		},
	}
	svc := NewEmployeeService(repo, zap.NewNop())

	results, err := svc.SearchEmployees(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearchEmployees_CaseInsensitive(t *testing.T) {
	repo := &mockRepo{
		SearchEmployeesFn: func(ctx context.Context, searchQuery string) ([]*domain.Employee, error) {
			if strings.ToLower(searchQuery) == "john" {
				return []*domain.Employee{
					{ID: 1, Name: "john doe", Phone: "+77777777777", City: "almaty"},
				}, nil
			}
			return []*domain.Employee{}, nil
		},
	}
	svc := NewEmployeeService(repo, zap.NewNop())

	results, err := svc.SearchEmployees(context.Background(), "JOHN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "john doe" {
		t.Fatalf("unexpected name: %s", results[0].Name)
	}
}
