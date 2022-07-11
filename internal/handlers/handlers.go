package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/forms"
	"github.com/burakkarasel/bookings/internal/models"
	"github.com/burakkarasel/bookings/internal/utils"
	"log"
	"net/http"
)

// Repository holds our app's configurations
type Repository struct {
	App *config.AppConfig
}

var Repo *Repository

// jsonResponse lets us to marshal & unmarshal json data that comes with the request
type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// NewRepo lets us to create a new repository that keeps app's configurations in it
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers holds our handlers in Repository struct
func NewHandlers(r *Repository) {
	Repo = r
}

// Home renders and displays the home page
func (repo *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	repo.App.Session.Put(r.Context(), "remote_ip", remoteIP)
	utils.RenderTemplate(w, r, "home.page.gohtml", &models.TemplateData{})
}

// About renders and displays the about page
func (repo *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again"

	remoteIP := repo.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	utils.RenderTemplate(w, r, "about.page.gohtml", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Generals renders the room's page
func (repo *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, r, "generals.page.gohtml", &models.TemplateData{})
}

// Majors renders the room's page
func (repo *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, r, "majors.page.gohtml", &models.TemplateData{})
}

// Availability renders the search page for availability
func (repo *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, r, "search-availability.page.gohtml", &models.TemplateData{})
}

// Contact renders the contact page
func (repo *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, r, "contact.page.gohtml", &models.TemplateData{})
}

// MakeReservation renders our reservation page
func (repo *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation

	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	utils.RenderTemplate(w, r, "make-reservation.page.gohtml", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostAvailability sends our request
func (repo *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	_, _ = w.Write([]byte(fmt.Sprintf("You made your reservation from %s to %s!\n", start, end)))
}

// AvailabilityJSON handles request for availability and sends back JSON response
func (repo *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	tempRes := jsonResponse{
		OK:      true,
		Message: "Available!",
	}

	res, err := json.MarshalIndent(tempRes, "", "  ")

	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res)
}

// PostMakeReservation handles the posting of a reservation form
func (repo *Repository) PostMakeReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Println(err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	// with this we create a new form struct and pass url values we receive from request with r.PostForm
	form := forms.New(r.PostForm)

	// first checks if required are is filled or not then checks for length
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		utils.RenderTemplate(w, r, "make-reservation.page.gohtml", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// we put the value we receive from form into our session so we can display it when we redirect to reservation
	// summary route
	repo.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

}

// ReservationSummary renders summary of reservation according to user's inputs
func (repo *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	// we need to pass our data type that we want to pass the values into
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		repo.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// removed the reservation data from session after storing it in a variable
	repo.App.Session.Remove(r.Context(), "reservation")

	// we create data everytime and we pass whatever we want into it
	data := make(map[string]interface{})
	data["reservation"] = reservation

	utils.RenderTemplate(w, r, "reservation-summary.page.gohtml", &models.TemplateData{
		Data: data,
	})
}
