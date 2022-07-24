package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/driver"
	"github.com/burakkarasel/bookings/internal/dsn"
	"github.com/burakkarasel/bookings/internal/handlers"
	"github.com/burakkarasel/bookings/internal/helpers"
	"github.com/burakkarasel/bookings/internal/models"
	"github.com/burakkarasel/bookings/internal/utils"
)

const port = ":3000"

var app config.AppConfig

var session *scs.SessionManager

var infoLog *log.Logger
var errorLog *log.Logger

func main() {

	db, err := run()

	if err != nil {
		log.Fatal(err)
	}

	// we closed our DB here because when run function stops running our database was going to close itself
	defer db.SQL.Close()

	// here I close my mail channel because closing it in run func doesn't make sense
	defer close(app.MailChan)
	log.Println("Starting mail listener!")
	listenForMail()

	fmt.Println("starting at port", port)

	srv := &http.Server{
		Addr:    port,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}

// run func includes most of the code we have in main func, and we check anything that might return an error
// we build run func because we don't want to test func main
func run() (*driver.DB, error) {
	app.InProduction = true

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// We used gob here to keep non-primitive types in our session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	// we set secure to false because we are not using https right now our sessions are not going to be encrypted
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to DB
	log.Println("Connecting to DB...")
	db, err := driver.ConnectSQL(dsn.Dsn)

	if err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}

	log.Println("Connected to DB!")

	tc, err := utils.CreateTemplateCache()

	if err != nil {
		return nil, err
	}
	app.TemplateCache = tc
	// it gives us to access to developer mode, so we can make changes on templates
	// but as soon as we are done we should assign UseCache to true otherwise it will start reading from disc again
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	utils.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
