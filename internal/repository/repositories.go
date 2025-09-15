package repository

import (
	"context"
	"database/sql"
	"employer/internal/domain"

	"go.uber.org/zap"
)

// EmployeeRepository интерфейс для работы с БД
type EmployeeRepository interface {
	Create(ctx context.Context, employee *domain.Employee) error
	GetByID(ctx context.Context, id int) (*domain.Employee, error)
	GetAll(ctx context.Context) ([]*domain.Employee, error)
	Update(ctx context.Context, employee *domain.Employee) error
	Delete(ctx context.Context, id int) error
	GetByPhone(ctx context.Context, phone string) (*domain.Employee, error)
}

// Repositories объединяет все репозитории
type IRepositories struct {
	Employee EmployeeRepository
}

// NewRepositories создает все репозитории
func NewRepositories(db *sql.DB, logger *zap.Logger) *IRepositories {
	return &IRepositories{
		Employee: NewEmployeeRepository(db, logger),
	}
}
