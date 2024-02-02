package driver

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-xray-sdk-go/xray"
	_ "github.com/lib/pq"
	"os"
	"strconv"
)

func NewDatabase() (*sql.DB, error) {
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	dbEndpoint := fmt.Sprintf("%s:%d", os.Getenv("DB_HOST"), dbPort)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error: " + err.Error())
	}
	authT, err := auth.BuildAuthToken(context.TODO(), dbEndpoint, os.Getenv("S3_REGION"), os.Getenv("DB_USER"), cfg.Credentials)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("AuthT:   ", authT)

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		os.Getenv("DB_HOST"), dbPort, os.Getenv("DB_USER"), authT, os.Getenv("DB_NAME"),
	)
	db, err := xray.SQLContext("postgres", dsn)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return db, nil
}
