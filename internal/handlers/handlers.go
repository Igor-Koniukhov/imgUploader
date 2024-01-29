package handlers

import (
	"imageAploaderS3/internal/config"
	"imageAploaderS3/internal/repository/dbrepo"
)

type Handlers struct {
	User
}

func NewHandlers(app *config.AppConfig, repo dbrepo.UserRepository) *Handlers {
	return &Handlers{
		User: NewUserHandlers(app, repo),
	}

}
