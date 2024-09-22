package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	user := User{Username: "testuser", Password: "password123"}
	userData, _ := json.Marshal(user)

	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(userData))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RegisterUser)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	loadUsers()
	if len(users) != 1 || users[0].Username != "testuser" {
		t.Errorf("Benutzer wurde nicht korrekt registriert")
	}
}

func TestLoginUser(t *testing.T) {
	TestRegisterUser(t)

	user := User{Username: "testuser", Password: "password123"}
	userData, _ := json.Marshal(user)

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(userData))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(LoginUser)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if currentUserID != 1 {
		t.Errorf("Benutzer wurde nicht korrekt angemeldet")
	}
}

func TestCreateEvent(t *testing.T) {
	TestRegisterUser(t)
	TestLoginUser(t)

	event := Event{Title: "Test Event", Description: "Ein Testevent", Date: "2023-09-22", Location: "Testort"}
	eventData, _ := json.Marshal(event)

	req, err := http.NewRequest("POST", "/events/create", bytes.NewBuffer(eventData))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateEvent)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if len(events) != 1 || events[0].Title != "Test Event" {
		t.Errorf("Event wurde nicht korrekt erstellt")
	}
}

func TestGetEvents(t *testing.T) {
	TestCreateEvent(t)

	req, err := http.NewRequest("GET", "/events", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetEvents)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var fetchedEvents []Event
	json.Unmarshal(rr.Body.Bytes(), &fetchedEvents)
	if len(fetchedEvents) != 1 || fetchedEvents[0].Title != "Test Event" {
		t.Errorf("Events wurden nicht korrekt abgerufen")
	}
}

func TestJoinEvent(t *testing.T) {
	TestCreateEvent(t)

	reqBody := struct {
		EventID int `json:"event_id"`
	}{EventID: events[0].ID}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/events/join", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(JoinEvent)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if len(events[0].Participants) != 1 || events[0].Participants[0] != 1 {
		t.Errorf("Benutzer wurde nicht korrekt als Teilnehmer hinzugefügt")
	}
}

func TestGetEventParticipants(t *testing.T) {
	TestJoinEvent(t)

	req, err := http.NewRequest("GET", "/events/participants?id="+strconv.Itoa(events[0].ID), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetEventParticipants)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var participantUsernames []string
	json.Unmarshal(rr.Body.Bytes(), &participantUsernames)
	if len(participantUsernames) != 1 || participantUsernames[0] != "testuser" {
		t.Errorf("Teilnehmer wurden nicht korrekt abgerufen")
	}
}
