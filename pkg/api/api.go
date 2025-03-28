package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/geofpwhite/ufc_db_go/pkg/database"
)

type handler func(w http.ResponseWriter, r *http.Request)

func fighterHandler() (handler, handler, handler) {
	db := database.Init()
	byPage := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		switch r.Method {
		case http.MethodGet:
			page := r.PathValue("page")
			pageNum, err := strconv.Atoi(page)
			if err != nil {
				http.Error(w, "Invalid page number", http.StatusBadRequest)
			}
			fighters, err := db.GetAllFighters(pageNum)
			if err != nil {
				fmt.Println(err)
				http.Error(w, "Fighter not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			jsonResponse, err := json.Marshal(fighters)
			if err != nil {
				http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
				return
			}
			w.Write(jsonResponse)
			return
		}
	}
	byName := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusBadRequest)
			fName, lName := r.PathValue("firstName"), r.PathValue("lastName")
			if fName == "" && lName == "" {
				http.Error(w, "Fighter not found", http.StatusNotFound)
				return
			}
			fighter, err := db.GetFighterByFirstAndLastName(fName, lName)
			if err != nil {
				http.Error(w, "Fighter not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			jsonResponse, err := json.Marshal(fighter)
			if err != nil {
				http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
				return
			}
			w.Write(jsonResponse)
			return
		}

	}
	byID := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		switch r.Method {
		case http.MethodPost:
			var id string
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}
			err = json.Unmarshal(body, &id)
			if err != nil {
				http.Error(w, "Error unmarshaling JSON", http.StatusBadRequest)
				return
			}

			defer r.Body.Close()

			idNum, err := strconv.Atoi(id)
			if err != nil {
				fName, lName := r.PathValue("firstName"), r.PathValue("lastName")
				if fName == "" && lName == "" {
					http.Error(w, "Fighter not found", http.StatusNotFound)
					return
				}
				fighter, err := db.GetFighterByFirstAndLastName(fName, lName)
				if err != nil {
					http.Error(w, "Fighter not found", http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusOK)
				jsonResponse, err := json.Marshal(fighter)
				if err != nil {
					http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
					return
				}
				w.Write(jsonResponse)

				break
			}
			fighter, err := db.GetFighterByID(uint(idNum))
			if err != nil {
				http.Error(w, "Fighter not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "Fighter: %s %s\n", fighter.FirstName, fighter.LastName)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}
	return byID, byName, byPage
}

func Serve() {
	mux := http.NewServeMux()
	byID, byName, byPage := fighterHandler()
	mux.HandleFunc("/fighters/all/{page}", byPage)
	mux.HandleFunc("/fighters/{id}", byID)
	mux.HandleFunc("/fighters/{firstName}/{lastName}", byName)
	fmt.Println("Server starting on port 8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
