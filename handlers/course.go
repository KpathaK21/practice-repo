package handlers

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/KpathaK21/practice-repo/db"
	"github.com/KpathaK21/practice-repo/models"
)

// Middleware to check if user is a professor
func ProfessorOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In a real app, get the user from the session
		email := r.URL.Query().Get("email")
		if email == "" {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		var user models.User
		if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if !user.IsProfessor() {
			http.Error(w, "Unauthorized: Professors only", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// Middleware to check if user is a TA
func TAOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In a real app, get the user from the session
		email := r.URL.Query().Get("email")
		if email == "" {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		var user models.User
		if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if !user.IsTA() {
			http.Error(w, "Unauthorized: TAs only", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// Course creation handler
func CreateCourse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Render the create course form
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Get the user from the query parameter (in a real app, use session)
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Parse form data
	title := r.FormValue("title")
	description := r.FormValue("description")
	term := r.FormValue("term")
	syllabus := r.FormValue("syllabus")
	isPublicStr := r.FormValue("is_public")

	isPublic := isPublicStr == "on" || isPublicStr == "true"

	// Validate required fields
	if title == "" || term == "" {
		http.Error(w, "Title and term are required", http.StatusBadRequest)
		return
	}

	// Create the course
	course := models.Course{
		Title:       title,
		Description: description,
		Term:        term,
		Syllabus:    syllabus,
		IsPublic:    isPublic,
		ProfessorID: user.ID,
	}

	if err := db.DB.Create(&course).Error; err != nil {
		http.Error(w, "Error creating course: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to the course page
	http.Redirect(w, r, fmt.Sprintf("/course/%d?email=%s", course.ID, email), http.StatusSeeOther)
}

// Course listing handler
func ListCourses(w http.ResponseWriter, r *http.Request) {
	// Get the user from the query parameter (in a real app, use session)
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	var courses []models.Course

	// Different queries based on user role
	if user.IsProfessor() {
		// Professors see courses they teach
		db.DB.Where("professor_id = ?", user.ID).Find(&courses)
	} else if user.IsTA() {
		// TAs see courses they assist with
		db.DB.Joins("JOIN course_assistants ON course_assistants.course_id = courses.id").Where("course_assistants.user_id = ?", user.ID).Find(&courses)
	} else {
		// Students see courses they're enrolled in
		db.DB.Joins("JOIN user_courses ON user_courses.course_id = courses.id").Where("user_courses.user_id = ?", user.ID).Find(&courses)

		// Students can also see public courses they're not enrolled in
		var publicCourses []models.Course
		db.DB.Where("is_public = ? AND id NOT IN (SELECT course_id FROM user_courses WHERE user_id = ?)", true, user.ID).Find(&publicCourses)

		// Add a field to distinguish enrolled vs. public courses
		// In a real app, you might want to use a struct with additional fields
	}

	// Render the courses template
	tmpl := template.Must(template.ParseFiles("static/courses.html"))
	tmpl.Execute(w, struct {
		Courses  []models.Course
		Username string
		Role     string
	}{
		Courses:  courses,
		Username: user.Username,
		Role:     user.Role,
	})
}

// Course detail handler
func ViewCourse(w http.ResponseWriter, r *http.Request) {
	// Get the course ID from the URL
	courseIDStr := r.URL.Path[len("/course/"):]
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// Get the user from the query parameter (in a real app, use session)
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Get the course
	var course models.Course
	if err := db.DB.First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Check permissions
	hasAccess := false
	if user.IsProfessor() && course.ProfessorID == user.ID {
		hasAccess = true
	} else if user.IsTA() {
		// Check if user is a TA for this course
		var count int
		db.DB.Table("course_assistants").Where("course_id = ? AND user_id = ?", courseID, user.ID).Count(&count)
		hasAccess = count > 0
	} else {
		// Check if user is enrolled or course is public
		var count int
		db.DB.Table("user_courses").Where("course_id = ? AND user_id = ?", courseID, user.ID).Count(&count)
		hasAccess = count > 0 || course.IsPublic
	}

	if !hasAccess {
		http.Error(w, "You do not have access to this course", http.StatusForbidden)
		return
	}

	// Get course materials
	var materials []models.Material
	db.DB.Where("course_id = ?", courseID).Find(&materials)

	// Get course assignments
	var assignments []models.Assignment
	db.DB.Where("course_id = ?", courseID).Find(&assignments)

	// Get course announcements
	var announcements []models.Announcement
	db.DB.Where("course_id = ?", courseID).Order("created_at DESC").Find(&announcements)

	// Render the course template
	tmpl := template.Must(template.ParseFiles("static/course_detail.html"))
	tmpl.Execute(w, struct {
		Course        models.Course
		Materials     []models.Material
		Assignments   []models.Assignment
		Announcements []models.Announcement
		Username      string
		Role          string
		IsProfessor   bool
		IsTA          bool
	}{
		Course:        course,
		Materials:     materials,
		Assignments:   assignments,
		Announcements: announcements,
		Username:      user.Username,
		Role:          user.Role,
		IsProfessor:   user.IsProfessor() && course.ProfessorID == user.ID,
		IsTA:          user.IsTA() && hasAccess,
	})
}

// Export grades as CSV
func ExportGrades(w http.ResponseWriter, r *http.Request) {
	// Get the course ID from the URL
	courseIDStr := r.URL.Path[len("/course/") : len(r.URL.Path)-len("/export-grades")]
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// Get the user from the query parameter (in a real app, use session)
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Check if user is professor or TA for this course
	hasAccess := false
	if user.IsProfessor() {
		var course models.Course
		if err := db.DB.Where("id = ? AND professor_id = ?", courseID, user.ID).First(&course).Error; err == nil {
			hasAccess = true
		}
	} else if user.IsTA() {
		var count int
		db.DB.Table("course_assistants").Where("course_id = ? AND user_id = ?", courseID, user.ID).Count(&count)
		hasAccess = count > 0
	}

	if !hasAccess {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get the course
	var course models.Course
	if err := db.DB.First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Get all assignments for this course
	var assignments []models.Assignment
	db.DB.Where("course_id = ?", courseID).Find(&assignments)

	// Get all students enrolled in this course
	var students []models.User
	db.DB.Joins("JOIN user_courses ON user_courses.user_id = users.id").Where("user_courses.course_id = ? AND users.role = 'student'", courseID).Find(&students)

	// Set up CSV writer
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_grades.csv", course.Title))
	writer := csv.NewWriter(w)

	// Write header row
	header := []string{"Student ID", "Student Name", "Email"}
	for _, assignment := range assignments {
		header = append(header, assignment.Title)
	}
	header = append(header, "Total")
	writer.Write(header)

	// Write student rows
	for _, student := range students {
		row := []string{fmt.Sprintf("%d", student.ID), student.Username, student.Email}
		totalPoints := 0.0
		totalPossible := 0.0

		for _, assignment := range assignments {
			var submission models.Submission
			result := db.DB.Where("assignment_id = ? AND user_id = ?", assignment.ID, student.ID).First(&submission)

			if result.Error == nil {
				row = append(row, fmt.Sprintf("%.2f", submission.Grade))
				totalPoints += submission.Grade
			} else {
				row = append(row, "Not Submitted")
			}

			totalPossible += assignment.PointsValue
		}

		// Calculate total percentage
		percentage := 0.0
		if totalPossible > 0 {
			percentage = (totalPoints / totalPossible) * 100
		}
		row = append(row, fmt.Sprintf("%.2f%%", percentage))

		writer.Write(row)
	}

	writer.Flush()
}
