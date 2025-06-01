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
	http.HandleFunc("/course/update", handlers.ProfessorOnly(handlers.UpdateCourse))
	http.HandleFunc("/course/assign-ta", handlers.ProfessorOnly(handlers.AssignTA))
	http.HandleFunc("/course/remove-ta", handlers.ProfessorOnly(handlers.RemoveTA))
	http.HandleFunc("/course/enroll-student", handlers.ProfessorOnly(handlers.EnrollStudent))
	http.HandleFunc("/course/export-grades", handlers.CourseStaffOnly(handlers.ExportGrades))

	// Content Management routes
	http.HandleFunc("/course/materials/create", handlers.CourseStaffOnly(handlers.CreateMaterial))
	http.HandleFunc("/course/materials/update", handlers.CourseStaffOnly(handlers.UpdateMaterial))
	http.HandleFunc("/course/materials/delete", handlers.CourseStaffOnly(handlers.DeleteMaterial))

	// Assignment routes
	http.HandleFunc("/course/assignments/create", handlers.CourseStaffOnly(handlers.CreateAssignment))
	http.HandleFunc("/course/assignments/update", handlers.CourseStaffOnly(handlers.UpdateAssignment))
	http.HandleFunc("/course/assignments/delete", handlers.CourseStaffOnly(handlers.DeleteAssignment))
	http.HandleFunc("/course/assignments/submit", handlers.EnrolledOnly(handlers.SubmitAssignment))
	http.HandleFunc("/course/assignments/grade", handlers.CourseStaffOnly(handlers.GradeSubmission))

	// Quiz routes
	http.HandleFunc("/course/quizzes/create", handlers.CourseStaffOnly(handlers.CreateQuiz))
	http.HandleFunc("/course/quizzes/update", handlers.CourseStaffOnly(handlers.UpdateQuiz))
	http.HandleFunc("/course/quizzes/delete", handlers.CourseStaffOnly(handlers.DeleteQuiz))
	http.HandleFunc("/course/quizzes/take", handlers.EnrolledOnly(handlers.TakeQuiz))
	http.HandleFunc("/course/quizzes/grade", handlers.CourseStaffOnly(handlers.GradeQuiz))
	http.HandleFunc("/enroll-by-email", handlers.EnrollByEmail)

	// Communication routes
	http.HandleFunc("/course/announcements/create", handlers.CourseStaffOnly(handlers.CreateAnnouncement))
	http.HandleFunc("/course/discussions/create", handlers.EnrolledOnly(handlers.CreateDiscussion))
	http.HandleFunc("/course/discussions/reply", handlers.EnrolledOnly(handlers.ReplyToDiscussion))
	http.HandleFunc("/messages/send", handlers.SendMessage)

	// Calendar routes
	http.HandleFunc("/course/events/create", handlers.CourseStaffOnly(handlers.CreateEvent))
	http.HandleFunc("/calendar", handlers.ViewCalendar)

	// This should be last as it's a catch-all
	http.HandleFunc("/course/", handlers.ViewCourse)

	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
