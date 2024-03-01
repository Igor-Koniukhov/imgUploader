package handlers

import (
	"imageUploader/internal/config"
	"imageUploader/internal/repository/dbrepo"
)

type Handlers struct {
	User
}

func NewHandlers(app *config.AppConfig, repo dbrepo.UserRepository) *Handlers {
	return &Handlers{
		User: NewUserHandlers(app, repo),
	}

}
