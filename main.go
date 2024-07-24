package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/delaneyj/datastar"
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

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var vote Vote
		db.Preload("Choices").First(&vote)
		// fmt.Println(vote)
		content := Root("Strangers", vote)
		body := Page(content)
		body.Render(r.Context(), w)
	}).Methods("GET")

	router.HandleFunc("/frank", func(w http.ResponseWriter, r *http.Request) {
		// Greeting("Frank").Render(r.Context(), w)
		// NewVote().Render(r.Context(), w)
		// DataStar().Render(r.Context(), w)
		update, _ := gabs.ParseJSONBuffer(r.Body)
		fmt.Println(update.String())
		// This is how you send HTML down HTMX style
		datastar.RenderFragmentTempl(datastar.NewSSE(w, r), BetterDataStar())
		update.Set(true, "choice_pizza")
		// This is how you change the data-store
		datastar.PatchStore(datastar.NewSSE(w, r), update)
	}).Methods("POST")

	router.HandleFunc("/vote/{id}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		// ignoring error checking here which is bad under normal circumstances
		id, _ := strconv.Atoi(params["id"])
		var vote Vote
		db.Preload("Choices").First(&vote, id)
		// fmt.Println(vote)
		body := Page(VoteOne(vote))
		body.Render(r.Context(), w)
	}).Methods("GET")

	router.HandleFunc("/results/{id}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		// ignoring error checking here which is bad under normal circumstances
		id, _ := strconv.Atoi(params["id"])
		var vote Vote
		db.Preload("Choices").First(&vote, id)
		// fmt.Println(vote)

		// Something to note with this: When going directly to this URL
		// htmx will not be loaded so it is a static page. Coming from the vote apge
		// htmx is already loaded and the polling will happen. This to be acceptable.
		VoteResults(vote).Render(r.Context(), w)
	}).Methods("GET")

	router.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		// update, _ := gabs.ParseJSONBuffer(r.Body)
		// fmt.Println(update.String())
		// datastar.PatchStore(datastar.NewSSE(w, r), update)
		// datastar.RenderFragmentTempl(datastar.NewSSE(w, r), ThanksForVoting())
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
		// update, _ := gabs.ParseJSONBuffer(r.Body)
		// fmt.Println(update.String())
		// datastar.PatchStore(datastar.NewSSE(w, r), update)
		// datastar.RenderFragmentTempl(datastar.NewSSE(w, r), ThanksForVoting())

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
		VoteOne(vote).Render(r.Context(), w)
	}).Methods("POST")

	router.HandleFunc("/users", FormHandler).Methods("PATCH")

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
