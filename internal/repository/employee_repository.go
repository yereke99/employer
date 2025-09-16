package repository

import (
	"context"
	"database/sql"
	"employer/internal/domain"
	"fmt"
	"strings"

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

// SearchEmployees ищет сотрудников по имени, телефону или городу
func (r *employeeRepository) SearchEmployees(ctx context.Context, searchQuery string) ([]*domain.Employee, error) {
	// Валидация входных данных
	searchQuery = strings.TrimSpace(searchQuery)
	if searchQuery == "" {
		r.logger.Warn("пустой поисковый запрос")
		return []*domain.Employee{}, nil
	}

	// SQL запрос с поиском по всем полям
	query := `
		SELECT id, name, phone, city 
		FROM employees 
		WHERE LOWER(name) LIKE LOWER($1) 
		   OR LOWER(phone) LIKE LOWER($1) 
		   OR LOWER(city) LIKE LOWER($1)
		ORDER BY 
			CASE 
				WHEN LOWER(name) LIKE LOWER($2) THEN 1
				WHEN LOWER(phone) LIKE LOWER($2) THEN 2
				WHEN LOWER(city) LIKE LOWER($2) THEN 3
				ELSE 4
			END,
			name ASC
		LIMIT 100`

	searchPattern := "%" + searchQuery + "%"
	exactSearchPattern := searchQuery + "%"

	rows, err := r.db.QueryContext(ctx, query, searchPattern, exactSearchPattern)
	if err != nil {
		r.logger.Error("ошибка выполнения поискового запроса",
			zap.Error(err),
			zap.String("search_query", searchQuery))
		return nil, fmt.Errorf("поиск сотрудников: %w", err)
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		employee := &domain.Employee{}
		err := rows.Scan(&employee.ID, &employee.Name, &employee.Phone, &employee.City)
		if err != nil {
			r.logger.Error("ошибка сканирования результата поиска", zap.Error(err))
			return nil, fmt.Errorf("сканирование результата поиска: %w", err)
		}
		employees = append(employees, employee)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("ошибка итерации по результатам поиска", zap.Error(err))
		return nil, fmt.Errorf("итерация по результатам поиска: %w", err)
	}

	r.logger.Info("поиск сотрудников выполнен",
		zap.String("search_query", searchQuery),
		zap.Int("results_count", len(employees)))

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

// GetEmployeeStats получает статистику по сотрудникам (дополнительный метод)
func (r *employeeRepository) GetEmployeeStats(ctx context.Context) (*EmployeeStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_count,
			COUNT(DISTINCT city) as cities_count,
			(SELECT city FROM employees GROUP BY city ORDER BY COUNT(*) DESC LIMIT 1) as most_common_city
		FROM employees`

	stats := &EmployeeStats{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalCount,
		&stats.CitiesCount,
		&stats.MostCommonCity,
	)

	if err != nil {
		r.logger.Error("ошибка получения статистики сотрудников", zap.Error(err))
		return nil, fmt.Errorf("получение статистики сотрудников: %w", err)
	}

	r.logger.Info("статистика сотрудников получена",
		zap.Int("total", stats.TotalCount),
		zap.Int("cities", stats.CitiesCount))

	return stats, nil
}

// GetEmployeesByCity получает сотрудников по городу
func (r *employeeRepository) GetEmployeesByCity(ctx context.Context, city string) ([]*domain.Employee, error) {
	query := `SELECT id, name, phone, city FROM employees WHERE LOWER(city) = LOWER($1) ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, city)
	if err != nil {
		r.logger.Error("ошибка получения сотрудников по городу",
			zap.Error(err),
			zap.String("city", city))
		return nil, fmt.Errorf("получение сотрудников по городу: %w", err)
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		employee := &domain.Employee{}
		err := rows.Scan(&employee.ID, &employee.Name, &employee.Phone, &employee.City)
		if err != nil {
			r.logger.Error("ошибка сканирования сотрудника по городу", zap.Error(err))
			return nil, fmt.Errorf("сканирование сотрудника: %w", err)
		}
		employees = append(employees, employee)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("ошибка итерации по сотрудникам города", zap.Error(err))
		return nil, fmt.Errorf("итерация по сотрудникам: %w", err)
	}

	r.logger.Info("получены сотрудники по городу",
		zap.String("city", city),
		zap.Int("count", len(employees)))

	return employees, nil
}

// CheckPhoneExists проверяет существование телефона
func (r *employeeRepository) CheckPhoneExists(ctx context.Context, phone string, excludeID ...int) (bool, error) {
	var query string
	var args []interface{}

	if len(excludeID) > 0 {
		query = `SELECT EXISTS(SELECT 1 FROM employees WHERE phone = $1 AND id != $2)`
		args = []interface{}{phone, excludeID[0]}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM employees WHERE phone = $1)`
		args = []interface{}{phone}
	}

	var exists bool
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		r.logger.Error("ошибка проверки существования телефона",
			zap.Error(err),
			zap.String("phone", phone))
		return false, fmt.Errorf("проверка существования телефона: %w", err)
	}

	return exists, nil
}

// EmployeeStats статистика сотрудников
type EmployeeStats struct {
	TotalCount     int    `json:"total_count"`
	CitiesCount    int    `json:"cities_count"`
	MostCommonCity string `json:"most_common_city"`
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
