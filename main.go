package main

import (
	"github.com/joho/godotenv"
	"imageAploaderS3/internal/config"
	"imageAploaderS3/internal/handlers"
	"imageAploaderS3/internal/render"
	"log"
	"net/http"
	"os"
)

var app config.AppConfig

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}
	log.Println("The API has started.")
	repo := handlers.NewRepository(&app)
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache: ", err)
	}

	app.TemplateCache = tc
	app.UseCache = false
	render.NewTemplates(&app)

	srv := &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: routes(repo),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
