package config

import (
	"github.com/alexedwards/scs/v2"
	"github.com/burakkarasel/bookings/internal/models"
	"html/template"
	"log"
)

//AppConfig holds the application config
type AppConfig struct {
	TemplateCache map[string]*template.Template
	UseCache      bool
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
	MailChan      chan models.MailData
}
