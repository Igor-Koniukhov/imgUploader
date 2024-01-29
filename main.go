package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"imageAploaderS3/driver"
	"imageAploaderS3/internal/config"
	"imageAploaderS3/internal/render"
	"log"
	"net/http"
	"os"
)

var app config.AppConfig

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Env load error: ", err)
	}
	db, err := driver.NewDatabase()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println("Error db connection: ", err)
		}
	}(db)
	if err != nil {
		log.Println("DB initiation error: ", err)
	}
	log.Println("The API has started.")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache: ", err)
	}

	app.TemplateCache = tc
	app.UseCache = false
	render.NewTemplates(&app)

	srv := &http.Server{
		Addr:    os.Getenv("PORT"),
		Handler: routes(&app, db),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
