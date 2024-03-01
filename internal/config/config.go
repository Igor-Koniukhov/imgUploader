package config

import (
	"html/template"
	"imageUploader/models"
)

type AppConfig struct {
	UseCache          bool
	UserInfoFromCache *models.User
	TemplateCache     map[string]*template.Template
	UserId            string
	Email             string
	Name              string
	Birthdate         string
	ErrorMessage      string
}
