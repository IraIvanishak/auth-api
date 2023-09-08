package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var db *sql.DB

type Credentials struct {
	Username string `json:"username", db:"username"`
	Password string `json:"password", db:"password"`
}

var store sessions.Store

func init() {
	authKey := securecookie.GenerateRandomKey(32)
	encrKey := securecookie.GenerateRandomKey(32)
	store = sessions.NewCookieStore(authKey, encrKey)
}

func main() {
	// Initialize the database connection
	var err error
	db, err = sql.Open("postgres", "host=localhost user=postgres password=coolproger dbname=testpet sslmode=disable")
	if err != nil {
		log.Fatalf("Error opening the database: %v", err)
	}
	defer db.Close()
	// Define HTTP routes and handlers
	http.HandleFunc("/log-in", LogIn)
	http.HandleFunc("/sign-up", SignUp)
	http.HandleFunc("/wellcome", Wellcome) // Corrected the endpoint name
	http.HandleFunc("/log-out", LogOut)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Hello")) })

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// SignUp handles user registration
func SignUp(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	// Decode user registration data from JSON
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Printf("Failed to decode credentials: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Hash the user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Insert user data into the database
	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", creds.Username, string(hashedPassword))
	if err != nil {
		log.Printf("Failed to insert into DB: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create a new session and save the user's username
	session, err := store.Get(r, "user-session")
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	session.Values["user"] = creds.Username
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Registration is successful. Welcome, " + creds.Username))
}

// LogIn handles user login
func LogIn(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var creds Credentials

	// Decode user login data from JSON
	err = json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Printf("Failed to decode credentials: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var found Credentials

	// Query the database to find the user
	err = db.QueryRow("SELECT username, password FROM users WHERE username = $1 LIMIT 1", creds.Username).Scan(&found.Username, &found.Password)

	if err == sql.ErrNoRows {
		log.Printf("User does not exist in the database.")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	} else if err != nil {
		log.Printf("Error executing query: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Compare the hashed password with the provided password
	err = bcrypt.CompareHashAndPassword([]byte(found.Password), []byte(creds.Password))

	if err != nil {
		log.Printf("Password does not match.")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Update the session with the user's username
	session.Values["user"] = creds.Username

	err = session.Save(r, w)
	if err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authorization is successful"))
}

// Welcome greets the logged-in user
func Wellcome(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Hello, " + session.Values["user"].(string)))
	w.WriteHeader(http.StatusOK)
}

// LogOut handles user logout
func LogOut(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	userNickname, ok := session.Values["user"].(string)
	if !ok {
		log.Println("You are not logged in")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Remove the user from the session
	delete(session.Values, "user")

	// Save the session to persist the changes
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("See you soon, " + userNickname))
}
