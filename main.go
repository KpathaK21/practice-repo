package main

import (
	"net/http"
	"log"

	"github.com/KpathaK21/practice-repo/db"
	"github.com/KpathaK21/practice-repo/handlers"
)

func main() {
	db.Init()

	// Serve static HTML form
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/signup", handlers.SignUp)
	http.HandleFunc("/signin", handlers.SignIn)

	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
	
	
	
}
