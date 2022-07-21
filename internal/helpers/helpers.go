package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/burakkarasel/bookings/internal/config"
)

var app *config.AppConfig

// NewHelpers sets up app config for helpers func
func NewHelpers(a *config.AppConfig) {
	app = a
}

// ClientError responses with a status and status's text, and logs the information to terminal
func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

// ServerError logs error, and it's trace to terminal and gives a http error to the user
func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// IsAuthenticated returns if a user logged in or not by checking the session for user_id variable
func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}
