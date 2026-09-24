package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	httpHandler "github.com/ran-jita/task-api/internal/handler/http"
	"github.com/ran-jita/task-api/internal/repository/postgres"
	"github.com/ran-jita/task-api/internal/usecase"
)

func main() {
	godotenv.Load()

	dsn := mustGetEnv("DATABASE_URL")
	jwtSecret := []byte(mustGetEnv("JWT_SECRET"))
	port := getEnv("PORT", "8080")

	db, err := postgres.NewDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	// --- Repository layer ---
	userRepo := postgres.NewUserRepository(db)
	taskRepo := postgres.NewTaskRepository(db)
	taskLogRepo := postgres.NewTaskLogRepository(db)
	idempotencyRepo := postgres.NewIdempotencyRepository(db)
	txManager := postgres.NewTxManager(db)

	// --- Usecase layer ---
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtSecret)
	taskUsecase := usecase.NewTaskUsecase(taskRepo)
	assignUsecase := usecase.NewAssignUsecase(taskRepo, taskLogRepo, userRepo, txManager)
	idempotencyUsecase := usecase.NewIdempotencyUsecase(idempotencyRepo)

	// --- Handler layer ---
	authHandler := httpHandler.NewAuthHandler(authUsecase)
	taskHandler := httpHandler.NewTaskHandler(taskUsecase, idempotencyUsecase)
	assignHandler := httpHandler.NewAssignHandler(assignUsecase)

	// --- Router ---
	router := httpHandler.NewRouter(authHandler, taskHandler, assignHandler, jwtSecret)

	log.Printf("server starting on port %s", port)
	if err := router.Start(":" + port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("missing required environment variable: %s", key)
	}
	return val
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
