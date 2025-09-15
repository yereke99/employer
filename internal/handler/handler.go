package handler

import (
	"encoding/json"
	"net/http"
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
// CreateEmployee создает нового сотрудника
// POST /api/employees
// @Summary      Создать сотрудника
// @Description  Создает нового сотрудника
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        request  body      domain.CreateEmployeeRequest true  "Данные сотрудника"
// @Success      201      {object}  domain.EmployeeResponse
// @Failure      400      {object}  domain.ErrorResponse        "некорректный JSON / ошибка валидации"
// @Failure      500      {object}  domain.ErrorResponse        "внутренняя ошибка сервера"
// @Router       /api/employees [post]
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
// @Summary      Получить сотрудника
// @Description  Возвращает сотрудника по ID
// @Tags         employees
// @Produce      json
// @Param        id   path      int  true  "Employee ID"
// @Success      200  {object}  domain.EmployeeResponse
// @Failure      400  {object}  domain.ErrorResponse  "некорректный ID"
// @Failure      404  {object}  domain.ErrorResponse  "сотрудник не найден"
// @Failure      500  {object}  domain.ErrorResponse  "внутренняя ошибка сервера"
// @Router       /api/employees/{id} [get]
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
// @Summary      Получить всех сотрудников
// @Description  Возвращает список сотрудников
// @Tags         employees
// @Produce      json
// @Success      200  {array}   domain.EmployeeResponse
// @Failure      500  {object}  domain.ErrorResponse  "внутренняя ошибка сервера"
// @Router       /api/employees [get]
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
// @Summary      Обновить сотрудника
// @Description  Обновляет данные сотрудника по ID
// @Tags         employees
// @Accept       json
// @Produce      json
// @Param        id       path      int                            true  "Employee ID"
// @Param        request  body      domain.UpdateEmployeeRequest   true  "Данные сотрудника"
// @Success      200      {object}  domain.EmployeeResponse
// @Failure      400      {object}  domain.ErrorResponse  "некорректный ID / некорректный JSON / ошибка валидации"
// @Failure      404      {object}  domain.ErrorResponse  "сотрудник не найден"
// @Failure      500      {object}  domain.ErrorResponse  "внутренняя ошибка сервера"
// @Router       /api/employees/{id} [put]
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
// @Summary      Удалить сотрудника
// @Description  Удаляет сотрудника по ID
// @Tags         employees
// @Produce      json
// @Param        id   path      int  true  "Employee ID"
// @Success      204  "No Content"
// @Failure      400  {object}  domain.ErrorResponse  "некорректный ID"
// @Failure      404  {object}  domain.ErrorResponse  "сотрудник не найден"
// @Failure      500  {object}  domain.ErrorResponse  "внутренняя ошибка сервера"
// @Router       /api/employees/{id} [delete]
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
