package handlers

import (
	"net/http"

	"github.com/KpathaK21/practice-repo/db"
	"github.com/KpathaK21/practice-repo/models"
)

// SignUp handles POST /signup
func SignUp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")

	if password != confirm {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existing models.User
	if err := db.DB.Where("username = ? OR email = ?", username, email).First(&existing).Error; err == nil {
		http.Error(w, "Username or email already taken", http.StatusConflict)
		return
	}

	// Create and hash password
	user := models.User{
		Username: username,
		Email:    email,
		Password: password,
	}
	if err := user.SetPassword(password); err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// SignIn handles POST /signin
func SignIn(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if !user.CheckPassword(password) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// For now, just respond with success (no session or JWT yet)
	w.Write([]byte("Sign in successful!"))
}
