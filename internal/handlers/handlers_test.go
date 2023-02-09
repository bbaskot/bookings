package handlers

import (
	"fmt"
	"log"
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
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"room", "/room", "GET", []postData{}, http.StatusOK},
	{"apartment", "/apartment", "GET", []postData{}, http.StatusOK},
	{"reservation", "/reservation", "GET", []postData{}, http.StatusOK},
	{"search-availability", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"Postreservation", "/reservation", "POST", []postData{
		{
			key:   "first_name",
			value: "MIrko",
		},
		{
			key:   "last_name",
			value: "Markovic",
		},
		{
			key:   "email",
			value: "mirko@gmail.com",
		},
		{
			key:   "phone_number",
			value: "666555",
		},
	}, http.StatusOK},
	{"PostSearchAvailability", "/search-availability", "POST", []postData{
		{
			key:   "start",
			value: "2022-01-01",
		},
		{
			key:   "end",
			value: "2023-02-02",
		},
	}, http.StatusOK},
	{"PostSearchAvailabilityJson", "/search-availability-json", "POST", []postData{
		{
			key:   "start",
			value: "2022-01-01",
		},
		{
			key:   "end",
			value: "2023-02-02",
		},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != e.expectedStatusCode {
				log.Println(fmt.Sprintf("for %s excpetced %d and got %d", e.name, e.expectedStatusCode, resp.StatusCode))
				t.Error(fmt.Sprintf("for %s excpetced %d and got %d", e.name, e.expectedStatusCode, resp.StatusCode))
			}
		} else {
			values := url.Values{}
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}
			resp, err := ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}
			if resp.StatusCode != e.expectedStatusCode {
				log.Println(fmt.Sprintf("for %s excpetced %d and got %d", e.name, e.expectedStatusCode, resp.StatusCode))
				t.Error(fmt.Sprintf("for %s excpetced %d and got %d", e.name, e.expectedStatusCode, resp.StatusCode))
			}
		}

	}

}
