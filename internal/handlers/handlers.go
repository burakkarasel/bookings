package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/burakkarasel/bookings/internal/config"
	"github.com/burakkarasel/bookings/internal/driver"
	"github.com/burakkarasel/bookings/internal/forms"
	"github.com/burakkarasel/bookings/internal/helpers"
	"github.com/burakkarasel/bookings/internal/models"
	"github.com/burakkarasel/bookings/internal/repository"
	"github.com/burakkarasel/bookings/internal/repository/dbrepo"
	"github.com/burakkarasel/bookings/internal/utils"
	"github.com/go-chi/chi"
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

// NewTestRepo lets us create a new repository that keeps app's configurations in it just for unit testing
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
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
		repo.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomData, err := repo.DB.GetRoomById(res.RoomID)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
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
	r.ParseForm()
	start := r.Form.Get("start_date")
	end := r.Form.Get("end_date")

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, start)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "start date is invalid")
		http.Redirect(w, r, "/search-availability", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, end)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "end date is invalid")
		http.Redirect(w, r, "/search-availability", http.StatusTemporaryRedirect)
		return
	}

	availRooms, err := repo.DB.SearchAvailabilityForAllRooms(startDate, endDate)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "DB error")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
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
	err := r.ParseForm()

	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "error during parsing form",
		}

		res, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)

	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "error during parsing start date",
		}

		res, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)

		repo.App.Session.Put(r.Context(), "error", "invalid start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)

	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "error during parsing end date",
		}

		res, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)

		repo.App.Session.Put(r.Context(), "error", "invalid end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))

	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "error during parsing room_id",
		}

		res, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)

		repo.App.Session.Put(r.Context(), "error", "invalid room id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	available, err := repo.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)

	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "error cannot reach database",
		}

		res, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)

		repo.App.Session.Put(r.Context(), "error", "invalid room id")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
	_, err = w.Write(res)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

// PostMakeReservation handles the posting of a reservation form
func (repo *Repository) PostMakeReservation(w http.ResponseWriter, r *http.Request) {
	// after putting the updated data at make reservation we put out the last version of reservation
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	err := r.ParseForm()

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "invalid form", http.StatusSeeOther)
		err := utils.Template(w, r, "make-reservation.page.gohtml", &models.TemplateData{
			Form:      form,
			Data:      data,
			StringMap: stringMap,
		})

		if err != nil {
			repo.App.Session.Put(r.Context(), "error", "cannot rerender make reservation page")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
		return
	}

	// after validating our form we insert reservation data to DB
	newReservationID, err := repo.DB.InsertReservation(reservation)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't insert reservation to database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
		repo.App.Session.Put(r.Context(), "error", "can't insert room restriction")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send notifications - first to guest
	htmlGuestMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong>
    	<br>
    	Dear %s, 
		<br>
    	This is confirmation for your reservation from %s to %s to in %s
	`, reservation.FirstName+" "+reservation.LastName, reservation.StartDate.Format("2006-01-02"),
		reservation.EndDate.Format("2006-01-02"), reservation.Room.RoomName)

	guestMSG := models.MailData{
		To:       reservation.Email,
		From:     "me@here.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlGuestMessage,
		Template: "basic.gohtml",
	}

	repo.App.MailChan <- guestMSG

	htmlOwnerMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong>
    	<br>
    	Dear Owner, 
		<br>
    	This is confirmation for reservation of your %s from %s to %s.
		You can reach the guest via this email : %s.
	`, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"),
		reservation.EndDate.Format("2006-01-02"), reservation.Email)

	ownerMessage := models.MailData{
		To:       "owner@here.com",
		From:     "me@here.com",
		Subject:  "New Reservation",
		Content:  htmlOwnerMessage,
		Template: "basic.gohtml",
	}

	repo.App.MailChan <- ownerMessage

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

// ChooseRoom is the handler of chosen room which caries the ID of the room in the URL after capturing the chosen roomID
// from the URL, put it back in the session variable reservation, and redirect to /make-reservation route
func (repo *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	//roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	//
	//if err != nil {
	//	repo.App.Session.Put(r.Context(), "error", "invalid room id")
	//	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	//	return
	//}

	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		repo.App.Session.Put(r.Context(), "error", "cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
		repo.App.Session.Put(r.Context(), "error", "cannot parse id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	startDate := r.URL.Query().Get("s")
	endDate := r.URL.Query().Get("e")

	var res models.Reservation

	res.RoomID = roomID

	sd, err := time.Parse("2006-01-02", startDate)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "invalid start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	ed, err := time.Parse("2006-01-02", endDate)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "invalid end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.StartDate = sd
	res.EndDate = ed

	room, err := repo.DB.GetRoomById(roomID)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "DB error")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	res.Room.RoomName = room.RoomName

	repo.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// ShowLogin renders the login page
func (repo *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	utils.Template(w, r, "login.page.gohtml", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging the user in
func (repo *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	// after attempt to log in we renew our csrf token immediately for safety from session fixation attacks
	_ = repo.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)

	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		utils.Template(w, r, "login.page.gohtml", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := repo.DB.Authenticate(email, password)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	repo.App.Session.Put(r.Context(), "user_id", id)

	repo.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
func (repo *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = repo.App.Session.Destroy(r.Context())
	_ = repo.App.Session.RenewToken(r.Context())

	repo.App.Session.Put(r.Context(), "warning", "Succesfully logged out!")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// AdminDashboard renders the admin-dashboard template
func (repo *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	utils.Template(w, r, "admin-dashboard.page.gohtml", &models.TemplateData{})
}

// AdminNewReservations shows all new reservations in admin dashboard
func (repo *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := repo.DB.AllNewReservations()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	utils.Template(w, r, "admin-new-reservations.page.gohtml", &models.TemplateData{
		Data: data,
	})
}

// AdminNewReservations shows all reservations in admin dashboard
func (repo *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := repo.DB.AllReservations()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	utils.Template(w, r, "admin-all-reservations.page.gohtml", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservationDetail shows the reservation's details in dashboard
func (repo *Repository) AdminShowReservationDetail(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")

	// last one in this slice is my id
	id, err := strconv.Atoi(exploded[4])

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// before the last one my source
	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	res, err := repo.DB.GetReservationById(id)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	utils.Template(w, r, "admin-reservation-detail.page.gohtml", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// AdminPostShowReservationDetail updates the reservation according to form
func (repo *Repository) AdminPostShowReservationDetail(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[4])

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	res, err := repo.DB.GetReservationById(id)

	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "cannot get reservation from DB")
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = repo.DB.UpdateReservation(res)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

	repo.App.Session.Put(r.Context(), "flash", "Successfully changed the reservation")

}

// AdminReservationsCalendar marks a reservation processed
func (repo *Repository) AdminProcessedReservation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := chi.URLParam(r, "src")

	processedVal := 1

	err = repo.DB.UpdateProcessedForReservation(id, processedVal)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

	repo.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")
}

// AdminDeleteReservation deletes a reservation from DB
func (repo *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := chi.URLParam(r, "src")

	err = repo.DB.DeleteReservation(id)

	if err != nil {
		helpers.ServerError(w, err)
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

	repo.App.Session.Put(r.Context(), "warning", "Reservation deleted successfully")
}

// AdminReservationsCalendar displays the reservation calendar
func (repo *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, err := strconv.Atoi(r.URL.Query().Get("y"))

		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		month, err := strconv.Atoi(r.URL.Query().Get("m"))

		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	previous := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	previousMonth := previous.Format("01")
	previousMonthYear := previous.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["previous_month"] = previousMonth
	stringMap["previous_month_year"] = previousMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_year"] = now.Format("2006")

	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	// here we add 1 month and reduce 1 day so we got the last day of the month
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := repo.DB.AllRooms()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	// first we range over rooms
	for _, x := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		// then we range over for each room and set their reservation and block values to 0
		for d := firstOfMonth; d.Before(lastOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0

		}

		// than we check for given room's restrictions and put them in a slice
		restrictions, err := repo.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)

		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		// then we range over the slice and update restrictions according to if the restriction is a reservation or a block
		for _, res := range restrictions {
			if res.ReservationID > 0 {
				for d := res.StartDate; d.Before(res.EndDate); d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = res.ReservationID
				}
			} else {
				blockMap[res.StartDate.Format("2006-01-2")] = res.ID
			}
		}

		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		repo.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	utils.Template(w, r, "admin-reservations-calendar.page.gohtml", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// AdminPostReservationsCalendar handles the changes that made on calendar page
func (repo *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	rooms, err := repo.DB.AllRooms()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	// this handles removed blocks
	for _, x := range rooms {
		curMap := repo.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for name, value := range curMap {
			if val, ok := curMap[name]; ok {
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						err := repo.DB.RemoveBlockForRoom(value)
						if err != nil {
							helpers.ServerError(w, err)
							return
						}
					}
				}
			}
		}
	}

	// this handles new blocks
	for name := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			// insert a new restriction

			date, err := time.Parse("2006-01-2", exploded[len(exploded)-1])

			if err != nil {
				helpers.ServerError(w, err)
				return
			}

			err = repo.DB.InsertBlockForRoom(roomID, date)

			if err != nil {
				helpers.ServerError(w, err)
				return
			}
		}
	}

	repo.App.Session.Put(r.Context(), "flash", "Chages saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}
