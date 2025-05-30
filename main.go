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

	http.HandleFunc("/signup", handlers.SignUp)
	http.HandleFunc("/signin", handlers.SignIn)
	http.HandleFunc("/verify", handlers.Verify)
	http.HandleFunc("/dashboard", handlers.Dashboard)

	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
