package main

import (
	"database/sql"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"url-shortener/internal/auth"
	"url-shortener/internal/click"
	"url-shortener/internal/url"
	"url-shortener/internal/user"
)

func main() {
	db, err := sql.Open(
		"postgres",
		"host=localhost port=5432 user=postgres password=123 dbname=urlshortener sslmode=disable",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.Default()

	// ✅ CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// ===== init repos & services =====
	userRepo := user.NewRepository(db)
	authService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(authService)

	urlRepo := url.NewRepository(db)
	clickRepo := click.NewRepository(db)
	clickService := click.NewService(clickRepo)
	urlService := url.NewService(urlRepo, clickService)
	urlHandler := url.NewHandler(urlService)

	// ===== routes =====
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
			urlHandler.List,
		)

		api.GET("/urls/stats",
			auth.Middleware(auth.JWTService),
			urlHandler.UserStats,
		)
	}

	// redirect short url
	r.GET("/:code", urlHandler.Redirect)

	log.Println("🚀 Server running at :8080")
	r.Run(":8080")
}
