package main

import (
	"net/http"

	"github.com/KpathaK21/practice-repo/db"
	"github.com/KpathaK21/practice-repo/handlers"
)

func main() {
	db.Init()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>Sign Up / Sign In</title></head>
			<body>
				<h2>Sign Up</h2>
				<form action="/signup" method="POST">
					<input name="username" placeholder="Username" required><br>
					<input name="email" placeholder="Email" required><br>
					<input type="password" name="password" placeholder="Password" required><br>
					<input type="password" name="confirm" placeholder="Re-enter Password" required><br>
					<button type="submit">Sign Up</button>
				</form>

				<h2>Sign In</h2>
				<form action="/signin" method="POST">
					<input name="email" placeholder="Email" required><br>
					<input type="password" name="password" placeholder="Password" required><br>
					<button type="submit">Sign In</button>
				</form>
			</body>
			</html>
		`))
	})

	http.HandleFunc("/signup", handlers.SignUp)
	http.HandleFunc("/signin", handlers.SignIn)

	http.ListenAndServe(":8080", nil)
}
