package handlers

import (
	"fmt"
	"github.com/burakkarasel/bookings/pkg/config"
	"github.com/burakkarasel/bookings/pkg/models"
	"github.com/burakkarasel/bookings/pkg/utils"
	"net/http"
)

type Repository struct {
	App *config.AppConfig
}

var Repo *Repository

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
	utils.RenderTemplate(w, r, "make-reservation.page.gohtml", &models.TemplateData{})
}

// PostAvailability sends our request
func (repo *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	w.Write([]byte(fmt.Sprintf("You made your reservation from %s to %s!\n", start, end)))
}
