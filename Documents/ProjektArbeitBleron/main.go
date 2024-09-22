package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var (
	mutex     = &sync.Mutex{}
	userData  = "data.json"
	eventData = "events.json"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Event struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Date         string `json:"date"`
	Location     string `json:"location"`
	CreatorID    int    `json:"creator_id"`
	Participants []int  `json:"participants"`
}

var events []Event
var users []User
var currentUserID int

func main() {
	loadUsers()
	loadEvents()

	http.HandleFunc("/", serveTemplate)
	http.HandleFunc("/register", RegisterUser)
	http.HandleFunc("/login", LoginUser)
	http.HandleFunc("/logout", LogoutUser)
	http.HandleFunc("/events", GetEvents)
	http.HandleFunc("/events/create", CreateEvent)
	http.HandleFunc("/events/details", GetEventDetails)
	http.HandleFunc("/events/join", JoinEvent)
	http.HandleFunc("/events/participants", GetEventParticipants)

	log.Println("Server läuft auf http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadUsers() {
	mutex.Lock()
	defer mutex.Unlock()

	file, err := ioutil.ReadFile(userData)
	if err == nil {
		json.Unmarshal(file, &users)
	}
}

func saveUsers() {
	mutex.Lock()
	defer mutex.Unlock()

	data, _ := json.Marshal(users)
	ioutil.WriteFile(userData, data, 0644)
}

func loadEvents() {
	mutex.Lock()
	defer mutex.Unlock()

	file, err := ioutil.ReadFile(eventData)
	if err == nil {
		json.Unmarshal(file, &events)
	}
}

func saveEvents() {
	mutex.Lock()
	defer mutex.Unlock()

	data, _ := json.Marshal(events)
	ioutil.WriteFile(eventData, data, 0644)
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("template.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	user.ID = len(users) + 1
	users = append(users, user)
	saveUsers()

	w.WriteHeader(http.StatusCreated)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	for _, u := range users {
		if u.Username == user.Username && u.Password == user.Password {
			currentUserID = u.ID
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Ungültige Anmeldedaten", http.StatusUnauthorized)
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	currentUserID = 0
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GetEvents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(events)
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	if currentUserID == 0 {
		http.Error(w, "Du musst angemeldet sein", http.StatusUnauthorized)
		return
	}

	var event Event
	json.NewDecoder(r.Body).Decode(&event)

	event.ID = len(events) + 1
	event.CreatorID = currentUserID
	events = append(events, event)

	saveEvents()

	w.WriteHeader(http.StatusCreated)
}

func GetEventDetails(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		http.Error(w, "Ungültige Event-ID", http.StatusBadRequest)
		return
	}

	for _, event := range events {
		if event.ID == eventID {
			var creatorUsername string
			for _, user := range users {
				if user.ID == event.CreatorID {
					creatorUsername = user.Username
					break
				}
			}

			eventDetails := struct {
				Event
				CreatorUsername string `json:"creator_username"`
			}{
				Event:           event,
				CreatorUsername: creatorUsername,
			}

			json.NewEncoder(w).Encode(eventDetails)
			return
		}
	}
	http.Error(w, "Event nicht gefunden", http.StatusNotFound)
}

func JoinEvent(w http.ResponseWriter, r *http.Request) {
	if currentUserID == 0 {
		http.Error(w, "Du musst angemeldet sein", http.StatusUnauthorized)
		return
	}

	var request struct {
		EventID int `json:"event_id"`
	}
	json.NewDecoder(r.Body).Decode(&request)

	for i, event := range events {
		if event.ID == request.EventID {
			events[i].Participants = append(events[i].Participants, currentUserID)
			saveEvents()
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Event nicht gefunden", http.StatusNotFound)
}

func GetEventParticipants(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		http.Error(w, "Ungültige Event-ID", http.StatusBadRequest)
		return
	}

	for _, event := range events {
		if event.ID == eventID {
			var participantUsernames []string
			for _, participantID := range event.Participants {
				for _, user := range users {
					if user.ID == participantID {
						participantUsernames = append(participantUsernames, user.Username)
						break
					}
				}
			}
			json.NewEncoder(w).Encode(participantUsernames)
			return
		}
	}
	http.Error(w, "Event nicht gefunden", http.StatusNotFound)
}
