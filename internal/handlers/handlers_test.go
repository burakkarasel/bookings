package handlers

import (
	"context"
	"fmt"
	"github.com/burakkarasel/bookings/internal/models"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
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
}

// TestGetHandlers is our test func for handlers, it tests our handlers according to their request type
func TestGetHandlers(t *testing.T) {
	routes := getRoutes()
	// we created a test server to run our tests
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, tt := range theTests {
		if tt.method == "GET" {
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
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
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// testing session not available
	req, _ = http.NewRequest("POST", "/post-make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostmakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
	// failing form validation with invalid data

	postedData.Set("first_name", "Jo")
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

// TestRepository_AvailabilityJSON tests AvailabilityJSON handler
func TestRepository_AvailabilityJSON(t *testing.T) {

	tests := []struct {
		TestName           string
		StartDate          string
		EndDate            string
		RoomID             string
		ExpectedStatusCode int
	}{
		{
			TestName:           "Success",
			StartDate:          "start_date=2050-01-01",
			EndDate:            "end_date=2050-01-02",
			RoomID:             "room_id=1",
			ExpectedStatusCode: http.StatusOK,
		}, {
			TestName:           "Invalid Start Date",
			StartDate:          "start_date=invalid",
			EndDate:            "end_date=2050-01-02",
			RoomID:             "room_id=1",
			ExpectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			TestName:           "Invalid End Date",
			StartDate:          "start_date=2050-01-01",
			EndDate:            "end_date=invalid",
			RoomID:             "room_id=1",
			ExpectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			TestName:           "Invalid Room ID",
			StartDate:          "start_date=2050-01-01",
			EndDate:            "end_date=2050-01-02",
			RoomID:             "room_id=invalid",
			ExpectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			TestName:           "DB fail",
			StartDate:          "start_date=2050-01-01",
			EndDate:            "end_date=2050-01-02",
			RoomID:             "room_id=17",
			ExpectedStatusCode: http.StatusSeeOther,
		},
	}

	for _, test := range tests {
		reqBody := test.StartDate
		reqBody = fmt.Sprintf("%s&%s", reqBody, test.EndDate)
		reqBody = fmt.Sprintf("%s&%s", reqBody, test.RoomID)

		req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(rr, req)

		if rr.Code != test.ExpectedStatusCode {
			t.Errorf("AvailabilityJSON handler returned wrong status code: got %d, wanted %d", rr.Code, test.ExpectedStatusCode)
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
			ExpectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			TestName:           "Invalid end date",
			StartDate:          "start_date=2050-01-01",
			EndDate:            "end_date=invalid",
			ExpectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			TestName:           "DB error",
			StartDate:          "start_date=2023-02-19",
			EndDate:            "end_date=2023-02-21",
			ExpectedStatusCode: http.StatusSeeOther,
		},
	}

	for _, test := range tests {
		reqBody := test.StartDate
		reqBody = fmt.Sprintf("%s&%s", reqBody, test.EndDate)

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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("ReservationSummary handler returned wrong status code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

func TestRepository_ChooseRoom(t *testing.T) {
	// without session
	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.RequestURI = "/choose-room/1"

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.ChooseRoom)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
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
			ExpectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			TestName:           "Invalid  start date",
			RoomID:             "id=1",
			StartDate:          "s=invalid",
			EndDate:            "e=2050-01-02",
			ExpectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			TestName:           "Invalid  end date",
			RoomID:             "id=1",
			StartDate:          "s=2050-01-01",
			EndDate:            "e=invalid",
			ExpectedStatusCode: http.StatusTemporaryRedirect,
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
