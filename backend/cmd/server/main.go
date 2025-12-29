package main

import (
	"database/sql"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/cloudinary/cloudinary-go/v2"
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
	cld, err := cloudinary.NewFromParams("dh7bridgn", "958111531242264", "QFTx26bP9MxnA_rzDpZUrcezwxI")
	if err != nil {
		log.Fatal("Cloudinary init error:", err)
	}
	// ===== init repos & services =====
	userRepo := user.NewRepository(db)
	authService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(authService)

	urlRepo := url.NewRepository(db)
	clickRepo := click.NewRepository(db)
	clickService := click.NewService(clickRepo)
	urlService := url.NewService(urlRepo, clickService, cld)
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

	log.Println("🚀 Server running at :8080")
	r.Run(":8080")
}
