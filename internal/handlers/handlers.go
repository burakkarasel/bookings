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

type Repository struct {
	App *config.AppConfig
}

var Repo *Repository

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}

func (repo *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	repo.App.Session.Put(r.Context(), "remote_ip", remoteIP)
	utils.RenderTemplate(w, r, "home.page.gohtml", &models.TemplateData{})
}

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
	w.Write([]byte(fmt.Sprintf("You made your reservation from %s to %s!\n", start, end)))
}

func (repo *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	tempRes := jsonResponse{
		OK:      true,
		Message: "Avaliable!",
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

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3, r)
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

	repo.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

}

func (repo *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		log.Println("cannot get item from session")
		repo.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	repo.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	utils.RenderTemplate(w, r, "reservation-summary.page.gohtml", &models.TemplateData{
		Data: data,
	})
}
