package service

import (
	"context"
	"employer/internal/domain"
	"employer/internal/repository"

	"go.uber.org/zap"
)

type EmployeeService interface {
	CreateEmployee(ctx context.Context, employee *domain.Employee) error
	GetEmployee(ctx context.Context, id int) (*domain.Employee, error)
	GetAllEmployees(ctx context.Context) ([]*domain.Employee, error)
	UpdateEmployee(ctx context.Context, employee *domain.Employee) error
	DeleteEmployee(ctx context.Context, id int) error
	SearchEmployees(ctx context.Context, searchQuery string) ([]*domain.Employee, error)
}

// Services объединяет все сервисы
type IServices struct {
	Employee EmployeeService
}

// NewServices создает все сервисы
func NewServices(repos *repository.IRepositories, logger *zap.Logger) *IServices {
	return &IServices{
		Employee: NewEmployeeService(repos.Employee, logger),
	}
}
