package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{
		name:               "home",
		url:                "/",
		method:             "GET",
		params:             []postData{},
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "about",
		url:                "/about",
		method:             "GET",
		params:             []postData{},
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "generals",
		url:                "/generals-quarters",
		method:             "GET",
		params:             []postData{},
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "majors",
		url:                "/majors-suite",
		method:             "GET",
		params:             []postData{},
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "search availability",
		url:                "/search-availability",
		method:             "GET",
		params:             []postData{},
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "contact",
		url:                "/contact",
		method:             "GET",
		params:             []postData{},
		expectedStatusCode: http.StatusOK,
	}, {
		name:               "make reservation",
		url:                "/make-reservation",
		method:             "GET",
		params:             []postData{},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "reservation summary",
		url:                "/reservation-summary",
		method:             "GET",
		params:             []postData{},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:   "post search availability",
		url:    "/search-availability",
		method: "POST",
		params: []postData{
			{key: "start", value: "2020-01-01"},
			{key: "end", value: "2020-01-02"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:   "post search availability json",
		url:    "/search-availability-json",
		method: "POST",
		params: []postData{
			{key: "start", value: "2020-01-01"},
			{key: "end", value: "2020-01-02"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:   "post make reservation",
		url:    "/make-reservation",
		method: "POST",
		params: []postData{
			{key: "first_name", value: "John"},
			{key: "last_name", value: "Smith"},
			{key: "email", value: "john@smith.com"},
			{key: "phone", value: "555-555-5555"},
		},
		expectedStatusCode: http.StatusOK,
	},
}

func TestHandlers(t *testing.T) {
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
		} else {
			values := url.Values{}
			for _, v := range tt.params {
				values.Add(v.key, v.value)
			}

			resp, err := ts.Client().PostForm(ts.URL+tt.url, values)

			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != tt.expectedStatusCode {
				t.Errorf("for %s expected %d but got %d", tt.name, tt.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}
