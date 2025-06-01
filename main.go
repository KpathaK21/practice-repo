package main

import (
	"log"
	"net/http"

	"github.com/KpathaK21/practice-repo/db"
	"github.com/KpathaK21/practice-repo/handlers"
)

func main() {
	// Initialize environment variables
	handlers.InitEnvVars()

	// Initialize database
	db.Init()

	// Serve static HTML form
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// Auth routes
	http.HandleFunc("/signup", handlers.SignUp)
	http.HandleFunc("/signin", handlers.SignIn)
	http.HandleFunc("/verify", handlers.Verify)
	http.HandleFunc("/dashboard", handlers.Dashboard)

	// Course routes
	http.HandleFunc("/courses", handlers.ListCourses)
	http.HandleFunc("/course/create", handlers.ProfessorOnly(handlers.CreateCourse))
	http.HandleFunc("/course/", handlers.ViewCourse) // This should be last as it's a catch-all

	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
