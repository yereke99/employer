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
	"syscall"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
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

	zapLogger.Info("запуск приложения TezJet",
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
	// Swagger UI (по адресу /swagger/index.html)
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Регистрация маршрутов для API сотрудников
	employeeHandler.RegisterRoutes(router)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
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

	// Запуск сервера в отдельной горутине
	go func() {
		zapLogger.Info("HTTP server started", zap.String("address", cfg.GetServerAddress()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Error("failed to start HTTP server", zap.Error(err))
			cancel()
		}
	}()

	// Ожидание сигнала завершения
	select {
	case <-stop:
		zapLogger.Info("shutdown signal received")
	case <-ctx.Done():
		zapLogger.Info("context cancelled")
	}

	// Graceful shutdown сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	zapLogger.Info("shutting down server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("failed to shutdown server", zap.Error(err))
	}

	zapLogger.Info("application stopped successfully")
}
