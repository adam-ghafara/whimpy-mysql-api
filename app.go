package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// Structure for user data insertion and retrieval to mysql
type internUser struct {
	InternID    int    `json:"intern_id"`
	Username    string `json:"intern_username"`
	FullName    string `json:"intern_fullname"`
	Email       string `json:"intern_email"`
	Password    string `json:"intern_password"`
	PhoneNumber int    `json:"intern_phone"`
}

type mentorUser struct {
	Username  string `json:"mentor_username"`
	FullName  string `json:"mentor_fullname"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type detailsUserinfo struct {
	Education    string `json:"intern_edu"`
	Major        string `json:"intern_major"`
	DepartmentID int    `json:"department_id"`
}

var db *sql.DB

func main() {
	// // Connecting to the MySQL Database
	var err error
	dsn := "root:your_password@tcp(localhost:8111)/mydb?parseTime=true"
	db, err = sql.Open("whimpy_intern_db", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// // API Initialization
	r := mux.NewRouter()
	fmt.Println("Initializing API...")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "█   █ █   █ ███ █   █ ████  █   █  ")
		fmt.Fprintf(w, "█░  █░█░  █░ █░░██ ██░█░░░█  █ █ ░ ")
		fmt.Fprintf(w, "█░█ █░█████░░█░░█░█ █░████░░  █ ░ ░")
		fmt.Fprintf(w, "██░██░█░░░█░░█░░█░░░█░█░░░░ ░ █░ ░ ")
		fmt.Fprintf(w, "█░░ █░█░░░█░███░█░░ █░█░░░░░  █░░  ")
		fmt.Fprintf(w, "Whimpy API Server (MySQL Database) | Created by: github.com/adam-ghafara")
		fmt.Fprintf(w, "See the Documentation guide to know more about the API functions.")
		fmt.Fprintf(w, "------------------------------------------------------------------------")
		fmt.Fprintf(w, "Swagger Page : localhost:2105/swagger")
		db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/whimpy_intern_db")
		if err != nil {
			panic(err.Error())
		}
		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
	}).Methods("GET")
	r.HandleFunc("/", BaseHandler)
	// // Intern User Handlers
	r.HandleFunc("/intern/register", AddInternUserHandler).Methods("POST")
	r.HandleFunc("/intern", GetInternUsersHandler).Methods("GET")
	r.HandleFunc("/intern/update/{id}", UpdateInternHandler).Methods("UPDATE")
	r.HandleFunc("/intern/advanced-update/{id}", AdvancedInternUserUpdateHandler).Methods("PUT")
	r.HandleFunc("/getuser/{id}", GetInternUserByIDHandler).Methods("GET")
	r.HandleFunc("/intern/login/{id}", LoginInternUserHandler).Methods("POST")
	r.HandleFunc("/intern/delete/{id}", DeleteInternUserHandler).Methods("DELETE")
	// r.HandleFunc("/updateuser/{id}", UpdateInternUserHandler).Methods("PUT")
	// r.HandleFunc("/deleteuser/{id}", DeleteInternUserHandler).Methods("DELETE")
	// // Mentor User Handlers
	// // r.HandleFunc("/addmentor", AddMentorUserHandler).Methods("POST")
	// // r.HandleFunc("/getmentors", GetMentorUsersHandler).Methods("GET")
	// // r.HandleFunc("/getmentor/{id}", GetMentorUserByIDHandler).Methods("GET")
	// // r.HandleFunc("/updatementor/{id}", UpdateMentorUserHandler).Methods("PUT")
	// // r.HandleFunc("/deletementor/{id}", DeleteMentorUserHandler).Methods("DELETE")
	// // Server Handler
	log.Println("Initializing Server...")
	if err := http.ListenAndServe(":2105", r); err != nil {
		log.Fatal(err)
	}
}

// This Function is responsibility for being used as a Basic API Test.
func BaseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "█   █ █   █ ███ █   █ ████  █   █  ")
	fmt.Fprintf(w, "█░  █░█░  █░ █░░██ ██░█░░░█  █ █ ░ ")
	fmt.Fprintf(w, "█░█ █░█████░░█░░█░█ █░████░░  █ ░ ░")
	fmt.Fprintf(w, "██░██░█░░░█░░█░░█░░░█░█░░░░ ░ █░ ░ ")
	fmt.Fprintf(w, "█░░ █░█░░░█░███░█░░ █░█░░░░░  █░░  ")
	fmt.Fprintf(w, "Whimpy API Server (MySQL Database) | Created by: github.com/adam-ghafara")
}

// CRUD List (Mentor, Intern)
func AddInternUserHandler(w http.ResponseWriter, r *http.Request) {
	var u internUser
	// generate random number for intern_id
	u.InternID = rand.Intn(1000000)
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}
	res, err := db.Exec("INSERT INTO wp_intern_list (intern_id, intern_username, intern_fullname, intern_email, intern_password) VALUES (?,?,?,?,?)", u.Username, u.FullName, u.Email, u.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	u.InternID = int(id)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(u)
}

func GetInternUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT intern_id, intern_username, intern_fullname, intern_email, intern_password from wp_intern_list")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	intern_list := []internUser{}
	for rows.Next() {
		var u internUser
		if err := rows.Scan(&u.InternID, u.Username, u.FullName, u.Email, u.Password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		intern_list = append(intern_list, u)
	}
	json.NewEncoder(w).Encode(intern_list)
}

func UpdateInternHandler(w http.ResponseWriter, r *http.Request) {
	getID := mux.Vars(r)["id"]
	var u internUser
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	upt, err := db.Exec("UPDATE wp_intern_list SET intern_username = ?, intern_fullname = ?, intern_email = ? WHERE intern_id = ?", u.Username, u.FullName, u.Email, getID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := upt.RowsAffected(); n == 0 {
		http.Error(w, "User Not Found (or Nothing Changed)", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func AdvancedInternUserUpdateHandler(w http.ResponseWriter, r *http.Request) {

}

func GetInternUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	getID := mux.Vars(r)["id"]
	var u internUser
	err := db.QueryRow("SELECT intern_id, intern_username, intern_fullname, intern_email, intern_password FROM wp_intern_list WHERE intern_id = ?", getID).Scan(&u.InternID, &u.Username, &u.FullName, &u.Email, &u.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User Not Found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(u)
}

// Login and Create a SessionID, Send the SessionID to the Client.
func LoginInternUserHandler(w http.ResponseWriter, r *http.Request) {
	getID := mux.Vars(r)["id"]
	var u internUser
	err := db.QueryRow("SELECT intern_id, intern_username, intern_fullname, intern_email, intern_password FROM wp_intern_list WHERE intern_id = ?", getID).Scan(&u.InternID, &u.Username, &u.FullName, &u.Email, &u.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User Not Found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(u)
}

func DeleteInternUserHandler(w http.ResponseWriter, r *http.Request) {
	getID := mux.Vars(r)["id"]
	del, err := db.Exec("DELETE FROM wp_intern_list WHERE intern_id = ?", getID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := del.RowsAffected(); n == 0 {
		http.Error(w, "User Not Found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Passive Operations for Inactive Userss
func SendEmailToInactiveUsers(w http.ResponseWriter, r *http.Request) {
	getID := mux.Vars(r)["id"]
	// Implementation for sending email to inactive users
	// This is a placeholder for the actual email sending logic
	log.Println("Email sent to inactive users with ID:", getID)
}

func DeleteInactiveUsers(w http.ResponseWriter, r *http.Request) {
	getID := mux.Vars(r)["id"]
	_, err := db.Exec("DELETE FROM wp_intern_list WHERE intern_id = ?", getID)
	if err != nil {
		log.Println("Error deleting inactive users:", err)
		return
	}
	log.Println("Inactive users deleted successfully")
}
