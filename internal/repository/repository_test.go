package repository_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"employer/internal/domain"
	"employer/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func newRepo(t *testing.T) (*repository.IRepositories, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	repo := repository.NewRepositories(db, zap.NewNop())
	return repo, mock, func() { _ = db.Close() }
}

func TestCreate_SimpleSuccess(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	q := regexp.QuoteMeta(`
		INSERT INTO employees (name, phone, city) 
		VALUES ($1, $2, $3) 
		RETURNING id`)
	mock.ExpectQuery(q).
		WithArgs("Alice", "+7701", "Almaty").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	e := &domain.Employee{Name: "Alice", Phone: "+7701", City: "Almaty"}
	if err := repo.Employee.Create(context.Background(), e); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if e.ID != 10 {
		t.Fatalf("want ID=10 got %d", e.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	q := regexp.QuoteMeta(`SELECT id, name, phone, city FROM employees WHERE id = $1`)
	mock.ExpectQuery(q).WithArgs(404).WillReturnError(sql.ErrNoRows)

	_, err := repo.Employee.GetByID(context.Background(), 404)
	if err == nil {
		t.Fatalf("want error")
	}
	if _, ok := err.(*repository.NotFoundError); !ok {
		t.Fatalf("want NotFoundError got %T", err)
	}
}

// --- Search Tests ---

func TestSearchEmployees_Success(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	searchQuery := "john"
	searchPattern := "%john%"
	exactSearchPattern := "john%"

	q := regexp.QuoteMeta(`
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
		LIMIT 100`)

	rows := sqlmock.NewRows([]string{"id", "name", "phone", "city"}).
		AddRow(1, "John Doe", "+77777777777", "Almaty").
		AddRow(2, "John Smith", "+77777777778", "Astana")

	mock.ExpectQuery(q).
		WithArgs(searchPattern, exactSearchPattern).
		WillReturnRows(rows)

	results, err := repo.Employee.SearchEmployees(context.Background(), searchQuery)
	if err != nil {
		t.Fatalf("SearchEmployees: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Name != "John Doe" || results[0].Phone != "+77777777777" {
		t.Fatalf("unexpected first result: %+v", results[0])
	}

	if results[1].Name != "John Smith" || results[1].City != "Astana" {
		t.Fatalf("unexpected second result: %+v", results[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchEmployees_NoResults(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	searchQuery := "nonexistent"
	searchPattern := "%nonexistent%"
	exactSearchPattern := "nonexistent%"

	q := regexp.QuoteMeta(`
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
		LIMIT 100`)

	rows := sqlmock.NewRows([]string{"id", "name", "phone", "city"})

	mock.ExpectQuery(q).
		WithArgs(searchPattern, exactSearchPattern).
		WillReturnRows(rows)

	results, err := repo.Employee.SearchEmployees(context.Background(), searchQuery)
	if err != nil {
		t.Fatalf("SearchEmployees: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchEmployees_EmptyQuery(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	// Empty query should return empty results without database call
	results, err := repo.Employee.SearchEmployees(context.Background(), "")
	if err != nil {
		t.Fatalf("SearchEmployees: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty query, got %d", len(results))
	}

	// No database expectations since it should return early
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchEmployees_WhitespaceQuery(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	// Whitespace-only query should return empty results without database call
	results, err := repo.Employee.SearchEmployees(context.Background(), "   ")
	if err != nil {
		t.Fatalf("SearchEmployees: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results for whitespace query, got %d", len(results))
	}

	// No database expectations since it should return early
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchEmployees_ByPhone(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	searchQuery := "777"
	searchPattern := "%777%"
	exactSearchPattern := "777%"

	q := regexp.QuoteMeta(`
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
		LIMIT 100`)

	rows := sqlmock.NewRows([]string{"id", "name", "phone", "city"}).
		AddRow(5, "Alice Johnson", "+77777777777", "Almaty")

	mock.ExpectQuery(q).
		WithArgs(searchPattern, exactSearchPattern).
		WillReturnRows(rows)

	results, err := repo.Employee.SearchEmployees(context.Background(), searchQuery)
	if err != nil {
		t.Fatalf("SearchEmployees: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Phone != "+77777777777" {
		t.Fatalf("expected phone +77777777777, got %s", results[0].Phone)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchEmployees_ByCity(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	searchQuery := "almaty"
	searchPattern := "%almaty%"
	exactSearchPattern := "almaty%"

	q := regexp.QuoteMeta(`
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
		LIMIT 100`)

	rows := sqlmock.NewRows([]string{"id", "name", "phone", "city"}).
		AddRow(3, "Alice Brown", "+77777777779", "Almaty").
		AddRow(4, "Bob Green", "+77777777780", "Almaty")

	mock.ExpectQuery(q).
		WithArgs(searchPattern, exactSearchPattern).
		WillReturnRows(rows)

	results, err := repo.Employee.SearchEmployees(context.Background(), searchQuery)
	if err != nil {
		t.Fatalf("SearchEmployees: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Results should be ordered by name
	if results[0].Name != "Alice Brown" || results[1].Name != "Bob Green" {
		t.Fatalf("unexpected ordering: %s, %s", results[0].Name, results[1].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchEmployees_DatabaseError(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	searchQuery := "test"
	searchPattern := "%test%"
	exactSearchPattern := "test%"

	q := regexp.QuoteMeta(`
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
		LIMIT 100`)

	mock.ExpectQuery(q).
		WithArgs(searchPattern, exactSearchPattern).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.Employee.SearchEmployees(context.Background(), searchQuery)
	if err == nil {
		t.Fatalf("expected database error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchEmployees_ScanError(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	searchQuery := "test"
	searchPattern := "%test%"
	exactSearchPattern := "test%"

	q := regexp.QuoteMeta(`
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
		LIMIT 100`)

	// Return invalid data that will cause scan error
	rows := sqlmock.NewRows([]string{"id", "name", "phone", "city"}).
		AddRow("invalid_id", "John Doe", "+77777777777", "Almaty")

	mock.ExpectQuery(q).
		WithArgs(searchPattern, exactSearchPattern).
		WillReturnRows(rows)

	_, err := repo.Employee.SearchEmployees(context.Background(), searchQuery)
	if err == nil {
		t.Fatalf("expected scan error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSearchEmployees_CaseInsensitive(t *testing.T) {
	repo, mock, done := newRepo(t)
	defer done()

	searchQuery := "JOHN"
	searchPattern := "%JOHN%"
	exactSearchPattern := "JOHN%"

	q := regexp.QuoteMeta(`
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
		LIMIT 100`)

	rows := sqlmock.NewRows([]string{"id", "name", "phone", "city"}).
		AddRow(1, "john doe", "+77777777777", "almaty")

	mock.ExpectQuery(q).
		WithArgs(searchPattern, exactSearchPattern).
		WillReturnRows(rows)

	results, err := repo.Employee.SearchEmployees(context.Background(), searchQuery)
	if err != nil {
		t.Fatalf("SearchEmployees: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Name != "john doe" {
		t.Fatalf("unexpected name: %s", results[0].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
