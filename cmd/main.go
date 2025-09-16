package main

import (
	"context"
	"employer/config"
	"employer/internal/handler"
	"employer/internal/repository"
	"employer/internal/service"
	"employer/traits/database"
	"employer/traits/logger"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	// Инициализация логгера
	zapLogger, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync()

	// Загрузка конфигурации
	cfg, err := config.NewConfig()
	if err != nil {
		zapLogger.Error("ошибка инициализации конфига", zap.Error(err))
		return
	}

	// Валидация конфигурации
	if err := cfg.ValidateConfig(); err != nil {
		zapLogger.Error("некорректная конфигурация", zap.Error(err))
		return
	}

	zapLogger.Info("запуск приложения Emplyee",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
		zap.String("db_name", cfg.DBName),
	)

	// Инициализация базы данных
	db, err := database.InitDatabase(cfg, zapLogger)
	if err != nil {
		zapLogger.Error("ошибка инициализации БД", zap.Error(err))
		return
	}
	defer db.Close()

	// Создание таблиц БД
	if err := database.CreateTables(db, zapLogger); err != nil {
		zapLogger.Error("ошибка создания таблиц", zap.Error(err))
		return
	}

	// Инициализация репозиториев
	repos := repository.NewRepositories(db, zapLogger)

	// Инициализация сервисов
	services := service.NewServices(repos, zapLogger)

	// Создание HTTP обработчиков
	employeeHandler := handler.NewEmployeeHandler(services.Employee, zapLogger)

	// Настройка маршрутизации
	router := mux.NewRouter()

	// CORS middleware для API запросов
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Применяем CORS только к API запросам
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}

	// Middleware для логирования запросов
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)

			// Логируем только важные запросы (не статические файлы)
			if !strings.HasPrefix(r.URL.Path, "/static/") && r.URL.Path != "/health" {
				zapLogger.Info("HTTP request",
					zap.String("method", r.Method),
					zap.String("url", r.URL.Path),
					zap.String("remote_addr", getClientIP(r)),
					zap.Duration("duration", time.Since(start)),
				)
			}
		})
	}

	// Применение middleware
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware)

	// Регистрация маршрутов для API сотрудников
	employeeHandler.RegisterRoutes(router)

	// Статические файлы (CSS, JS, изображения)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Функция для обслуживания HTML страницы
	serveEmployeePage := func(w http.ResponseWriter, r *http.Request) {
		htmlPath := "./static/employee.html"

		// Устанавливаем заголовки
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Обслуживаем файл
		http.ServeFile(w, r, htmlPath)

		zapLogger.Info("employee page served",
			zap.String("remote_addr", getClientIP(r)),
			zap.String("path", r.URL.Path),
		)
	}

	// Маршруты для веб-интерфейса
	router.HandleFunc("/", serveEmployeePage).Methods("GET")
	router.HandleFunc("/employees", serveEmployeePage).Methods("GET")
	router.HandleFunc("/employee", serveEmployeePage).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK","service":"Employee Management"}`))
	}).Methods("GET")

	// Debug endpoint для проверки маршрутов
	router.HandleFunc("/debug/routes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		routes := []string{
			"GET /",
			"GET /employees",
			"GET /employee",
			"GET /health",
			"GET /static/{file}",
			"GET /api/employees",
			"POST /api/employees",
			"GET /api/employees/{id}",
			"PUT /api/employees/{id}",
			"DELETE /api/employees/{id}",
		}

		response := `{"available_routes":["` + strings.Join(routes, `","`) + `"]}`
		w.Write([]byte(response))
	}).Methods("GET")

	// Создание HTTP сервера
	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.GetServerAddress(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Настройка graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для получения сигналов ОС
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Проверяем существование статических файлов при запуске
	checkStaticFiles(zapLogger)

	// Запуск сервера в отдельной горутине
	go func() {
		zapLogger.Info("🚀 Web App HTTP server started",
			zap.String("local_address", cfg.GetServerAddress()),
			zap.String("environment", cfg.Environment),
		)
		zapLogger.Info("📱 Employee Management Web Interface: https://meily.kz")
		zapLogger.Info("🔧 API Endpoints: https://meily.kz/api/employees")
		zapLogger.Info("🏥 Health Check: https://meily.kz/health")
		zapLogger.Info("🐛 Debug Routes: https://meily.kz/debug/routes")
		zapLogger.Info("📁 Static Files: https://meily.kz/static/")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Error("failed to start HTTP server", zap.Error(err))
			cancel()
		}
	}()

	// Ожидание сигнала завершения
	select {
	case <-stop:
		zapLogger.Info("🛑 shutdown signal received")
	case <-ctx.Done():
		zapLogger.Info("🛑 context cancelled")
	}

	// Graceful shutdown сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	zapLogger.Info("🔄 shutting down server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("❌ failed to shutdown server", zap.Error(err))
	}

	zapLogger.Info("✅ Application stopped successfully")
}

// checkStaticFiles проверяет существование необходимых статических файлов
func checkStaticFiles(logger *zap.Logger) {
	staticDir := "./static"
	employeeHTML := "./static/employee.html"

	// Проверяем папку static
	if _, err := os.Stat(employeeHTML); os.IsNotExist(err) {
		logger.Warn("static directory does not exist, creating it", zap.String("path", staticDir))
		if err := os.MkdirAll(staticDir, 0755); err != nil {
			logger.Error("failed to create static directory", zap.Error(err))
		}
	}

	// Проверяем employee.html
	if _, err := os.Stat(employeeHTML); os.IsNotExist(err) {
		logger.Warn("employee.html not found",
			zap.String("expected_path", employeeHTML),
			zap.String("solution", "Please create static/employee.html file"),
		)
	} else {
		logger.Info("✅ employee.html found", zap.String("path", employeeHTML))
	}
}

// getClientIP получает реальный IP клиента с учетом прокси
func getClientIP(r *http.Request) string {
	// Проверяем заголовки прокси в порядке приоритета
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// Берем первый IP из списка (клиентский)
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}

	if xClientIP := r.Header.Get("X-Client-IP"); xClientIP != "" {
		return strings.TrimSpace(xClientIP)
	}

	// Возвращаем RemoteAddr как последний вариант
	return r.RemoteAddr
}
