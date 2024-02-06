package main

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"imageAploaderS3/internal/config"
	"imageAploaderS3/internal/handlers"
	"imageAploaderS3/internal/repository"
	"net/http"
)

func routes(app *config.AppConfig, db *sql.DB, primaryRC *redis.Client, readerRC *redis.Client) http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	dbRepo := repository.NewRepository(db, primaryRC, readerRC)
	repo := handlers.NewHandlers(app, dbRepo)
	mux.Group(func(r chi.Router) {
		r.Use(repo.AuthSet)
		r.Use(repo.UserDataSet)
		r.Use(repo.XRayMiddleware("imgUploader"))
		r.Get("/", repo.HomePage)
		r.Post("/upload", repo.UploadFileHandler)
		r.Get("/user-name", repo.GetUserName)
	})

	mux.Get("/signup", repo.AuthPageHandler)
	mux.Post("/signup", repo.Signup)
	mux.Get("/login", repo.LoginPageHandler)
	mux.Post("/login", repo.LoginHandler)
	mux.Get("/verify", repo.VerifyPageHandler)
	mux.Post("/verify", repo.VerifyHandler)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
