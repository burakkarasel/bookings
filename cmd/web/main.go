package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/driver"
	"github.com/burakkarasel/bookings/internal/handlers"
	"github.com/burakkarasel/bookings/internal/helpers"
	"github.com/burakkarasel/bookings/internal/models"
	"github.com/burakkarasel/bookings/internal/utils"
)

var port = ":8080"

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

	// read flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", false, "Stop using template cache")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database username")
	dbPw := flag.String("dbpw", "", "Database password")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()

	if *dbName == "" || *dbUser == "" {
		fmt.Println("missing required flags")
		os.Exit(1)
	}

	app.InProduction = *inProduction

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	// we set secure to false because we are not using https right now our sessions are not going to be encrypted
	session.Cookie.Secure = *inProduction

	app.Session = session

	// connect to DB
	log.Println("Connecting to DB...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPw, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)

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
	app.UseCache = *useCache

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	utils.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
