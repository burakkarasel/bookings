package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/burakkarasel/bookings/internal/models"
)

// theTests holds our test cases
var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{
		name:               "home",
		url:                "/",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "about",
		url:                "/about",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "generals",
		url:                "/generals-quarters",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "majors",
		url:                "/majors-suite",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "search availability",
		url:                "/search-availability",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "contact",
		url:                "/contact",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "show-login",
		url:                "/user/login",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "show-logout",
		url:                "/user/logout",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "admin-dashboard",
		url:                "/admin/dashboard",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "none existent page",
		url:                "/don/don/don",
		method:             "GET",
		expectedStatusCode: http.StatusNotFound,
	},
	{
		name:               "new reservations",
		url:                "/admin/reservations-new",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "all reservations",
		url:                "/admin/reservations-all",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "show reservations details",
		url:                "/admin/reservations/new/3/show",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "show reservation calendar",
		url:                "/admin/reservations-calendar",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "show reservation calendar with params",
		url:                "/admin/reservations-calendar?y=2022&m=12",
		method:             "GET",
		expectedStatusCode: http.StatusOK,
	},
}

// TestGetHandlers is our test func for handlers, it tests only our render handlers
func TestGetHandlers(t *testing.T) {
	routes := getRoutes()
	// we created a test server to run our tests
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, tt := range theTests {
		resp, err := ts.Client().Get(ts.URL + tt.url)

		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != tt.expectedStatusCode {
			t.Errorf("for %s expected %d, got %d", tt.name, tt.expectedStatusCode, resp.StatusCode)
		}
	}
}

// TestRepository_MakeReservation tests our MakeReservation handler by creating a new session
func TestRepository_MakeReservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}
	// we created a new request then used our getCtx func to add session to our request
	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// NewRecorder creates a fake request response life cycle
	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	// here we made our MakeReservation func a HandlerFunc
	handler := http.HandlerFunc(Repo.MakeReservation)
	// and here we call it without it's route just using our fake req-res cycle and request with context we created
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// test case where reservation not in session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test with non-existent room, so I will get no availability
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	// updated roomID
	reservation.RoomID = 500
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

// getCtx adds session to our request
func getCtx(r *http.Request) context.Context {
	ctx, err := session.Load(r.Context(), r.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	return ctx
}

// TestRepository_PostMakeReservation test PostMakeReservation handler by creating a new session putting the information
// in session, and putting the remaining information from the form and makes the request
func TestRepository_PostMakeReservation(t *testing.T) {
	// filled necessary part of the reservation and put it in the session
	sd, _ := time.Parse("2006-01-02", "2050-01-01")
	ed, _ := time.Parse("2006-01-02", "2050-01-02")

	reservation := models.Reservation{
		RoomID:    1,
		StartDate: sd,
		EndDate:   ed,
	}

	// put remaining missing parts to body of my request

	postedData := url.Values{}
	postedData.Add("first_name", "John")
	postedData.Add("last_name", "Smith")
	postedData.Add("email", "john@here.com")
	postedData.Add("phone", "555-555-5555")

	req, _ := http.NewRequest("POST", "/post-make-reservation", strings.NewReader(postedData.Encode()))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	session.Put(ctx, "reservation", reservation)

	// this header says the request you are receiving is a form post
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostMakeReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostMakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// testing if body has no info init
	req, _ = http.NewRequest("POST", "/post-make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	session.Put(ctx, "reservation", reservation)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostMakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// testing session not available
	req, _ = http.NewRequest("POST", "/post-make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostmakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
	// failing form validation with invalid data

	postedData.Set("first_name", "John")
	postedData.Set("last_name", "Smith")
	postedData.Set("email", "john@here.com")
	postedData.Set("phone", "555-555-5555")

	req, _ = http.NewRequest("POST", "/post-make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	session.Put(ctx, "reservation", reservation)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostmakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// failing insert data
	postedData.Set("first_name", "John")

	req, _ = http.NewRequest("POST", "/post-make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reservation.RoomID = 2
	session.Put(ctx, "reservation", reservation)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostMakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// failing insert room restriction

	req, _ = http.NewRequest("POST", "post-make-reservation", strings.NewReader(postedData.Encode()))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reservation.RoomID = 0
	session.Put(ctx, "reservation", reservation)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostMakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

// TestRepository_AvailabilityJSON tests AvailabilityJSON handler
func TestRepository_AvailabilityJSON(t *testing.T) {

	tests := []struct {
		TestName     string
		StartDate    string
		EndDate      string
		RoomID       string
		ExpectedJson jsonResponse
	}{
		{
			TestName:  "Success",
			StartDate: "start_date=2050-01-01",
			EndDate:   "end_date=2050-01-02",
			RoomID:    "room_id=1",
			ExpectedJson: jsonResponse{
				OK:        true,
				StartDate: "2050-01-01",
				EndDate:   "2050-01-02",
				RoomID:    "1",
			},
		}, {
			TestName:  "Invalid Start Date",
			StartDate: "start_date=invalid",
			EndDate:   "end_date=2050-01-02",
			RoomID:    "room_id=1",
			ExpectedJson: jsonResponse{
				OK:      false,
				Message: "error during parsing start date",
			},
		},
		{
			TestName:  "Invalid End Date",
			StartDate: "start_date=2050-01-01",
			EndDate:   "end_date=invalid",
			RoomID:    "room_id=1",
			ExpectedJson: jsonResponse{
				OK:      false,
				Message: "error during parsing end date",
			},
		},
		{
			TestName:  "Invalid Room ID",
			StartDate: "start_date=2050-01-01",
			EndDate:   "end_date=2050-01-02",
			RoomID:    "room_id=invalid",
			ExpectedJson: jsonResponse{
				OK:      false,
				Message: "error during parsing room_id",
			},
		},
		{
			TestName:  "DB fail",
			StartDate: "start_date=2050-01-01",
			EndDate:   "end_date=2050-01-02",
			RoomID:    "room_id=17",
			ExpectedJson: jsonResponse{
				OK:      false,
				Message: "error cannot reach database",
			},
		},
	}

	for _, test := range tests {
		reqBody := fmt.Sprintf("%s&%s&%s", test.StartDate, test.EndDate, test.RoomID)

		req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(rr, req)

		var j jsonResponse
		err := json.Unmarshal([]byte(rr.Body.Bytes()), &j)

		if err != nil {
			t.Error("failed to parse json")
		}

		if j != test.ExpectedJson {
			t.Errorf("AvailabilityJSON handler returned wrong response: got %v, wanted %v", j, test.ExpectedJson)
		}
	}
}

// TestRepository_PostAvailability func tests PostAvailability handler
func TestRepository_PostAvailability(t *testing.T) {
	tests := []struct {
		TestName           string
		StartDate          string
		EndDate            string
		ExpectedStatusCode int
	}{
		{
			TestName:           "Success",
			StartDate:          "start_date=2050-01-01",
			EndDate:            "end_date=2050-01-02",
			ExpectedStatusCode: http.StatusOK,
		},
		{
			TestName:           "Invalid start date",
			StartDate:          "start_date=invalid",
			EndDate:            "end_date=2050-01-02",
			ExpectedStatusCode: http.StatusSeeOther,
		},
		{
			TestName:           "Invalid end date",
			StartDate:          "start_date=2050-01-01",
			EndDate:            "end_date=invalid",
			ExpectedStatusCode: http.StatusSeeOther,
		},
		{
			TestName:           "DB error",
			StartDate:          "start_date=2023-02-19",
			EndDate:            "end_date=2023-02-21",
			ExpectedStatusCode: http.StatusSeeOther,
		},
	}

	for _, test := range tests {
		reqBody := fmt.Sprintf("%s&%s", test.StartDate, test.EndDate)

		req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.PostAvailability)
		handler.ServeHTTP(rr, req)

		if rr.Code != test.ExpectedStatusCode {
			t.Errorf("POST availability handler returned wrong status code: got %d, wanted %d", rr.Code, test.ExpectedStatusCode)
		}
	}
}

// TestRepository_ReservationSummary tests ReservationSummary handler
func TestRepository_ReservationSummary(t *testing.T) {
	// fail cannot pull reservation from session
	req, _ := http.NewRequest("GET", "/reservation-summary", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.ReservationSummary)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ReservationSummary handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// success after pulling reservation out from session
	reservation := models.Reservation{}

	req, _ = http.NewRequest("GET", "/reservation-summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	session.Put(ctx, "reservation", reservation)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ReservationSummary handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusOK)
	}
}

// TestRepository_ChooseRoom tests ChooseRoom handler
func TestRepository_ChooseRoom(t *testing.T) {
	// without session
	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.RequestURI = "/choose-room/1"

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// invalid room id
	req, _ = http.NewRequest("GET", "/choose-room/iii", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	reservation := models.Reservation{}
	session.Put(ctx, "reservation", reservation)

	req.RequestURI = "/choose/room/iii"

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// successful
	req, _ = http.NewRequest("GET", "/choose-room/1", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	session.Put(ctx, "reservation", reservation)

	req.RequestURI = "/choose-room/1"

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}
}

// TestRepository_BookRoom tests BookRoom handler
func TestRepository_BookRoom(t *testing.T) {
	tests := []struct {
		TestName           string
		RoomID             string
		StartDate          string
		EndDate            string
		ExpectedStatusCode int
	}{
		{
			TestName:           "Success",
			RoomID:             "id=1",
			StartDate:          "s=2050-01-01",
			EndDate:            "e=2050-01-02",
			ExpectedStatusCode: http.StatusSeeOther,
		}, {
			TestName:           "Invalid room id",
			RoomID:             "id=iii",
			StartDate:          "s=2050-01-01",
			EndDate:            "e=2050-01-02",
			ExpectedStatusCode: http.StatusSeeOther,
		},
		{
			TestName:           "Invalid  start date",
			RoomID:             "id=1",
			StartDate:          "s=invalid",
			EndDate:            "e=2050-01-02",
			ExpectedStatusCode: http.StatusSeeOther,
		},
		{
			TestName:           "Invalid  end date",
			RoomID:             "id=1",
			StartDate:          "s=2050-01-01",
			EndDate:            "e=invalid",
			ExpectedStatusCode: http.StatusSeeOther,
		},
		{
			TestName:           "DB error",
			RoomID:             "id=5",
			StartDate:          "s=2050-01-01",
			EndDate:            "e=2050-01-02",
			ExpectedStatusCode: http.StatusSeeOther,
		},
	}

	for _, test := range tests {
		requestURL := fmt.Sprintf("/book-room?%s&%s&%s", test.RoomID, test.StartDate, test.EndDate)
		req, _ := http.NewRequest("GET", requestURL, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		session.Put(ctx, "reservation", models.Reservation{})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.BookRoom)
		handler.ServeHTTP(rr, req)

		if rr.Code != test.ExpectedStatusCode {
			t.Errorf("BookRoom handler returned wrong status code: got %d, wanted %d", rr.Code, test.ExpectedStatusCode)
		}
	}
}

// TestRepository_PostShowLogin tests PostShowLogin handler
func TestRepository_PostShowLogin(t *testing.T) {
	var loginTests = []struct {
		name                string
		email               string
		expecetedStatusCode int
		expectedHTML        string
		expectedLocation    string
	}{
		{
			name:                "valid credentials",
			email:               "me@here.ca",
			expecetedStatusCode: http.StatusSeeOther,
			expectedLocation:    "/",
		},
		{
			name:                "invalid credentials",
			email:               "jack@nimble.com",
			expecetedStatusCode: http.StatusSeeOther,
			expectedLocation:    "/user/login",
		},
		{
			name:                "invalid data",
			email:               "c",
			expecetedStatusCode: http.StatusOK,
			expectedHTML:        `action="/user/login"`,
		},
	}

	for _, tt := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", tt.email)
		postedData.Add("password", "password")

		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expecetedStatusCode {
			t.Errorf("%s failed : got %d, wanted %d", tt.name, rr.Code, tt.expecetedStatusCode)
		}

		if tt.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if tt.expectedLocation != actualLocation.String() {
				t.Errorf("%s failed: redirected to %s instead of %s", tt.name, actualLocation.String(), tt.expectedLocation)
			}
		}

		if tt.expectedHTML != "" {
			html := rr.Body.String()
			if !strings.Contains(html, tt.expectedHTML) {
				t.Errorf("%s failed: got body %s, expected %s", tt.name, html, tt.expectedHTML)
			}
		}
	}
}

// TestRepository_AdminPostReservationsCalendar tests AdminPostReservationsCalendar handler
func TestRepository_AdminPostReservationsCalendar(t *testing.T) {
	var tests = []struct {
		name                 string
		postedData           url.Values
		expectedResponseCode int
		expectedLocation     string
		expectedHTML         string
		blocks               int
		reservations         int
	}{
		{
			name: "cal",
			postedData: url.Values{
				"year":  {time.Now().Format("2006")},
				"month": {time.Now().Format("01")},
				fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
			},
			expectedResponseCode: http.StatusSeeOther,
		},
		{
			name:                 "cal-blocks",
			postedData:           url.Values{},
			expectedResponseCode: http.StatusSeeOther,
			blocks:               1,
		},
		{
			name:                 "cal-res",
			postedData:           url.Values{},
			expectedResponseCode: http.StatusSeeOther,
			reservations:         1,
		},
	}

	for _, tt := range tests {
		var req *http.Request
		if tt.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", strings.NewReader(tt.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)

		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; d.Before(lastOfMonth); d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-2")] = 0
			bm[d.Format("2006-01-2")] = 0
		}

		if tt.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = tt.blocks
		}

		if tt.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = tt.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedResponseCode {
			t.Errorf("failed %s: got code %d, wanted %d", tt.name, rr.Code, tt.expectedResponseCode)
		}
	}
}

// TestRepository_AdminPostShowReservationDetail tests AdminPostShowReservationDetail handler
func TestRepository_AdminPostShowReservationDetail(t *testing.T) {
	var tests = []struct {
		name               string
		url                string
		postedData         url.Values
		expectedLocation   string
		expectedStatuscode int
	}{
		{
			name: "valid calendar",
			url:  "/admin/reservations/cal/3/show",
			postedData: url.Values{
				"first_name": {"John"},
				"last_name":  {"Smith"},
				"email":      {"john@smith.com"},
				"phone":      {"555-555-5555"},
				"year":       {"2022"},
				"month":      {"07"},
			},
			expectedLocation:   "/admin/reservations-calendar?y=2022&m=07",
			expectedStatuscode: http.StatusSeeOther,
		},
		{
			name: "valid all",
			url:  "/admin/reservations/all/3/show",
			postedData: url.Values{
				"first_name": {"John"},
				"last_name":  {"Smith"},
				"email":      {"john@smith.com"},
				"phone":      {"555-555-5555"},
			},
			expectedLocation:   "/admin/reservations-all",
			expectedStatuscode: http.StatusSeeOther,
		},
		{
			name: "valid new",
			url:  "/admin/reservations/new/4/show",
			postedData: url.Values{
				"first_name": {"John"},
				"last_name":  {"Smith"},
				"email":      {"john@smith.com"},
				"phone":      {"555-555-5555"},
			},
			expectedLocation:   "/admin/reservations-new",
			expectedStatuscode: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		var req *http.Request
		if tt.postedData != nil {
			req, _ = http.NewRequest("POST", tt.url, strings.NewReader(tt.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", tt.url, nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = tt.url
		req.Form = tt.postedData

		req.Header.Set("Content-Type", "x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminPostShowReservationDetail)
		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatuscode {
			t.Errorf("for %s: got status code %d, wanted %d", tt.name, rr.Code, tt.expectedStatuscode)
		}

		if tt.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if tt.expectedLocation != actualLocation.String() {
				t.Errorf("for %s: got location %s, wanted %s", tt.name, actualLocation.String(), tt.expectedLocation)
			}
		}
	}
}

// TestRepository_AdminProcessedReservation tests AdminProcessedReservation handler
func TestRepository_AdminProcessedReservation(t *testing.T) {
	var tests = []struct {
		name               string
		url                string
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name:               "valid cal",
			url:                "/admin/process-reservations/cal/1/do?y=2022&m=07",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/admin/reservations-calendar?y=2022&m=07",
		},
		{
			name:               "valid all",
			url:                "/admin/process-reservations/all/1/do",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/admin/reservations-all",
		},
		{
			name:               "valid new",
			url:                "/admin/process-reservations/new/1/do",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/admin/reservations-new",
		},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest("GET", tt.url, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = tt.url

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminProcessedReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("for %s: got %d, wanted %d", tt.name, rr.Code, tt.expectedStatusCode)
		}

		if tt.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if actualLocation.String() != tt.expectedLocation {
				t.Errorf("for %s: got location %s, wanted %s", tt.name, actualLocation.String(), tt.expectedLocation)
			}
		}
	}
}

// TestRepository_AdminDeleteReservation tests AdminDeleteReservation handler
func TestRepository_AdminDeleteResevation(t *testing.T) {
	var tests = []struct {
		name               string
		url                string
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name:               "del cal",
			url:                "/admin/delete-reservations/cal/1/do?y=2022&m=07",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/admin/reservations-calendar?y=2022&m=07",
		},
		{
			name:               "del all",
			url:                "/admin/delete-reservations/all/1/do",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/admin/reservations-all",
		},
		{
			name:               "del new",
			url:                "/admin/delete-reservations/new/1/do",
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/admin/reservations-new",
		},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest("GET", tt.url, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = tt.url

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatusCode {
			t.Errorf("for %s : got status code %d, wanted %d", tt.name, rr.Code, tt.expectedStatusCode)
		}

		if tt.expectedLocation != "" {
			actualLocation, _ := rr.Result().Location()
			if actualLocation.String() != tt.expectedLocation {
				t.Errorf("for %s: got location %s, wanted %s", tt.name, actualLocation.String(), tt.expectedLocation)
			}
		}
	}
}
