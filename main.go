package main

import (
	"log"
	"os"
	"path/filepath"

	"go-restful-api/config"
	"go-restful-api/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 2. Initialize PostgreSQL connection
	if err := config.InitDB(); err != nil {
		log.Fatal("Database connection failed: ", err)
	}

	// 3. Close database connection pool when main exits normally
	defer func() {
		if err := config.CloseDB(); err != nil {
			log.Println("Failed to close database: ", err)
		}
	}()

	// 4. Create upload directories
	uploadDirs := []string{"articles", "users"}

	for _, dir := range uploadDirs {
		path := filepath.Join("uploads", dir)

		if err := os.MkdirAll(path, 0755); err != nil {
			log.Println("Failed to create upload directory: ", err)
			return
		}
	}

	// 5. Initialize Gin
	r := gin.Default()

	// 6. Serve uploaded files
	r.Static("/uploads", "./uploads")

	// 7. Register API routes
	routes.Serve(r)

	// 8. Configure server port
	port := os.Getenv("PORT")

	if port == "" {
		port = "5001"
	}

	// 9. Start HTTP server
	if err := r.Run(":" + port); err != nil {
		log.Println("Failed to start server: ", err)
	}
}
