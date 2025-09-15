package repository

import (
	"context"
	"database/sql"
	"employer/internal/domain"
	"fmt"

	"go.uber.org/zap"
)

type employeeRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewEmployeeRepository(db *sql.DB, logger *zap.Logger) *employeeRepository {
	return &employeeRepository{
		db:     db,
		logger: logger,
	}
}

// Create создает нового сотрудника в БД
func (r *employeeRepository) Create(ctx context.Context, employee *domain.Employee) error {
	query := `
		INSERT INTO employees (name, phone, city) 
		VALUES ($1, $2, $3) 
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query, employee.Name, employee.Phone, employee.City).Scan(&employee.ID)
	if err != nil {
		r.logger.Error("ошибка создания сотрудника", zap.Error(err))
		return fmt.Errorf("создание сотрудника: %w", err)
	}

	r.logger.Info("сотрудник создан", zap.Int("id", employee.ID))
	return nil
}

// GetByID получает сотрудника по ID
func (r *employeeRepository) GetByID(ctx context.Context, id int) (*domain.Employee, error) {
	employee := &domain.Employee{}
	query := `SELECT id, name, phone, city FROM employees WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&employee.ID, &employee.Name, &employee.Phone, &employee.City,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("сотрудник не найден", zap.Int("id", id))
			return nil, &NotFoundError{Entity: "employee", ID: id}
		}
		r.logger.Error("ошибка получения сотрудника", zap.Error(err), zap.Int("id", id))
		return nil, fmt.Errorf("получение сотрудника: %w", err)
	}

	return employee, nil
}

// GetAll получает всех сотрудников
func (r *employeeRepository) GetAll(ctx context.Context) ([]*domain.Employee, error) {
	query := `SELECT id, name, phone, city FROM employees ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("ошибка получения списка сотрудников", zap.Error(err))
		return nil, fmt.Errorf("получение списка сотрудников: %w", err)
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		employee := &domain.Employee{}
		err := rows.Scan(&employee.ID, &employee.Name, &employee.Phone, &employee.City)
		if err != nil {
			r.logger.Error("ошибка сканирования сотрудника", zap.Error(err))
			return nil, fmt.Errorf("сканирование сотрудника: %w", err)
		}
		employees = append(employees, employee)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("ошибка итерации по результатам", zap.Error(err))
		return nil, fmt.Errorf("итерация по результатам: %w", err)
	}

	r.logger.Info("получен список сотрудников", zap.Int("count", len(employees)))
	return employees, nil
}

// Update обновляет сотрудника
func (r *employeeRepository) Update(ctx context.Context, employee *domain.Employee) error {
	query := `
		UPDATE employees 
		SET name = $2, phone = $3, city = $4 
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, employee.ID, employee.Name, employee.Phone, employee.City)
	if err != nil {
		r.logger.Error("ошибка обновления сотрудника", zap.Error(err), zap.Int("id", employee.ID))
		return fmt.Errorf("обновление сотрудника: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("ошибка получения количества обновленных строк", zap.Error(err))
		return fmt.Errorf("получение количества обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("сотрудник для обновления не найден", zap.Int("id", employee.ID))
		return &NotFoundError{Entity: "employee", ID: employee.ID}
	}

	r.logger.Info("сотрудник обновлен", zap.Int("id", employee.ID))
	return nil
}

// Delete удаляет сотрудника
func (r *employeeRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM employees WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("ошибка удаления сотрудника", zap.Error(err), zap.Int("id", id))
		return fmt.Errorf("удаление сотрудника: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("ошибка получения количества удаленных строк", zap.Error(err))
		return fmt.Errorf("получение количества удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("сотрудник для удаления не найден", zap.Int("id", id))
		return &NotFoundError{Entity: "employee", ID: id}
	}

	r.logger.Info("сотрудник удален", zap.Int("id", id))
	return nil
}

// GetByPhone получает сотрудника по телефону
func (r *employeeRepository) GetByPhone(ctx context.Context, phone string) (*domain.Employee, error) {
	employee := &domain.Employee{}
	query := `SELECT id, name, phone, city FROM employees WHERE phone = $1`

	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&employee.ID, &employee.Name, &employee.Phone, &employee.City,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("сотрудник не найден по телефону", zap.String("phone", phone))
			return nil, &NotFoundError{Entity: "employee by phone", Data: phone}
		}
		r.logger.Error("ошибка получения сотрудника по телефону", zap.Error(err), zap.String("phone", phone))
		return nil, fmt.Errorf("получение сотрудника по телефону: %w", err)
	}

	return employee, nil
}

// NotFoundError ошибка "не найден"
type NotFoundError struct {
	Entity string
	ID     int
	Data   interface{}
}

func (e *NotFoundError) Error() string {
	if e.ID != 0 {
		return fmt.Sprintf("%s с ID %d не найден", e.Entity, e.ID)
	}
	return fmt.Sprintf("%s не найден: %v", e.Entity, e.Data)
}
