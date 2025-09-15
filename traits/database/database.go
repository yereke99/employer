package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Config интерфейс для конфигурации БД
type Config interface {
	GetDBHost() string
	GetDBPort() string
	GetDBUser() string
	GetDBPassword() string
	GetDBName() string
	GetDBSSLMode() string
}

// InitDatabase инициализирует подключение к PostgreSQL
func InitDatabase(cfg Config, logger *zap.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.GetDBHost(),
		cfg.GetDBPort(),
		cfg.GetDBUser(),
		cfg.GetDBPassword(),
		cfg.GetDBName(),
		cfg.GetDBSSLMode(),
	)

	logger.Info("подключение к БД",
		zap.String("host", cfg.GetDBHost()),
		zap.String("port", cfg.GetDBPort()),
		zap.String("dbname", cfg.GetDBName()),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия соединения с БД: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка пинга БД: %w", err)
	}

	logger.Info("подключение к БД успешно")
	return db, nil
}

// CreateTables создает необходимые таблицы
func CreateTables(db *sql.DB, logger *zap.Logger) error {
	logger.Info("создание таблиц")

	// Создание таблицы сотрудников
	if err := createEmployeesTable(db, logger); err != nil {
		return fmt.Errorf("ошибка создания таблицы employees: %w", err)
	}

	// Создание индексов
	if err := createIndexes(db, logger); err != nil {
		return fmt.Errorf("ошибка создания индексов: %w", err)
	}

	logger.Info("таблицы созданы успешно")
	return nil
}

// createEmployeesTable создает таблицу сотрудников
func createEmployeesTable(db *sql.DB, logger *zap.Logger) error {
	query := `
	CREATE TABLE IF NOT EXISTS employees (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		phone VARCHAR(50) NOT NULL UNIQUE,
		city VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(query); err != nil {
		logger.Error("ошибка создания таблицы employees", zap.Error(err))
		return err
	}

	logger.Info("таблица employees создана")
	return nil
}

// createIndexes создает индексы для оптимизации запросов
func createIndexes(db *sql.DB, logger *zap.Logger) error {
	indexes := []struct {
		name  string
		query string
	}{
		{
			name:  "idx_employees_phone",
			query: "CREATE INDEX IF NOT EXISTS idx_employees_phone ON employees(phone)",
		},
		{
			name:  "idx_employees_city",
			query: "CREATE INDEX IF NOT EXISTS idx_employees_city ON employees(city)",
		},
		{
			name:  "idx_employees_name",
			query: "CREATE INDEX IF NOT EXISTS idx_employees_name ON employees(name)",
		},
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx.query); err != nil {
			logger.Error("ошибка создания индекса",
				zap.String("index", idx.name),
				zap.Error(err),
			)
			return fmt.Errorf("создание индекса %s: %w", idx.name, err)
		}
		logger.Info("индекс создан", zap.String("name", idx.name))
	}

	return nil
}
