package main

import (
	"database/sql"
	"log"
	"net/http"

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

	// Repositories
	userRepo := user.NewRepository(db)
	authService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(authService)

	urlRepo := url.NewRepository(db)
	clickRepo := click.NewRepository(db)
	clickService := click.NewService(clickRepo)
	urlService := url.NewService(urlRepo, clickService)
	urlHandler := url.NewHandler(urlService)

	// Routes
	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/login", authHandler.Login)

	// URL (with auth middleware)
	mux.Handle("/api/urls", auth.Middleware(auth.JWTService, urlHandler))
	mux.Handle("/api/urls/stats", auth.Middleware(auth.JWTService, http.HandlerFunc(urlHandler.UserStats)))

	// Redirect
	mux.HandleFunc("/", urlHandler.Redirect)

	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
