package handlers

import (
	"encoding/json"
	"errors"
	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/driver"
	"github.com/burakkarasel/bookings/internal/forms"
	"github.com/burakkarasel/bookings/internal/helpers"
	"github.com/burakkarasel/bookings/internal/models"
	"github.com/burakkarasel/bookings/internal/repository"
	"github.com/burakkarasel/bookings/internal/repository/dbrepo"
	"github.com/burakkarasel/bookings/internal/utils"
	"github.com/go-chi/chi"
	"net/http"
	"strconv"
	"time"
)

// Repository holds our app's configurations
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

var Repo *Repository

// jsonResponse lets us marshal & unmarshal json data that comes with the request
type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// NewRepo lets us create a new repository that keeps app's configurations in it
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers holds our handlers in Repository struct
func NewHandlers(r *Repository) {
	Repo = r
}

// Home renders and displays the home page
func (repo *Repository) Home(w http.ResponseWriter, r *http.Request) {
	err := utils.Template(w, r, "home.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// About renders and displays the about page
func (repo *Repository) About(w http.ResponseWriter, r *http.Request) {
	err := utils.Template(w, r, "about.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// Generals renders the room's page
func (repo *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	err := utils.Template(w, r, "generals.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// Majors renders the room's page
func (repo *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	err := utils.Template(w, r, "majors.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// Availability renders the search page for availability
func (repo *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	err := utils.Template(w, r, "search-availability.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// Contact renders the contact page
func (repo *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	err := utils.Template(w, r, "contact.page.gohtml", &models.TemplateData{})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// MakeReservation renders our reservation page
func (repo *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	res, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w, errors.New("cannot get reservation from session"))
		return
	}

	roomData, err := repo.DB.GetRoomById(res.RoomID)

	if err != nil {
		helpers.ServerError(w, err)
	}

	res.Room.RoomName = roomData.RoomName
	// after updating the session we put back the updated data
	repo.App.Session.Put(r.Context(), "reservation", res)

	layout := "2006-01-02"
	sd := res.StartDate.Format(layout)
	ed := res.EndDate.Format(layout)

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	err = utils.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// PostAvailability sends our request
func (repo *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start_date")
	end := r.Form.Get("end_date")

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, start)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, end)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	availRooms, err := repo.DB.SearchAvailabilityForAllRooms(startDate, endDate)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(availRooms) == 0 {
		repo.App.Session.Put(r.Context(), "error", "No Availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = availRooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	repo.App.Session.Put(r.Context(), "reservation", res)

	err = utils.Template(w, r, "choose-room.page.gohtml", &models.TemplateData{
		Data: data,
	})
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// AvailabilityJSON handles request for availability and sends back JSON response
func (repo *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, ed)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	available, err := repo.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	tempRes := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	res, err := json.MarshalIndent(tempRes, "", "  ")

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res)
}

// PostMakeReservation handles the posting of a reservation form
func (repo *Repository) PostMakeReservation(w http.ResponseWriter, r *http.Request) {
	// after putting the updated data at make reservation we put out the last version of reservation
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("cannot get reservation from session"))
		return
	}
	err := r.ParseForm()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	// with this we create a new form struct and pass url values we receive from request with r.PostForm
	form := forms.New(r.PostForm)

	// first checks if required are is filled or not then checks for length
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		err := utils.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
			Form: form,
			Data: data,
		})

		if err != nil {
			helpers.ServerError(w, err)
		}
		return
	}

	// after validating our form we insert reservation data to DB
	newReservationID, err := repo.DB.InsertReservation(reservation)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// after inserting a new reservation we need to add restriction to these dates too
	// created a new restriction
	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	// then inserted it into room restriction table
	err = repo.DB.InsertRoomRestriction(restriction)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// we put the value we receive from form into our session as last version of the reservation,
	//so we can display it when we redirect to reservation summary route
	repo.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

}

// ReservationSummary renders summary of reservation according to user's inputs
func (repo *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	// we need to pass our data type that we want to pass the values into
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		repo.App.ErrorLog.Println("Can't get error from session")
		repo.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// removed the reservation data from session after storing it in a variable
	repo.App.Session.Remove(r.Context(), "reservation")

	// we create data everytime, and we pass whatever we want into it
	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	err := utils.Template(w, r, "reservation-summary.page.gohtml", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})

	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// ChooseRoom displays available rooms for the given date
func (repo *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		helpers.ServerError(w, err)
	}

	res, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w, err)
	}

	res.RoomID = roomID

	repo.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom handles requests from each room's page, takes url parameters, and builds a session variable to pass
// /make-reservation route
func (repo *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// id, s, e
	roomID, err := strconv.Atoi(r.URL.Query().Get("id"))

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	startDate := r.URL.Query().Get("s")
	endDate := r.URL.Query().Get("e")

	var res models.Reservation

	res.RoomID = roomID

	sd, err := time.Parse("2006-01-02", startDate)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	ed, err := time.Parse("2006-01-02", endDate)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.StartDate = sd
	res.EndDate = ed

	room, err := repo.DB.GetRoomById(roomID)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.Room.RoomName = room.RoomName

	repo.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
