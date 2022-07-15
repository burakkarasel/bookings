package dbrepo

import (
	"database/sql"
	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/repository"
)

// postgresDBRepo holds our DB and memory address of our app to connect db in main.go
type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewPostgresRepo takes app and connection to db as arguments and  returns postgresDbRepo that implements
// DatabaseRepo interface
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}
