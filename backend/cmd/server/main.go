package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
	"url-shortener/internal/auth"
	"url-shortener/internal/click"
	"url-shortener/internal/url"
	"url-shortener/internal/user"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system env / Fly secrets")
	}

	// Get env vars
	dbURL := mustGetEnv("DATABASE_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	frontendURL := mustGetEnv("FRONTEND_URL")

	// Connect DB với pgx và config tốt hơn
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("❌ Failed to open DB:", err)
	}
	defer db.Close()

	
	db.SetMaxOpenConns(10)           
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Ping DB với timeout
	log.Println("🔄 Connecting to database...")
	if err := db.Ping(); err != nil {
		log.Printf("❌ Cannot connect to DB: %v", err)
		log.Fatal(err)
	}
	log.Println("✅ DB connection successful")

	// Gin setup
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Cloudinary
	cld, err := cloudinary.NewFromParams(
		mustGetEnv("CLOUDINARY_CLOUD_NAME"),
		mustGetEnv("CLOUDINARY_API_KEY"),
		mustGetEnv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		log.Fatal("❌ Cloudinary init error:", err)
	}

	// Repositories & Services
	userRepo := user.NewRepository(db)
	authService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(authService)

	urlRepo := url.NewRepository(db)
	clickRepo := click.NewRepository(db)
	clickService := click.NewService(clickRepo)
	urlService := url.NewService(urlRepo, clickService, cld)
	urlHandler := url.NewHandler(urlService)

	// Routes
	api := r.Group("/api")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		api.POST("/urls",
			auth.Middleware(auth.JWTService),
			urlHandler.CreateShortURL,
		)

		api.GET("/urls",
			auth.Middleware(auth.JWTService),
			urlHandler.ListURLs,
		)

		api.GET("/urls/stats",
			auth.Middleware(auth.JWTService),
			urlHandler.UserStats,
		)

		api.DELETE("/urls/:id",
			auth.Middleware(auth.JWTService),
			urlHandler.DeleteURL,
		)
	}

	r.GET("/:code", urlHandler.Redirect)

	log.Println("🚀 Server running at :" + port)
	r.Run(":" + port)
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("❌ Missing env var: %s", key)
	}
	return value
}