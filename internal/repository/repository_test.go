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
