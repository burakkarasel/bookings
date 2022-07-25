package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/models"
	"github.com/burakkarasel/bookings/internal/utils"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/justinas/nosurf"
)

var app config.AppConfig
var session *scs.SessionManager

// pathToTemplates we needed to make another path for templates because we will run test files in handlers directory
var pathToTemplates = "./../../templates"
var functions = template.FuncMap{
	"humanDate":  utils.HumanDate,
	"formatDate": utils.FormatDate,
	"iterate":    utils.Iterate,
	"add":        utils.Add,
}

func TestMain(m *testing.M) {
	// main.go
	app.InProduction = false

	// We used gob here to keep non-primitive types in our session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	// we set secure to false because we are not using https right now our sessions are not gonna be encrypted
	session.Cookie.Secure = app.InProduction

	app.Session = session

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)

	listenForMail()

	tc, err := CreateTestTemplateCache()

	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = tc
	// it gives us to access to developer mode, so we can make changes on templates
	// but as soon as we are done we should assign Use-cache to false otherwise it will start reading from disc again
	app.UseCache = true

	repo := NewTestRepo(&app)
	NewHandlers(repo)

	utils.NewRenderer(&app)

	os.Exit(m.Run())
}

// getRoutes function is a preparation for our handler's test file because we will need routes
func getRoutes() http.Handler {
	// routes.go
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	// NoSurf adds CSRF protection to POST requests
	//mux.Use(NoSurf)
	// When we are testing our POST handlers with don't need to use CSRF protection that's we comment it out
	// Because we test it out in middleware tests
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)
	mux.Get("/contact", Repo.Contact)

	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)

	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostMakeReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	mux.Get("/choose-room/{id}", Repo.ChooseRoom)

	mux.Get("/book-room", Repo.BookRoom)

	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/user/logout", Repo.Logout)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Get("/admin/dashboard", Repo.AdminDashboard)

	mux.Get("/admin/reservations-new", Repo.AdminNewReservations)
	mux.Get("/admin/reservations-all", Repo.AdminAllReservations)
	mux.Get("/admin/reservations-calendar", Repo.AdminReservationsCalendar)
	mux.Post("/admin/reservations-calendar", Repo.AdminPostReservationsCalendar)

	mux.Get("/admin/process-reservations/{src}/{id}/do", Repo.AdminProcessedReservation)
	mux.Get("/admin/delete-reservations/{src}/{id}/do", Repo.AdminDeleteReservation)

	mux.Get("/admin/reservations/{src}/{id}/show", Repo.AdminShowReservationDetail)
	mux.Post("/admin/reservations/{src}/{id}", Repo.AdminPostShowReservationDetail)

	return mux
}

// middleware.go
// NoSurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves session data
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// CreateTestTemplateCache creates template cache as a map for this package
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.gohtml", pathToTemplates))

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))

		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}
	return myCache, nil
}

// listenForMail lets us to implement mail functionality to our test file
func listenForMail() {
	go func() {
		for {
			_ = <-app.MailChan
		}
	}()
}
