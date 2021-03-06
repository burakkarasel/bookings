package driver

import (
	"database/sql"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"time"
)

// DB holds the DB connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

// ConnectSQL creates DB pool for postgres
func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDataBase(dsn)
	if err != nil {
		panic(err)
	}

	// these will prevent loss of control on database
	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetConnMaxIdleTime(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbLifetime)

	dbConn.SQL = d

	err = testDB(d)

	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

// testDB tries to ping DB returns a potential error
func testDB(d *sql.DB) error {
	err := d.Ping()

	if err != nil {
		return err
	}

	return nil
}

// NewDataBase creates a new DB for the application
func NewDataBase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, err
	}

	if err = testDB(db); err != nil {
		return nil, err
	}

	return db, nil
}
