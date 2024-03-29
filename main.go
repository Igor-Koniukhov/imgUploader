package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"imageUploader/driver"
	"imageUploader/internal/config"
	"imageUploader/internal/render"
	"log"
	"net/http"
	"os"
)

var app config.AppConfig
var ctx = context.Background()

func main() {
	context.WithValue(ctx, "something", "value888888")
	err := godotenv.Load()
	if err != nil {
		log.Println("Env load error: ", err)
	}

	primaryRedisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("PRIMARY_ENDPOINT"),
		Password: "",
		DB:       0,
	})
	ctx, subSegRedis := xray.BeginSubsegment(ctx, "redis-operation")
	defer func() {
		if subSegRedis != nil {
			subSegRedis.Close(nil)
		}
	}()
	readerRedisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("READER_ENDPOINT"),
		Password: "",
		DB:       0,
	})

	pong, err := primaryRedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Error connecting to Redis:", err, pong)
	}
	fmt.Println("Connected to Redis:", pong)
	db, err := driver.NewDatabase(ctx)
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
		Handler: routes(&app, db, primaryRedisClient, readerRedisClient),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
