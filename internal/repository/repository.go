package repository

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
	"imageUploader/internal/repository/dbrepo"
)

type DBRepository struct {
	dbrepo.UserRepository
}

func NewRepository(db *sql.DB, pdb *redis.Client, rdb *redis.Client) *DBRepository {
	return &DBRepository{
		UserRepository: dbrepo.NewUserRepository(db, pdb, rdb),
	}

}
