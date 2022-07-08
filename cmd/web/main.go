package main

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/handlers"
	"github.com/burakkarasel/bookings/internal/models"
	"github.com/burakkarasel/bookings/internal/utils"
	"log"
	"net/http"
	"time"
)

const port = ":3000"

var app config.AppConfig

var session *scs.SessionManager

func main() {
	app.InProduction = false

	// We used gob here to keep non-primitive types in our session
	gob.Register(models.Reservation{})

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	// we set secure to false because we are not using https right now our sessions are not gonna be encrypted
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := utils.CreateTemplateCache()

	if err != nil {
		log.Fatalln("cannot create template cache")
	}

	app.TemplateCache = tc
	// it gives us to access to developer mode so we can make changes on templates
	// but as soon as we are done we should assign Usecache to false otherwise it will start reading from disc again
	app.UseCache = true

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	utils.NewTemplates(&app)

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
