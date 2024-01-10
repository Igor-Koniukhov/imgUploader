package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"imageAploaderS3/internal/handlers"
	"net/http"
)

func routes(repo *handlers.Repository) http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)

	mux.Get("/signup", repo.AuthPageHandler)
	mux.Post("/signup", repo.Signup)
	mux.Get("/login", repo.LoginPageHandler)
	mux.Post("/login", repo.LoginHandler)

	mux.Group(func(r chi.Router) {
		r.Use(repo.AuthSet)
		r.Get("/", repo.HomePage)
		r.Get("/verify", repo.VerifyPageHandler)
		r.Post("/verify", repo.VerifyHandler)
		r.Post("/upload", repo.UploadFileHandler)
	})

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}