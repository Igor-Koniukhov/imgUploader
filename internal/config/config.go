package config

import (
	"html/template"
)

type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	Email         string
	Name          string
	Birthdate     string
	ErrorMessage  string
}
