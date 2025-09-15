package domain

// Employee модель сотрудника
type Employee struct {
	ID    int    `json:"id" db:"id"`
	Name  string `json:"name" db:"name"`
	Phone string `json:"phone" db:"phone"`
	City  string `json:"city" db:"city"`
}

// DTOs для API
type CreateEmployeeRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	City  string `json:"city"`
}

type UpdateEmployeeRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	City  string `json:"city"`
}

type EmployeeResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	City  string `json:"city"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
