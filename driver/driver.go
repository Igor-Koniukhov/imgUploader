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

func NewDatabase(ctx context.Context) (*sql.DB, error) {
	ctx, subSeg := xray.BeginSubsegment(ctx, os.Getenv("DB_NAME"))
	defer func() {
		if subSeg != nil {
			subSeg.Close(nil)
		}
	}()
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("error converting DB_PORT: %w", err)
	}

	dbEndpoint := fmt.Sprintf("%s:%d", os.Getenv("DB_HOST"), dbPort)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}

	authToken, err := auth.BuildAuthToken(ctx, dbEndpoint, os.Getenv("S3_REGION"), os.Getenv("DB_USER"), cfg.Credentials)
	if err != nil {
		return nil, fmt.Errorf("error building auth token: %w", err)
	}
	fmt.Println(authToken)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		os.Getenv("DB_HOST"), dbPort, os.Getenv("DB_USER"), authToken, os.Getenv("DB_NAME"),
	)

	db, err := xray.SQLContext("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error initializing DB with X-Ray: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging DB: %w", err)
	}

	return db, nil
}
