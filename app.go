package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
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

type documentRequirements struct {
}

var db *sql.DB

func main() {
	// // Connecting to the MySQL Database
	var err error
	dsn := "root:your_password@tcp(localhost:8111)/whimpy_intern_db?parseTime=true"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// // API Initialization
	r := mux.NewRouter()
	fmt.Println("Please Wait...")
	fmt.Println("Initializing Server...")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
	}).Methods("GET")
	r.HandleFunc("/", BaseHandler)
	// Human Resource Handlers
	// r.HandleFunc("/hr/register", AddHRUserHandler).Methods("POST")
	// r.HandleFunc("/hr", GetHRUsersHandler).Methods("GET")
	// // Intern User Handlers
	r.HandleFunc("/intern/register", AddInternUserHandler).Methods("POST")
	r.HandleFunc("/intern", GetInternUsersHandler).Methods("GET")
	r.HandleFunc("/intern/update/{id}", UpdateInternHandler).Methods("UPDATE")
	r.HandleFunc("/intern/update-info/{id}", AdvancedInternUserUpdateHandler).Methods("PUT")
	r.HandleFunc("/getuser/{id}", GetInternUserByIDHandler).Methods("GET")
	r.HandleFunc("/intern/login/{id}", SigninInternUserHandler).Methods("POST")
	r.HandleFunc("/intern/logout/{id}", SignOutInternUserHandler).Methods("POST")
	r.HandleFunc("/intern/delete/{id}", DeleteInternUserHandler).Methods("DELETE")
	// // Mentor User Handlers
	// // r.HandleFunc("/addmentor", AddMentorUserHandler).Methods("POST")
	// // r.HandleFunc("/getmentors", GetMentorUsersHandler).Methods("GET")
	// // r.HandleFunc("/getmentor/{id}", GetMentorUserByIDHandler).Methods("GET")
	// // r.HandleFunc("/updatementor/{id}", UpdateMentorUserHandler).Methods("PUT")
	// // r.HandleFunc("/deletementor/{id}", DeleteMentorUserHandler).Methods("DELETE")
	// // Server Handler
	fmt.Println("█   █ █   █ ███ █   █ ████  █   █  ")
	fmt.Println("█░  █░█░  █░ █░░██ ██░█░░░█  █ █ ░ ")
	fmt.Println("█░█ █░█████░░█░░█░█ █░████░░  █ ░ ░")
	fmt.Println("██░██░█░░░█░░█░░█░░░█░█░░░░ ░ █░ ░ ")
	fmt.Println("█░░ █░█░░░█░███░█░░ █░█░░░░░  █░░  ")
	fmt.Println("Whimpy API Server (MySQL Database) | Created by: github.com/adam-ghafara")
	fmt.Println("See the Documentation guide to know more about the API functions.")
	if err := http.ListenAndServe(":2105", r); err != nil {
		log.Fatal(err)
		return
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

func RandomIDGenerator() int {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
}

func AddInternUserHandler(w http.ResponseWriter, r *http.Request) {
	var u internUser
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Retry Generation if exists
	for attempt := 0; attempt < 5; attempt++ {
		u.InternID = RandomIDGenerator()
		_, err := db.Exec("INSERT INTO wp_intern_list (internID, internUsername, internFullname, internEmail, internPassword) VALUES (?,?,?,?,?)", u.InternID, u.Username, u.FullName, u.Email, u.Password)
		if err == nil {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(u)
			return
		}

		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			if strings.Contains(me.Message, "for key 'wp_intern_list.intern_id'") || strings.Contains(me.Message, "Duplicate entry") {
				continue // Retry if the error is due to duplicate intern_id
			}
			http.Error(w, "Item already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		// var u internUser
		// // generate random number for intern_id
		// u.InternID = rand.Intn(1000000)
		// if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		// 	http.Error(w, "Invalid JSON", http.StatusBadRequest)
		// }
		// res, err := db.Exec("INSERT INTO wp_intern_list (intern_id, intern_username, intern_fullname, intern_email, intern_password) VALUES (?,?,?,?,?)", u.Username, u.FullName, u.Email, u.Password)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// id, _ := res.LastInsertId()
		// u.InternID = int(id)
		// w.WriteHeader(http.StatusAccepted)
		// json.NewEncoder(w).Encode(u)
	}
}

func GetInternUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT internId, internUsername, internFullname, internEmail, internPassword from wp_intern_list")
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
	upt, err := db.Exec("UPDATE wp_intern_list SET internUsername = ?, internFullname = ?, internEmail = ? WHERE intern_id = ?", u.Username, u.FullName, u.Email, getID)
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
	err := db.QueryRow("SELECT internId, internUsername, internFullname, internEmail, internPassword FROM wp_intern_list WHERE intern_id = ?", getID).Scan(&u.InternID, &u.Username, &u.FullName, &u.Email, &u.Password)
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

// Signin/Signout Handler + SessionIDGenerator + Send to Cookie

const sessionCookieName = "sessionID"

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func SessionIDGenerator() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", b)
}

func SigninInternUserHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var userID = mux.Vars(r)["id"]
	var hash string
	err := db.QueryRow("select internID, password_hash FROM wp_intern_list WHERE internUsername = ?", req.Username).Scan(&userID, &hash)
	if err == sql.ErrNoRows || (err == nil && bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sessionID := SessionIDGenerator()
	if err != nil {
		http.Error(w, "Serve Error..", http.StatusInternalServerError)
	}
	expires := time.Now().Add(24 * time.Hour)

	_, err = db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expires)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,                 // JS can't read it (blocks XSS theft)
		Secure:   false,                // set to true in production (HTTPS)
		SameSite: http.SameSiteLaxMode, // basic CSRF protection
	})

	json.NewEncoder(w).Encode(map[string]any{"message": "logged in", "user_id": userID})
}

func SignOutInternUserHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookieName); err == nil {
		db.Exec("DELETE FROM sessions WHERE id = ?", c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: "", Path: "/",
		MaxAge: -1, HttpOnly: true,
	})
	w.WriteHeader(http.StatusNoContent)
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
