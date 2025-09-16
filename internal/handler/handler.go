package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"employer/internal/domain"
	"employer/internal/service"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// EmployeeHandler обработчик для API сотрудников
type EmployeeHandler struct {
	service service.EmployeeService
	logger  *zap.Logger
}

// NewEmployeeHandler создает новый обработчик для сотрудников
func NewEmployeeHandler(service service.EmployeeService, logger *zap.Logger) *EmployeeHandler {
	return &EmployeeHandler{
		service: service,
		logger:  logger,
	}
}

// CreateEmployee создает нового сотрудника
// POST /api/employees
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("ошибка декодирования запроса", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	employee := &domain.Employee{
		Name:  req.Name,
		Phone: req.Phone,
		City:  req.City,
	}

	if err := h.service.CreateEmployee(r.Context(), employee); err != nil {
		if _, ok := err.(*service.ValidationError); ok {
			h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("ошибка создания сотрудника", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	response := &domain.EmployeeResponse{
		ID:    employee.ID,
		Name:  employee.Name,
		Phone: employee.Phone,
		City:  employee.City,
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// GetEmployee получает сотрудника по ID
// GET /api/employees/{id}
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "некорректный ID")
		return
	}

	employee, err := h.service.GetEmployee(r.Context(), id)
	if err != nil {
		if h.isNotFoundError(err) {
			h.writeErrorResponse(w, http.StatusNotFound, "сотрудник не найден")
			return
		}
		h.logger.Error("ошибка получения сотрудника", zap.Error(err), zap.Int("id", id))
		h.writeErrorResponse(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	response := &domain.EmployeeResponse{
		ID:    employee.ID,
		Name:  employee.Name,
		Phone: employee.Phone,
		City:  employee.City,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAllEmployees получает всех сотрудников
// GET /api/employees
func (h *EmployeeHandler) GetAllEmployees(w http.ResponseWriter, r *http.Request) {
	employees, err := h.service.GetAllEmployees(r.Context())
	if err != nil {
		h.logger.Error("ошибка получения списка сотрудников", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	response := make([]*domain.EmployeeResponse, len(employees))
	for i, emp := range employees {
		response[i] = &domain.EmployeeResponse{
			ID:    emp.ID,
			Name:  emp.Name,
			Phone: emp.Phone,
			City:  emp.City,
		}
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateEmployee обновляет сотрудника
// PUT /api/employees/{id}
func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "некорректный ID")
		return
	}

	var req domain.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("ошибка декодирования запроса", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	employee := &domain.Employee{
		ID:    id,
		Name:  req.Name,
		Phone: req.Phone,
		City:  req.City,
	}

	if err := h.service.UpdateEmployee(r.Context(), employee); err != nil {
		if _, ok := err.(*service.ValidationError); ok {
			h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if h.isNotFoundError(err) {
			h.writeErrorResponse(w, http.StatusNotFound, "сотрудник не найден")
			return
		}
		h.logger.Error("ошибка обновления сотрудника", zap.Error(err), zap.Int("id", id))
		h.writeErrorResponse(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	response := &domain.EmployeeResponse{
		ID:    employee.ID,
		Name:  employee.Name,
		Phone: employee.Phone,
		City:  employee.City,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// DeleteEmployee удаляет сотрудника
// DELETE /api/employees/{id}
func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "некорректный ID")
		return
	}

	if err := h.service.DeleteEmployee(r.Context(), id); err != nil {
		if h.isNotFoundError(err) {
			h.writeErrorResponse(w, http.StatusNotFound, "сотрудник не найден")
			return
		}
		h.logger.Error("ошибка удаления сотрудника", zap.Error(err), zap.Int("id", id))
		h.writeErrorResponse(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes регистрирует маршруты для API сотрудников
func (h *EmployeeHandler) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/employees").Subrouter()

	api.HandleFunc("", h.CreateEmployee).Methods("POST")
	api.HandleFunc("", h.GetAllEmployees).Methods("GET")
	api.HandleFunc("/{id:[0-9]+}", h.GetEmployee).Methods("GET")
	api.HandleFunc("/{id:[0-9]+}", h.UpdateEmployee).Methods("PUT")
	api.HandleFunc("/{id:[0-9]+}", h.DeleteEmployee).Methods("DELETE")
}

// ServeEmployeePage обслуживает страницу управления сотрудниками
// GET /
// GET /employees
func (h *EmployeeHandler) ServeEmployeePage(w http.ResponseWriter, r *http.Request) {
	// Путь к статическому файлу
	staticPath := filepath.Join("static", "employee.html")

	// Устанавливаем заголовки
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Обслуживаем файл
	http.ServeFile(w, r, staticPath)

	h.logger.Info("employee page served",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	)
}

// Вспомогательные методы
func (h *EmployeeHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *EmployeeHandler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, &domain.ErrorResponse{Error: message})
}

func (h *EmployeeHandler) isNotFoundError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "не найден") ||
		strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "notfound")
}
