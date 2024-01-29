package repository

import (
	"database/sql"
	"imageAploaderS3/internal/repository/dbrepo"
)

type DBRepository struct {
	dbrepo.UserRepository
}

func NewRepository(db *sql.DB) *DBRepository {
	return &DBRepository{
		UserRepository: dbrepo.NewUserRepository(db),
	}

}
