package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Vote struct {
	gorm.Model
	Title        string
	NumberVoters uint
	Choices      []Choice
}

type Choice struct {
	gorm.Model
	VoteID    uint
	Text      string
	Approvals uint
}

func main() {
	router := mux.NewRouter()
	db, err := gorm.Open(sqlite.Open("data/app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Unable to connect to database")
	}
	db.AutoMigrate(&Vote{})
	db.AutoMigrate(&Choice{})

	content := Root("Strangers")
	body := Page(content)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body.Render(r.Context(), w)
	}).Methods("GET")

	router.HandleFunc("/frank", func(w http.ResponseWriter, r *http.Request) {
		// Greeting("Frank").Render(r.Context(), w)
		NewVote().Render(r.Context(), w)
	}).Methods("POST")

	router.HandleFunc("/users", FormHandler).Methods("PATCH")

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", router)
}

func FormHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var sb strings.Builder
	sb.WriteString("Approved: ")
	for option, approval := range r.Form {
		approved := approval[0] == "on"
		if approved {
			sb.WriteString(option)
			sb.WriteString(" ")
		}
	}
	Toast(sb.String()).Render(r.Context(), w)
}
