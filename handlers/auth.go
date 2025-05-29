package handlers

import (
	"net/http"

	"github.com/KpathaK21/practice-repo/db"
	"github.com/KpathaK21/practice-repo/models"
)

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

	// Optional: check for existing user
	var existing models.User
	if err := db.DB.Where("username = ? OR email = ?", username, email).First(&existing).Error; err == nil {
		http.Error(w, "Username or email already taken", http.StatusConflict)
		return
	}

	user := models.User{
		Username: username,
		Email:    email,
		Password: password,
	}
	if err := user.SetPassword(password); err != nil {
		http.Error(w, "Password hashing failed", http.StatusInternalServerError)
		return
	}
	if err := db.DB.Create(&user).Error; err != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Sign up successful!"))
}



func SignIn(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	var stored models.User
	email := r.FormValue("email")
	password := r.FormValue("password")

	if err := db.DB.Where("email = ?", email).First(&stored).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	if !stored.CheckPassword(password) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}
	w.Write([]byte("Sign in successful"))
}
