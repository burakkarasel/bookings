package utils

import (
	"encoding/gob"
	"github.com/alexedwards/scs/v2"
	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/models"
	"net/http"
	"os"
	"testing"
	"time"
)

var session *scs.SessionManager
var testApp config.AppConfig

// myWriter is to satisfy RenderTemplate func
type myWriter struct{}

// TestMain builds environment to run our test for render.go
func TestMain(m *testing.M) {

	testApp.InProduction = false

	// We used gob here to keep non-primitive types in our session
	gob.Register(models.Reservation{})

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	// we set secure to false because we are not using https right now our sessions are not gonna be encrypted
	session.Cookie.Secure = false

	testApp.Session = session

	app = &testApp

	os.Exit(m.Run())
}

// Header is used to implement Writer interface
func (mw *myWriter) Header() http.Header {
	var h http.Header
	return h
}

// Write is used to implement Writer interface
func (mw *myWriter) Write(bs []byte) (int, error) {
	length := len(bs)
	return length, nil
}

// WriteHeader is used to implement Writer interface
func (mw *myWriter) WriteHeader(i int) {

}
