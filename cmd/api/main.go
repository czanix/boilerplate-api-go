package main

import (
	"context"
	"log"
	"os"

	"github.com/czanix/boilerplate-api-go/internal/application"
	"github.com/czanix/boilerplate-api-go/internal/infrastructure/postgres"
	"github.com/czanix/boilerplate-api-go/internal/presentation/handlers"
	"github.com/czanix/boilerplate-api-go/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// DI manual — sem container
	orderRepo := postgres.NewPgOrderRepository(pool)
	createOrderUC := application.NewCreateOrderUseCase(orderRepo)
	orderHandler := handlers.NewOrderHandler(createOrderUC)

	r := gin.Default()
	r.Use(middleware.SecurityHeaders())

	r.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(context.Background()); err != nil {
			c.JSON(503, gin.H{"status": "degraded", "database": "down"})
			return
		}
		c.JSON(200, gin.H{"status": "ok", "database": "up"})
	})

	api := r.Group("/api/v1")
	api.POST("/orders", orderHandler.Create)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server started on port %s", port)
	r.Run(":" + port)
}
