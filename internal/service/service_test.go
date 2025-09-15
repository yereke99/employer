package service

import (
	"context"
	"errors"
	"testing"

	"employer/internal/domain"
	"employer/internal/repository"

	"go.uber.org/zap"
)

// мок репозитория под интерфейс repository.EmployeeRepository
type mockRepo struct {
	CreateFn     func(ctx context.Context, e *domain.Employee) error
	GetByIDFn    func(ctx context.Context, id int) (*domain.Employee, error)
	GetAllFn     func(ctx context.Context) ([]*domain.Employee, error)
	UpdateFn     func(ctx context.Context, e *domain.Employee) error
	DeleteFn     func(ctx context.Context, id int) error
	GetByPhoneFn func(ctx context.Context, phone string) (*domain.Employee, error)
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

// Убедись, что тип удовлетворяет интерфейсу (компиляционная проверка)
var _ repository.EmployeeRepository = (*mockRepo)(nil)

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
