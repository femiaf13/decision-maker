package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
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

	// This will serve files under ./static/<filename>
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var votes []Vote
		db.Preload("Choices").Find(&votes)
		// fmt.Println(votes)
		Page(Root(votes)).Render(r.Context(), w)
	}).Methods("GET")

	router.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		var votes []Vote
		db.Preload("Choices").Find(&votes)
		body := Page(AllVotesTable(votes))
		body.Render(r.Context(), w)
	}).Methods("GET")

	router.HandleFunc("/vote/{id}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		// ignoring error checking here which is bad under normal circumstances
		id, _ := strconv.Atoi(params["id"])
		var vote Vote
		db.Preload("Choices").First(&vote, id)
		body := Page(VoteTemplate(vote))
		body.Render(r.Context(), w)
	}).Methods("GET")

	router.HandleFunc("/results/{id}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		// ignoring error checking here which is bad under normal circumstances
		id, _ := strconv.Atoi(params["id"])
		var vote Vote
		db.Preload("Choices").First(&vote, id)
		Page(VoteResults(vote)).Render(r.Context(), w)
	}).Methods("GET")

	router.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		// fmt.Println(r.Form)

		id, ok := r.Form["id"]
		if ok {
			var vote Vote
			db.Preload("Choices").First(&vote, id)
			// fmt.Println(vote)
			vote.NumberVoters++
			for key := range r.Form {
				for idx, choice := range vote.Choices {
					if key == choice.Text {
						vote.Choices[idx].Approvals++
					}
				}
			}
			// fmt.Println(vote)
			db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&vote)
			VoteResults(vote).Render(r.Context(), w)
		}
	}).Methods("PATCH")

	router.HandleFunc("/newvote", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		// fmt.Println(r.Form)

		var choices []Choice = make([]Choice, 0)
		var vote Vote
		for key, value := range r.Form {
			// fmt.Println(key)
			// fmt.Println(value[0])
			if strings.HasPrefix(key, "choice_") && len(value[0]) > 0 {
				choices = append(choices, Choice{
					Text: value[0],
				})
			}
			if key == "title" {
				vote = Vote{Title: value[0]}
			}
		}
		// fmt.Println(choices)
		vote.Choices = choices
		// fmt.Println(vote)
		db.Create(&vote)
		// fmt.Println(vote)
		VoteTemplate(vote).Render(r.Context(), w)
	}).Methods("POST")

	router.HandleFunc("/newvote", func(w http.ResponseWriter, r *http.Request) {
		CreateNewVote().Render(r.Context(), w)
	}).Methods("GET")

	router.HandleFunc("/newchoice", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		// fmt.Println(len(r.Form))
		numOptions := len(r.Form) - 1
		// fmt.Println(numOptions)
		CreateNewChoice(uint(numOptions), (numOptions < 4)).Render(r.Context(), w)
	}).Methods("POST")

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", router)
}
