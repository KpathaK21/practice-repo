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
	http.HandleFunc("/refresh", handlers.RefreshToken) // New refresh token endpoint
	http.HandleFunc("/logout", handlers.Logout)       // New logout endpoint
	http.HandleFunc("/dashboard", handlers.JWTMiddleware(handlers.Dashboard))

	// Course routes with JWT middleware
	http.HandleFunc("/courses", handlers.JWTMiddleware(handlers.ListCourses))
	http.HandleFunc("/course/create", handlers.ProfessorJWTMiddleware(handlers.CreateCourse))
	http.HandleFunc("/course/update", handlers.ProfessorJWTMiddleware(handlers.UpdateCourse))
	http.HandleFunc("/course/assign-ta", handlers.ProfessorJWTMiddleware(handlers.AssignTA))
	http.HandleFunc("/course/remove-ta", handlers.ProfessorJWTMiddleware(handlers.RemoveTA))
	http.HandleFunc("/course/enroll-student", handlers.ProfessorJWTMiddleware(handlers.EnrollStudent))
	http.HandleFunc("/course/export-grades", handlers.CourseStaffJWTMiddleware(handlers.ExportGrades))

	// Content Management routes
	http.HandleFunc("/course/materials/create", handlers.CourseStaffJWTMiddleware(handlers.CreateMaterial))
	http.HandleFunc("/course/materials/update", handlers.CourseStaffJWTMiddleware(handlers.UpdateMaterial))
	http.HandleFunc("/course/materials/delete", handlers.CourseStaffJWTMiddleware(handlers.DeleteMaterial))

	// Assignment routes
	http.HandleFunc("/course/assignments/create", handlers.CourseStaffJWTMiddleware(handlers.CreateAssignment))
	http.HandleFunc("/course/assignments/update", handlers.CourseStaffJWTMiddleware(handlers.UpdateAssignment))
	http.HandleFunc("/course/assignments/delete", handlers.CourseStaffJWTMiddleware(handlers.DeleteAssignment))
	http.HandleFunc("/course/assignments/submit", handlers.EnrolledJWTMiddleware(handlers.SubmitAssignment))
	http.HandleFunc("/course/assignments/grade", handlers.CourseStaffJWTMiddleware(handlers.GradeSubmission))

	// Quiz routes
	http.HandleFunc("/course/quizzes/create", handlers.CourseStaffJWTMiddleware(handlers.CreateQuiz))
	http.HandleFunc("/course/quizzes/update", handlers.CourseStaffJWTMiddleware(handlers.UpdateQuiz))
	http.HandleFunc("/course/quizzes/delete", handlers.CourseStaffJWTMiddleware(handlers.DeleteQuiz))
	http.HandleFunc("/course/quizzes/take", handlers.EnrolledJWTMiddleware(handlers.TakeQuiz))
	http.HandleFunc("/course/quizzes/grade", handlers.CourseStaffJWTMiddleware(handlers.GradeQuiz))
	http.HandleFunc("/enroll-by-email", handlers.JWTMiddleware(handlers.EnrollByEmail))

	// Communication routes
	http.HandleFunc("/course/announcements/create", handlers.CourseStaffJWTMiddleware(handlers.CreateAnnouncement))
	http.HandleFunc("/course/discussions/create", handlers.EnrolledJWTMiddleware(handlers.CreateDiscussion))
	http.HandleFunc("/course/discussions/reply", handlers.EnrolledJWTMiddleware(handlers.ReplyToDiscussion))
	http.HandleFunc("/messages/send", handlers.JWTMiddleware(handlers.SendMessage))

	// Calendar routes
	http.HandleFunc("/course/events/create", handlers.CourseStaffJWTMiddleware(handlers.CreateEvent))
	http.HandleFunc("/calendar", handlers.JWTMiddleware(handlers.ViewCalendar))

	// This should be last as it's a catch-all
	http.HandleFunc("/course/", handlers.JWTMiddleware(handlers.ViewCourse))

	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
