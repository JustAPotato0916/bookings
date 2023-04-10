package handlers

import (
	"context"
	"fmt"
	"github.com/JustAPotato0916/bookings/internal/models"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},

	//{"post-search-availability", "/search-availability", "POST", []postData{
	//	{key: "start", value: "2023-03-26"},
	//	{key: "end", value: "2023-03-27"},
	//}, http.StatusOK},
	//{"post-search-availability-json", "/search-availability-json", "POST", []postData{
	//	{key: "start", value: "2023-03-26"},
	//	{key: "end", value: "2023-03-27"},
	//}, http.StatusOK},
	//{"make reservation post", "/make-reservation", "POST", []postData{
	//	{key: "first_name", value: "Eric"},
	//	{key: "last_name", value: "Popo"},
	//	{key: "email", value: "me@here.com"},
	//	{key: "phone", value: "555-555-1111"},
	//}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// test case where reservation is not in session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test with non-existent room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, "2050-01-01")
	endDate, _ := time.Parse(layout, "2050-01-02")
	reservation := models.Reservation{
		RoomID:    1,
		StartDate: startDate,
		EndDate:   endDate,
	}

	reqBody := "first_name=popopo"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Chen")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=popo@chen.com")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test for missing session
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing session: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for missing post body
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	session.Put(ctx, "reservation", reservation)
	req.Header.Set("Content-Type", "application/x-www-from-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid data
	reqBody = "first_name=f"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Chen")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=popo@chen.com")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-from-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code for invalid data: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	}

	// test for failure to insert reservation into database
	//reqBody = "first_name=popopo"
	//reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Chen")
	//reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	//reqBody = fmt.Sprintf("%s&%s", reqBody, "email=popo@chen.com")
	//
	//req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	//ctx = getCtx(req)
	//
	//reservation.RoomID = 2
	//session.Put(ctx, "reservation", reservation)
	//
	//req = req.WithContext(ctx)
	//req.Header.Set("Content-Type", "application/x-www-from-urlencoded")
	//rr = httptest.NewRecorder()
	//
	//handler = http.HandlerFunc(Repo.PostReservation)
	//
	//handler.ServeHTTP(rr, req)
	//
	//if rr.Code != http.StatusTemporaryRedirect {
	//	t.Errorf("PostReservation handler failed when trying to fail inserting reservation: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	//}

	// test for failure to insert restriction into database
	//reqBody = "first_name=popopo"
	//reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Chen")
	//reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	//reqBody = fmt.Sprintf("%s&%s", reqBody, "email=popo@chen.com")
	//
	//req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	//ctx = getCtx(req)
	//
	//reservation.RoomID = 1000
	//session.Put(ctx, "reservation", reservation)
	//
	//req = req.WithContext(ctx)
	//req.Header.Set("Content-Type", "application/x-www-from-urlencoded")
	//rr = httptest.NewRecorder()
	//
	//handler = http.HandlerFunc(Repo.PostReservation)
	//
	//handler.ServeHTTP(rr, req)
	//
	//if rr.Code != http.StatusTemporaryRedirect {
	//	t.Errorf("PostReservation handler failed when trying to fail inserting reservation: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	//}
}

//func TestRepository_AvailabilityJSON(t *testing.T) {
//	/*****************************************
//	// first case -- rooms are not available
//	*****************************************/
//	// create our request body
//	reqBody := "start=2050-01-01"
//	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2050-01-02")
//	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
//
//	// create our request
//	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
//
//	// get the context with session
//	ctx := getCtx(req)
//	req = req.WithContext(ctx)
//
//	// set the request header
//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//	// create our response recorder, which satisfies the requirements
//	// for http.ResponseWriter
//	rr := httptest.NewRecorder()
//
//	// make our handler a http.HandlerFunc
//	handler := http.HandlerFunc(Repo.AvailabilityJSON)
//
//	// make the request to our handler
//	handler.ServeHTTP(rr, req)
//
//	// since we have no rooms available, we expect to get status http.StatusSeeOther
//	// this time we want to parse JSON and get the expected response
//	var j jsonResponse
//	err := json.Unmarshal([]byte(rr.Body.String()), &j)
//	if err != nil {
//		t.Error("failed to parse json!")
//	}
//
//	// since we specified a start date > 2049-12-31, we expect no availability
//	if j.OK {
//		t.Error("Got availability when none was expected in AvailabilityJSON")
//	}
//
//	/*****************************************
//	// second case -- rooms not available
//	*****************************************/
//	// create our request body
//	reqBody = "start=2040-01-01"
//	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2040-01-02")
//	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
//
//	// create our request
//	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
//
//	// get the context with session
//	ctx = getCtx(req)
//	req = req.WithContext(ctx)
//
//	// set the request header
//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//	// create our response recorder, which satisfies the requirements
//	// for http.ResponseWriter
//	rr = httptest.NewRecorder()
//
//	// make our handler a http.HandlerFunc
//	handler = http.HandlerFunc(Repo.AvailabilityJSON)
//
//	// make the request to our handler
//	handler.ServeHTTP(rr, req)
//
//	// this time we want to parse JSON and get the expected response
//	err = json.Unmarshal([]byte(rr.Body.String()), &j)
//	if err != nil {
//		t.Error("failed to parse json!")
//	}
//
//	// since we specified a start date < 2049-12-31, we expect availability
//	if !j.OK {
//		t.Error("Got no availability when some was expected in AvailabilityJSON")
//	}
//
//	/*****************************************
//	// third case -- no request body
//	*****************************************/
//	// create our request
//	req, _ = http.NewRequest("POST", "/search-availability-json", nil)
//
//	// get the context with session
//	ctx = getCtx(req)
//	req = req.WithContext(ctx)
//
//	// set the request header
//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//	// create our response recorder, which satisfies the requirements
//	// for http.ResponseWriter
//	rr = httptest.NewRecorder()
//
//	// make our handler a http.HandlerFunc
//	handler = http.HandlerFunc(Repo.AvailabilityJSON)
//
//	// make the request to our handler
//	handler.ServeHTTP(rr, req)
//
//	// this time we want to parse JSON and get the expected response
//	err = json.Unmarshal([]byte(rr.Body.String()), &j)
//	if err != nil {
//		t.Error("failed to parse json!")
//	}
//
//	// since we specified a start date < 2049-12-31, we expect availability
//	if j.OK || j.Message != "Internal server error" {
//		t.Error("Got availability when request body was empty")
//	}
//
//	/*****************************************
//	// fourth case -- database error
//	*****************************************/
//	// create our request body
//	reqBody = "start=2060-01-01"
//	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2060-01-02")
//	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
//	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
//
//	// get the context with session
//	ctx = getCtx(req)
//	req = req.WithContext(ctx)
//
//	// set the request header
//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//
//	// create our response recorder, which satisfies the requirements
//	// for http.ResponseWriter
//	rr = httptest.NewRecorder()
//
//	// make our handler a http.HandlerFunc
//	handler = http.HandlerFunc(Repo.AvailabilityJSON)
//
//	// make the request to our handler
//	handler.ServeHTTP(rr, req)
//
//	// this time we want to parse JSON and get the expected response
//	err = json.Unmarshal([]byte(rr.Body.String()), &j)
//	if err != nil {
//		t.Error("failed to parse json!")
//	}
//
//	// since we specified a start date < 2049-12-31, we expect availability
//	if j.OK || j.Message != "Error querying database" {
//		t.Error("Got availability when simulating database error")
//	}
//}

//func TestRepository_ReservationSummary(t *testing.T) {
//	/*****************************************
//	// first case -- reservation in session
//	*****************************************/
//	reservation := models.Reservation{
//		RoomID: 1,
//		Room: models.Room{
//			ID:       1,
//			RoomName: "General's Quarters",
//		},
//	}
//
//	req, _ := http.NewRequest("GET", "/reservation-summary", nil)
//	ctx := getCtx(req)
//	req = req.WithContext(ctx)
//
//	rr := httptest.NewRecorder()
//	session.Put(ctx, "reservation", reservation)
//
//	handler := http.HandlerFunc(Repo.ReservationSummary)
//
//	handler.ServeHTTP(rr, req)
//
//	if rr.Code != http.StatusOK {
//		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
//	}
//
//	/*****************************************
//	// second case -- reservation not in session
//	*****************************************/
//	req, _ = http.NewRequest("GET", "/reservation-summary", nil)
//	ctx = getCtx(req)
//	req = req.WithContext(ctx)
//
//	rr = httptest.NewRecorder()
//
//	handler = http.HandlerFunc(Repo.ReservationSummary)
//
//	handler.ServeHTTP(rr, req)
//
//	if rr.Code != http.StatusTemporaryRedirect {
//		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
//	}
//}
//
//func TestRepository_ChooseRoom(t *testing.T) {
//	/*****************************************
//	// first case -- reservation in session
//	*****************************************/
//	reservation := models.Reservation{
//		RoomID: 1,
//		Room: models.Room{
//			ID:       1,
//			RoomName: "General's Quarters",
//		},
//	}
//
//	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
//	ctx := getCtx(req)
//	req = req.WithContext(ctx)
//	// set the RequestURI on the request so that we can grab the ID
//	// from the URL
//	req.RequestURI = "/choose-room/1"
//
//	rr := httptest.NewRecorder()
//	session.Put(ctx, "reservation", reservation)
//
//	handler := http.HandlerFunc(Repo.ChooseRoom)
//
//	handler.ServeHTTP(rr, req)
//
//	if rr.Code != http.StatusSeeOther {
//		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
//	}
//
//	///*****************************************
//	//// second case -- reservation not in session
//	//*****************************************/
//	req, _ = http.NewRequest("GET", "/choose-room/1", nil)
//	ctx = getCtx(req)
//	req = req.WithContext(ctx)
//	req.RequestURI = "/choose-room/1"
//
//	rr = httptest.NewRecorder()
//
//	handler = http.HandlerFunc(Repo.ChooseRoom)
//
//	handler.ServeHTTP(rr, req)
//
//	if rr.Code != http.StatusTemporaryRedirect {
//		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
//	}
//
//	///*****************************************
//	//// third case -- missing url parameter, or malformed parameter
//	//*****************************************/
//	req, _ = http.NewRequest("GET", "/choose-room/fish", nil)
//	ctx = getCtx(req)
//	req = req.WithContext(ctx)
//	req.RequestURI = "/choose-room/fish"
//
//	rr = httptest.NewRecorder()
//
//	handler = http.HandlerFunc(Repo.ChooseRoom)
//
//	handler.ServeHTTP(rr, req)
//
//	if rr.Code != http.StatusTemporaryRedirect {
//		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
//	}
//}
//
//func TestRepository_BookRoom(t *testing.T) {
//	/*****************************************
//	// first case -- database works
//	*****************************************/
//	reservation := models.Reservation{
//		RoomID: 1,
//		Room: models.Room{
//			ID:       1,
//			RoomName: "General's Quarters",
//		},
//	}
//
//	req, _ := http.NewRequest("GET", "/book-room?s=2050-01-01&e=2050-01-02&id=1", nil)
//	ctx := getCtx(req)
//	req = req.WithContext(ctx)
//
//	rr := httptest.NewRecorder()
//	session.Put(ctx, "reservation", reservation)
//
//	handler := http.HandlerFunc(Repo.BookRoom)
//
//	handler.ServeHTTP(rr, req)
//
//	if rr.Code != http.StatusSeeOther {
//		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusSeeOther)
//	}
//
//	/*****************************************
//	// second case -- database failed
//	*****************************************/
//	req, _ = http.NewRequest("GET", "/book-room?s=2040-01-01&e=2040-01-02&id=4", nil)
//	ctx = getCtx(req)
//	req = req.WithContext(ctx)
//
//	rr = httptest.NewRecorder()
//
//	handler = http.HandlerFunc(Repo.BookRoom)
//
//	handler.ServeHTTP(rr, req)
//
//	if rr.Code != http.StatusTemporaryRedirect {
//		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
//	}
//}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
