package service

import (
	"context"
	"employer/internal/domain"
	"employer/internal/repository"

	"go.uber.org/zap"
)

// employeeService реализация сервиса
type employeeService struct {
	repo   repository.EmployeeRepository
	logger *zap.Logger
}

// NewEmployeeService создает новый сервис для сотрудников
func NewEmployeeService(repo repository.EmployeeRepository, logger *zap.Logger) *employeeService {
	return &employeeService{
		repo:   repo,
		logger: logger,
	}
}

// CreateEmployee создает нового сотрудника
func (s *employeeService) CreateEmployee(ctx context.Context, employee *domain.Employee) error {
	s.logger.Info("создание сотрудника", zap.String("name", employee.Name))

	if err := s.validateEmployee(employee); err != nil {
		s.logger.Error("валидация сотрудника", zap.Error(err))
		return err
	}

	return s.repo.Create(ctx, employee)
}

// GetEmployee получает сотрудника по ID
func (s *employeeService) GetEmployee(ctx context.Context, id int) (*domain.Employee, error) {
	s.logger.Info("получение сотрудника", zap.Int("id", id))
	return s.repo.GetByID(ctx, id)
}

// GetAllEmployees получает всех сотрудников
func (s *employeeService) GetAllEmployees(ctx context.Context) ([]*domain.Employee, error) {
	s.logger.Info("получение всех сотрудников")
	return s.repo.GetAll(ctx)
}

// UpdateEmployee обновляет сотрудника
func (s *employeeService) UpdateEmployee(ctx context.Context, employee *domain.Employee) error {
	s.logger.Info("обновление сотрудника", zap.Int("id", employee.ID))

	if err := s.validateEmployee(employee); err != nil {
		s.logger.Error("валидация сотрудника", zap.Error(err))
		return err
	}

	return s.repo.Update(ctx, employee)
}

// DeleteEmployee удаляет сотрудника
func (s *employeeService) DeleteEmployee(ctx context.Context, id int) error {
	s.logger.Info("удаление сотрудника", zap.Int("id", id))
	return s.repo.Delete(ctx, id)
}

// validateEmployee валидирует данные сотрудника
func (s *employeeService) validateEmployee(employee *domain.Employee) error {
	if employee.Name == "" {
		return &ValidationError{Field: "name", Message: "имя обязательно"}
	}
	if employee.Phone == "" {
		return &ValidationError{Field: "phone", Message: "телефон обязателен"}
	}
	if employee.City == "" {
		return &ValidationError{Field: "city", Message: "город обязателен"}
	}
	return nil
}

// ValidationError ошибка валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}
