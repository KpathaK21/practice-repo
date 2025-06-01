package handlers

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// Middleware to check if user is a professor or TA for a specific course
func CourseStaffOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get course ID from URL or form
		courseIDStr := r.URL.Query().Get("course_id")
		if courseIDStr == "" {
			http.Error(w, "Course ID is required", http.StatusBadRequest)
			return
		}

		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Course ID", http.StatusBadRequest)
			return
		}

		// Get user from session
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

		// Check if user is professor or TA for this course
		if !user.IsProfessorOf(uint(courseID)) && !user.IsTAOf(uint(courseID)) {
			http.Error(w, "Unauthorized: You are not staff for this course", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// Middleware to check if user is enrolled in a specific course
func EnrolledOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get course ID from URL or form
		courseIDStr := r.URL.Query().Get("course_id")
		if courseIDStr == "" {
			http.Error(w, "Course ID is required", http.StatusBadRequest)
			return
		}

		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Course ID", http.StatusBadRequest)
			return
		}

		// Get user from session
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

		// Check if user is professor, TA, or enrolled student for this course
		if !user.IsProfessorOf(uint(courseID)) && !user.IsTAOf(uint(courseID)) && !user.IsEnrolledIn(uint(courseID)) {
			http.Error(w, "Unauthorized: You are not enrolled in this course", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// Course Management Handlers

// CreateCourse - For professors to create new courses
func CreateCourse(w http.ResponseWriter, r *http.Request) {
	// Get email from query parameter for both GET and POST requests
	email := r.URL.Query().Get("email")
	
	if r.Method != http.MethodPost {
		// Render the create course form with the email parameter
		template.Must(template.ParseFiles("static/create_course.html")).Execute(w, struct {
			Email string
		}{
			Email: email,
		})
		return
	}

	// Get user from session
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Parse form data
	r.ParseForm()
	course := models.Course{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Term:        r.FormValue("term"),
		Syllabus:    r.FormValue("syllabus"),
		Status:      models.CourseStatus(r.FormValue("status")),
		ProfessorID: user.ID,
	}

	// Save course to database
	if err := db.DB.Create(&course).Error; err != nil {
		http.Error(w, "Error creating course: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard instead of course page
	http.Redirect(w, r, fmt.Sprintf("/dashboard?email=%s", email), http.StatusSeeOther)
}

// UpdateCourse - For professors to update course details
func UpdateCourse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get course ID from URL
		courseIDStr := r.URL.Query().Get("id")
		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Course ID", http.StatusBadRequest)
			return
		}

		// Get course details
		var course models.Course
		if err := db.DB.First(&course, courseID).Error; err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}

		// Render the edit course form
		template.Must(template.ParseFiles("static/edit_course.html")).Execute(w, course)
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course ID from form
	courseIDStr := r.FormValue("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Check if user is the professor for this course
	if !user.IsProfessorOf(uint(courseID)) {
		http.Error(w, "Unauthorized: Only the course professor can update course details", http.StatusForbidden)
		return
	}

	// Get course from database
	var course models.Course
	if err := db.DB.First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Update course details
	r.ParseForm()
	course.Title = r.FormValue("title")
	course.Description = r.FormValue("description")
	course.Term = r.FormValue("term")
	course.Syllabus = r.FormValue("syllabus")
	course.Status = models.CourseStatus(r.FormValue("status"))

	// Save changes to database
	if err := db.DB.Save(&course).Error; err != nil {
		http.Error(w, "Error updating course: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard instead of course page
	http.Redirect(w, r, fmt.Sprintf("/dashboard?email=%s", email), http.StatusSeeOther)
}

// ListCourses - List all courses based on user role
func ListCourses(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	var courses []models.Course

	// Different queries based on user role
	switch {
	case user.IsProfessor():
		// Professors see courses they teach
		db.DB.Where("professor_id = ?", user.ID).Find(&courses)
	case user.IsTA():
		// TAs see courses they assist with
		db.DB.Joins("JOIN course_assistants ON course_assistants.course_id = courses.id").Where("course_assistants.user_id = ?", user.ID).Find(&courses)
	default: // Students
		// Students see published courses they're enrolled in and public courses
		db.DB.Where("status = ? AND is_public = ?", models.Published, true).Or(
			"id IN (SELECT course_id FROM user_courses WHERE user_id = ?)", user.ID).Find(&courses)
	}

	// Render courses list template
	template.Must(template.ParseFiles("static/courses.html")).Execute(w, map[string]interface{}{
		"Courses": courses,
		"User":    user,
	})
}

// ViewCourse - View a specific course
func ViewCourse(w http.ResponseWriter, r *http.Request) {
	// Get course ID from URL
	courseIDStr := r.URL.Query().Get("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course details
	var course models.Course
	if err := db.DB.Preload("Professor").Preload("Assistants").Preload("Students").Preload("Materials").Preload("Assignments").Preload("Announcements").First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Check if user has access to this course
	hasAccess := user.IsProfessorOf(uint(courseID)) || user.IsTAOf(uint(courseID)) || user.IsEnrolledIn(uint(courseID)) || (course.IsPublic && course.Status == models.Published)
	if !hasAccess {
		http.Error(w, "Unauthorized: You do not have access to this course", http.StatusForbidden)
		return
	}

	// Render course view template
	template.Must(template.ParseFiles("static/course.html")).Execute(w, map[string]interface{}{
		"Course":      course,
		"User":        user,
		"IsProfessor": user.IsProfessorOf(uint(courseID)),
		"IsTA":        user.IsTAOf(uint(courseID)),
		"IsStudent":   user.IsEnrolledIn(uint(courseID)),
	})
}

// AssignTA - For professors to assign TAs to their courses
func AssignTA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get course ID from URL
		courseIDStr := r.URL.Query().Get("course_id")
		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Course ID", http.StatusBadRequest)
			return
		}

		// Get course details
		var course models.Course
		if err := db.DB.First(&course, courseID).Error; err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}

		// Get all users with TA role
		var tas []models.User
		db.DB.Where("role = ?", "ta").Find(&tas)

		// Render the assign TA form
		template.Must(template.ParseFiles("static/assign_ta.html")).Execute(w, map[string]interface{}{
			"Course": course,
			"TAs":    tas,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course ID from form
	courseIDStr := r.FormValue("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Check if user is the professor for this course
	if !user.IsProfessorOf(uint(courseID)) {
		http.Error(w, "Unauthorized: Only the course professor can assign TAs", http.StatusForbidden)
		return
	}

	// Get course from database
	var course models.Course
	if err := db.DB.First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Get TA ID from form
	taIDStr := r.FormValue("ta_id")
	taID, err := strconv.ParseUint(taIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid TA ID", http.StatusBadRequest)
		return
	}

	// Get TA from database
	var ta models.User
	if err := db.DB.First(&ta, taID).Error; err != nil {
		http.Error(w, "TA not found", http.StatusNotFound)
		return
	}

	// Check if TA has the TA role
	if !ta.IsTA() {
		http.Error(w, "Selected user is not a TA", http.StatusBadRequest)
		return
	}

	// Add TA to course assistants
	if err := db.DB.Model(&course).Association("Assistants").Append(&ta).Error; err != nil {
		http.Error(w, "Error assigning TA: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard instead of course page
	http.Redirect(w, r, fmt.Sprintf("/dashboard?email=%s", email), http.StatusSeeOther)
}

// RemoveTA - For professors to remove TAs from their courses
func RemoveTA(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course ID from URL
	courseIDStr := r.URL.Query().Get("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Check if user is the professor for this course
	if !user.IsProfessorOf(uint(courseID)) {
		http.Error(w, "Unauthorized: Only the course professor can remove TAs", http.StatusForbidden)
		return
	}

	// Get course from database
	var course models.Course
	if err := db.DB.First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Get TA ID from URL
	taIDStr := r.URL.Query().Get("ta_id")
	taID, err := strconv.ParseUint(taIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid TA ID", http.StatusBadRequest)
		return
	}

	// Get TA from database
	var ta models.User
	if err := db.DB.First(&ta, taID).Error; err != nil {
		http.Error(w, "TA not found", http.StatusNotFound)
		return
	}

	// Remove TA from course assistants
	if err := db.DB.Model(&course).Association("Assistants").Delete(&ta).Error; err != nil {
		http.Error(w, "Error removing TA: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard instead of course page
	http.Redirect(w, r, fmt.Sprintf("/dashboard?email=%s", email), http.StatusSeeOther)
}

// EnrollStudent - For professors to enroll students in their courses
func EnrollStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get course ID from URL
		courseIDStr := r.URL.Query().Get("course_id")
		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Course ID", http.StatusBadRequest)
			return
		}

		// Get course details
		var course models.Course
		if err := db.DB.First(&course, courseID).Error; err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}

		// Get all users with student role
		var students []models.User
		db.DB.Where("role = ?", "student").Find(&students)

		// Render the enroll student form
		template.Must(template.ParseFiles("static/enroll_student.html")).Execute(w, map[string]interface{}{
			"Course":   course,
			"Students": students,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course ID from form
	courseIDStr := r.FormValue("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Check if user is the professor for this course
	if !user.IsProfessorOf(uint(courseID)) {
		http.Error(w, "Unauthorized: Only the course professor can enroll students", http.StatusForbidden)
		return
	}

	// Get course from database
	var course models.Course
	if err := db.DB.First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Get student ID from form
	studentIDStr := r.FormValue("student_id")
	studentID, err := strconv.ParseUint(studentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		return
	}

	// Get student from database
	var student models.User
	if err := db.DB.First(&student, studentID).Error; err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	// Check if student has the student role
	if !student.IsStudent() {
		http.Error(w, "Selected user is not a student", http.StatusBadRequest)
		return
	}

	// Add student to course students
	if err := db.DB.Model(&course).Association("Students").Append(&student).Error; err != nil {
		http.Error(w, "Error enrolling student: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard instead of course page
	http.Redirect(w, r, fmt.Sprintf("/dashboard?email=%s", email), http.StatusSeeOther)
}

// EnrollByEmail - For professors to enroll students using their email addresses
func EnrollByEmail(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course ID from form
	courseIDStr := r.FormValue("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Check if user is the professor for this course
	if !user.IsProfessorOf(uint(courseID)) {
		http.Error(w, "Unauthorized: Only the course professor can enroll students", http.StatusForbidden)
		return
	}

	// Get course from database
	var course models.Course
	if err := db.DB.First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Process single email
	studentEmail := r.FormValue("student_email")
	if studentEmail != "" {
		// Find student by email
		var student models.User
		if err := db.DB.Where("email = ?", studentEmail).First(&student).Error; err != nil {
			http.Error(w, "Student with email "+studentEmail+" not found", http.StatusNotFound)
			return
		}

		// Check if student has the student role
		if !student.IsStudent() {
			http.Error(w, "User with email "+studentEmail+" is not a student", http.StatusBadRequest)
			return
		}

		// Add student to course students
		if err := db.DB.Model(&course).Association("Students").Append(&student).Error; err != nil {
			http.Error(w, "Error enrolling student: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Process multiple emails
	multipleEmails := r.FormValue("multiple_emails")
	if multipleEmails != "" {
		// Split the text area content by newline
		emails := strings.Split(multipleEmails, "\n")

		successCount := 0
		failures := []string{}

		for _, studentEmail := range emails {
			// Trim whitespace
			studentEmail = strings.TrimSpace(studentEmail)
			if studentEmail == "" {
				continue
			}

			// Find student by email
			var student models.User
			if err := db.DB.Where("email = ?", studentEmail).First(&student).Error; err != nil {
				failures = append(failures, studentEmail+" (not found)")
				continue
			}

			// Check if student has the student role
			if !student.IsStudent() {
				failures = append(failures, studentEmail+" (not a student)")
				continue
			}

			// Check if student is already enrolled
			if student.IsEnrolledIn(uint(courseID)) {
				failures = append(failures, studentEmail+" (already enrolled)")
				continue
			}

			// Add student to course students
			if err := db.DB.Model(&course).Association("Students").Append(&student).Error; err != nil {
				failures = append(failures, studentEmail+" (database error)")
				continue
			}

			successCount++
		}

		// If there were any failures, show them
		if len(failures) > 0 {
			message := fmt.Sprintf("Successfully enrolled %d students. Failed to enroll the following emails:\n%s",
				successCount, strings.Join(failures, "\n"))
			http.Error(w, message, http.StatusPartialContent)
			return
		}
	}

	// Redirect to dashboard instead of course page
	http.Redirect(w, r, fmt.Sprintf("/dashboard?email=%s", email), http.StatusSeeOther)
}

// ExportGrades - For professors to export grades as CSV
func ExportGrades(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course ID from URL
	courseIDStr := r.URL.Query().Get("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Check if user is the professor or TA for this course
	if !user.IsProfessorOf(uint(courseID)) && !user.IsTAOf(uint(courseID)) {
		http.Error(w, "Unauthorized: Only the course professor or TAs can export grades", http.StatusForbidden)
		return
	}

	// Get course from database
	var course models.Course
	if err := db.DB.Preload("Students").Preload("Assignments").First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Set response headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=grades-%s.csv", course.Title))

	// Create CSV writer
	writer := csv.NewWriter(w)

	// Write header row
	header := []string{"Student ID", "Student Name", "Email"}
	for _, assignment := range course.Assignments {
		header = append(header, assignment.Title)
	}
	header = append(header, "Total")
	writer.Write(header)

	// Write student rows
	for _, student := range course.Students {
		row := []string{fmt.Sprintf("%d", student.ID), student.Username, student.Email}
		total := 0.0

		// Get grades for each assignment
		for _, assignment := range course.Assignments {
			var submission models.Submission
			result := db.DB.Where("assignment_id = ? AND user_id = ?", assignment.ID, student.ID).First(&submission)
			if result.Error == nil {
				row = append(row, fmt.Sprintf("%.2f", submission.Grade))
				total += submission.Grade
			} else {
				row = append(row, "N/A")
			}
		}

		// Add total
		row = append(row, fmt.Sprintf("%.2f", total))
		writer.Write(row)
	}

	writer.Flush()
}

// Content Management Handlers

// CreateMaterial - For course staff to create new course materials
func CreateMaterial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get course ID from URL
		courseIDStr := r.URL.Query().Get("course_id")
		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Course ID", http.StatusBadRequest)
			return
		}

		// Get course details
		var course models.Course
		if err := db.DB.First(&course, courseID).Error; err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}

		// Render the create material form
		template.Must(template.ParseFiles("static/create_material.html")).Execute(w, map[string]interface{}{
			"Course": course,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course ID from form
	courseIDStr := r.FormValue("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(uint(courseID)) && !user.IsTAOf(uint(courseID)) {
		http.Error(w, "Unauthorized: Only course staff can create materials", http.StatusForbidden)
		return
	}

	// Parse form data
	r.ParseMultipartForm(10 << 20) // 10 MB max memory

	// Create new material
	material := models.Material{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		FileType:    r.FormValue("file_type"),
		CourseID:    uint(courseID),
	}

	// Handle file upload if present
	// Note: In a real implementation, you would save the file and set FilePath
	// material.FilePath = savedFilePath

	// Parse visible dates if provided
	if visibleFrom := r.FormValue("visible_from"); visibleFrom != "" {
		if date, err := time.Parse("2006-01-02", visibleFrom); err == nil {
			material.VisibleFrom = date
		}
	}

	if visibleTo := r.FormValue("visible_to"); visibleTo != "" {
		if date, err := time.Parse("2006-01-02", visibleTo); err == nil {
			material.VisibleTo = date
		}
	}

	// Save material to database
	if err := db.DB.Create(&material).Error; err != nil {
		http.Error(w, "Error creating material: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", courseID, email), http.StatusSeeOther)
}

// UpdateMaterial - For course staff to update existing materials
func UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get material ID from URL
		materialIDStr := r.URL.Query().Get("id")
		materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Material ID", http.StatusBadRequest)
			return
		}

		// Get material details
		var material models.Material
		if err := db.DB.First(&material, materialID).Error; err != nil {
			http.Error(w, "Material not found", http.StatusNotFound)
			return
		}

		// Render the update material form
		template.Must(template.ParseFiles("static/update_material.html")).Execute(w, map[string]interface{}{
			"Material": material,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get material ID from form
	materialIDStr := r.FormValue("id")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Material ID", http.StatusBadRequest)
		return
	}

	// Get material from database
	var material models.Material
	if err := db.DB.First(&material, materialID).Error; err != nil {
		http.Error(w, "Material not found", http.StatusNotFound)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(material.CourseID) && !user.IsTAOf(material.CourseID) {
		http.Error(w, "Unauthorized: Only course staff can update materials", http.StatusForbidden)
		return
	}

	// Update material details
	r.ParseMultipartForm(10 << 20) // 10 MB max memory
	material.Title = r.FormValue("title")
	material.Description = r.FormValue("description")

	// Handle file upload if present
	// Note: In a real implementation, you would update the file and FilePath

	// Parse visible dates if provided
	if visibleFrom := r.FormValue("visible_from"); visibleFrom != "" {
		if date, err := time.Parse("2006-01-02", visibleFrom); err == nil {
			material.VisibleFrom = date
		}
	}

	if visibleTo := r.FormValue("visible_to"); visibleTo != "" {
		if date, err := time.Parse("2006-01-02", visibleTo); err == nil {
			material.VisibleTo = date
		}
	}

	// Save changes to database
	if err := db.DB.Save(&material).Error; err != nil {
		http.Error(w, "Error updating material: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", material.CourseID, email), http.StatusSeeOther)
}

// DeleteMaterial - For course staff to delete materials
func DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get material ID from URL
	materialIDStr := r.URL.Query().Get("id")
	materialID, err := strconv.ParseUint(materialIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Material ID", http.StatusBadRequest)
		return
	}

	// Get material from database
	var material models.Material
	if err := db.DB.First(&material, materialID).Error; err != nil {
		http.Error(w, "Material not found", http.StatusNotFound)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(material.CourseID) && !user.IsTAOf(material.CourseID) {
		http.Error(w, "Unauthorized: Only course staff can delete materials", http.StatusForbidden)
		return
	}

	// Delete material from database
	if err := db.DB.Delete(&material).Error; err != nil {
		http.Error(w, "Error deleting material: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", material.CourseID, email), http.StatusSeeOther)
}

// Assignment Handlers

// CreateAssignment - For course staff to create new assignments
func CreateAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get course ID from URL
		courseIDStr := r.URL.Query().Get("course_id")
		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Course ID", http.StatusBadRequest)
			return
		}

		// Get course details
		var course models.Course
		if err := db.DB.First(&course, courseID).Error; err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}

		// Render the create assignment form
		template.Must(template.ParseFiles("static/create_assignment.html")).Execute(w, map[string]interface{}{
			"Course": course,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course ID from form
	courseIDStr := r.FormValue("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(uint(courseID)) && !user.IsTAOf(uint(courseID)) {
		http.Error(w, "Unauthorized: Only course staff can create assignments", http.StatusForbidden)
		return
	}

	// Parse form data
	r.ParseForm()

	// Parse points value
	pointsValue, err := strconv.ParseFloat(r.FormValue("points_value"), 64)
	if err != nil {
		pointsValue = 0
	}

	// Parse late penalty
	latePenalty, err := strconv.ParseFloat(r.FormValue("late_penalty"), 64)
	if err != nil {
		latePenalty = 0
	}

	// Parse due date
	dueDate, err := time.Parse("2006-01-02T15:04", r.FormValue("due_date"))
	if err != nil {
		dueDate = time.Now().Add(7 * 24 * time.Hour) // Default to 1 week from now
	}

	// Create new assignment
	assignment := models.Assignment{
		Title:          r.FormValue("title"),
		Description:    r.FormValue("description"),
		DueDate:        dueDate,
		PointsValue:    pointsValue,
		SubmissionType: r.FormValue("submission_type"),
		AllowLate:      r.FormValue("allow_late") == "on",
		LatePenalty:    latePenalty,
		CourseID:       uint(courseID),
	}

	// Parse visible dates if provided
	if visibleFrom := r.FormValue("visible_from"); visibleFrom != "" {
		if date, err := time.Parse("2006-01-02", visibleFrom); err == nil {
			assignment.VisibleFrom = date
		}
	}

	if visibleTo := r.FormValue("visible_to"); visibleTo != "" {
		if date, err := time.Parse("2006-01-02", visibleTo); err == nil {
			assignment.VisibleTo = date
		}
	}

	// Save assignment to database
	if err := db.DB.Create(&assignment).Error; err != nil {
		http.Error(w, "Error creating assignment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", courseID, email), http.StatusSeeOther)
}

// UpdateAssignment - For course staff to update existing assignments
func UpdateAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get assignment ID from URL
		assignmentIDStr := r.URL.Query().Get("id")
		assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Assignment ID", http.StatusBadRequest)
			return
		}

		// Get assignment details
		var assignment models.Assignment
		if err := db.DB.First(&assignment, assignmentID).Error; err != nil {
			http.Error(w, "Assignment not found", http.StatusNotFound)
			return
		}

		// Render the update assignment form
		template.Must(template.ParseFiles("static/update_assignment.html")).Execute(w, map[string]interface{}{
			"Assignment": assignment,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get assignment ID from form
	assignmentIDStr := r.FormValue("id")
	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Assignment ID", http.StatusBadRequest)
		return
	}

	// Get assignment from database
	var assignment models.Assignment
	if err := db.DB.First(&assignment, assignmentID).Error; err != nil {
		http.Error(w, "Assignment not found", http.StatusNotFound)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(assignment.CourseID) && !user.IsTAOf(assignment.CourseID) {
		http.Error(w, "Unauthorized: Only course staff can update assignments", http.StatusForbidden)
		return
	}

	// Parse form data
	r.ParseForm()

	// Update assignment details
	assignment.Title = r.FormValue("title")
	assignment.Description = r.FormValue("description")
	assignment.SubmissionType = r.FormValue("submission_type")
	assignment.AllowLate = r.FormValue("allow_late") == "on"

	// Parse points value
	if pointsValue, err := strconv.ParseFloat(r.FormValue("points_value"), 64); err == nil {
		assignment.PointsValue = pointsValue
	}

	// Parse late penalty
	if latePenalty, err := strconv.ParseFloat(r.FormValue("late_penalty"), 64); err == nil {
		assignment.LatePenalty = latePenalty
	}

	// Parse due date
	if dueDate, err := time.Parse("2006-01-02T15:04", r.FormValue("due_date")); err == nil {
		assignment.DueDate = dueDate
	}

	// Parse visible dates if provided
	if visibleFrom := r.FormValue("visible_from"); visibleFrom != "" {
		if date, err := time.Parse("2006-01-02", visibleFrom); err == nil {
			assignment.VisibleFrom = date
		}
	}

	if visibleTo := r.FormValue("visible_to"); visibleTo != "" {
		if date, err := time.Parse("2006-01-02", visibleTo); err == nil {
			assignment.VisibleTo = date
		}
	}

	// Save changes to database
	if err := db.DB.Save(&assignment).Error; err != nil {
		http.Error(w, "Error updating assignment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", assignment.CourseID, email), http.StatusSeeOther)
}

// DeleteAssignment - For course staff to delete assignments
func DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get assignment ID from URL
	assignmentIDStr := r.URL.Query().Get("id")
	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Assignment ID", http.StatusBadRequest)
		return
	}

	// Get assignment from database
	var assignment models.Assignment
	if err := db.DB.First(&assignment, assignmentID).Error; err != nil {
		http.Error(w, "Assignment not found", http.StatusNotFound)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(assignment.CourseID) && !user.IsTAOf(assignment.CourseID) {
		http.Error(w, "Unauthorized: Only course staff can delete assignments", http.StatusForbidden)
		return
	}

	// Delete assignment from database
	if err := db.DB.Delete(&assignment).Error; err != nil {
		http.Error(w, "Error deleting assignment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", assignment.CourseID, email), http.StatusSeeOther)
}

// SubmitAssignment - For students to submit assignments
func SubmitAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get assignment ID from URL
		assignmentIDStr := r.URL.Query().Get("id")
		assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Assignment ID", http.StatusBadRequest)
			return
		}

		// Get assignment details
		var assignment models.Assignment
		if err := db.DB.First(&assignment, assignmentID).Error; err != nil {
			http.Error(w, "Assignment not found", http.StatusNotFound)
			return
		}

		// Render the submit assignment form
		template.Must(template.ParseFiles("static/submit_assignment.html")).Execute(w, map[string]interface{}{
			"Assignment": assignment,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get assignment ID from form
	assignmentIDStr := r.FormValue("assignment_id")
	assignmentID, err := strconv.ParseUint(assignmentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Assignment ID", http.StatusBadRequest)
		return
	}

	// Get assignment from database
	var assignment models.Assignment
	if err := db.DB.First(&assignment, assignmentID).Error; err != nil {
		http.Error(w, "Assignment not found", http.StatusNotFound)
		return
	}

	// Check if user is enrolled in this course
	if !user.IsEnrolledIn(assignment.CourseID) {
		http.Error(w, "Unauthorized: You are not enrolled in this course", http.StatusForbidden)
		return
	}

	// Check if assignment is past due date and late submissions are not allowed
	now := time.Now()
	isLate := now.After(assignment.DueDate)
	if isLate && !assignment.AllowLate {
		http.Error(w, "This assignment is past due and does not accept late submissions", http.StatusBadRequest)
		return
	}

	// Parse form data
	r.ParseMultipartForm(10 << 20) // 10 MB max memory

	// Create new submission
	submission := models.Submission{
		AssignmentID: assignment.ID,
		UserID:       user.ID,
		Content:      r.FormValue("content"),
		SubmittedAt:  now,
		IsLate:       isLate,
	}

	// Handle file upload if present
	// Note: In a real implementation, you would save the file and set FilePath
	// submission.FilePath = savedFilePath

	// Save submission to database
	if err := db.DB.Create(&submission).Error; err != nil {
		http.Error(w, "Error submitting assignment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", assignment.CourseID, email), http.StatusSeeOther)
}

// GradeSubmission - For course staff to grade assignment submissions
func GradeSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get submission ID from URL
		submissionIDStr := r.URL.Query().Get("id")
		submissionID, err := strconv.ParseUint(submissionIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Submission ID", http.StatusBadRequest)
			return
		}

		// Get submission details
		var submission models.Submission
		if err := db.DB.Preload("Assignment").Preload("User").First(&submission, submissionID).Error; err != nil {
			http.Error(w, "Submission not found", http.StatusNotFound)
			return
		}

		// Render the grade submission form
		template.Must(template.ParseFiles("static/grade_submission.html")).Execute(w, map[string]interface{}{
			"Submission": submission,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get submission ID from form
	submissionIDStr := r.FormValue("submission_id")
	submissionID, err := strconv.ParseUint(submissionIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Submission ID", http.StatusBadRequest)
		return
	}

	// Get submission from database
	var submission models.Submission
	if err := db.DB.First(&submission, submissionID).Error; err != nil {
		http.Error(w, "Submission not found", http.StatusNotFound)
		return
	}

	// Then load the assignment separately
	var assignment models.Assignment
	if err := db.DB.First(&assignment, submission.AssignmentID).Error; err != nil {
		http.Error(w, "Assignment not found", http.StatusNotFound)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(assignment.CourseID) && !user.IsTAOf(assignment.CourseID) {
		http.Error(w, "Unauthorized: Only course staff can grade submissions", http.StatusForbidden)
		return
	}

	// Parse form data
	r.ParseForm()

	// Parse grade
	grade, err := strconv.ParseFloat(r.FormValue("grade"), 64)
	if err != nil {
		http.Error(w, "Invalid grade value", http.StatusBadRequest)
		return
	}

	// Update submission with grade and feedback
	submission.Grade = grade
	submission.Feedback = r.FormValue("feedback")
	submission.GradedBy = user.ID
	submission.GradedAt = time.Now()

	// Save changes to database
	if err := db.DB.Save(&submission).Error; err != nil {
		http.Error(w, "Error grading submission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", assignment.CourseID, email), http.StatusSeeOther)
}

// Quiz Handlers

// CreateQuiz - For course staff to create new quizzes
func CreateQuiz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get course ID from URL
		courseIDStr := r.URL.Query().Get("course_id")
		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Course ID", http.StatusBadRequest)
			return
		}

		// Get course details
		var course models.Course
		if err := db.DB.First(&course, courseID).Error; err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}

		// Render the create quiz form
		template.Must(template.ParseFiles("static/create_quiz.html")).Execute(w, map[string]interface{}{
			"Course": course,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get course ID from form
	courseIDStr := r.FormValue("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Course ID", http.StatusBadRequest)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(uint(courseID)) && !user.IsTAOf(uint(courseID)) {
		http.Error(w, "Unauthorized: Only course staff can create quizzes", http.StatusForbidden)
		return
	}

	// Parse form data
	r.ParseForm()

	// Parse points value
	pointsValue, err := strconv.ParseFloat(r.FormValue("points_value"), 64)
	if err != nil {
		pointsValue = 0
	}

	// Parse time limit
	timeLimit, err := strconv.Atoi(r.FormValue("time_limit"))
	if err != nil {
		timeLimit = 60 // Default to 60 minutes
	}

	// Parse attempts
	attempts, err := strconv.Atoi(r.FormValue("attempts"))
	if err != nil {
		attempts = 1 // Default to 1 attempt
	}

	// Parse due date
	dueDate, err := time.Parse("2006-01-02T15:04", r.FormValue("due_date"))
	if err != nil {
		dueDate = time.Now().Add(7 * 24 * time.Hour) // Default to 1 week from now
	}

	// Create new quiz
	quiz := models.Quiz{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		DueDate:     dueDate,
		TimeLimit:   timeLimit,
		Attempts:    attempts,
		PointsValue: pointsValue,
		CourseID:    uint(courseID),
	}

	// Parse visible dates if provided
	if visibleFrom := r.FormValue("visible_from"); visibleFrom != "" {
		if date, err := time.Parse("2006-01-02", visibleFrom); err == nil {
			quiz.VisibleFrom = date
		}
	}

	if visibleTo := r.FormValue("visible_to"); visibleTo != "" {
		if date, err := time.Parse("2006-01-02", visibleTo); err == nil {
			quiz.VisibleTo = date
		}
	}

	// Save quiz to database
	if err := db.DB.Create(&quiz).Error; err != nil {
		http.Error(w, "Error creating quiz: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to add questions page
	http.Redirect(w, r, fmt.Sprintf("/course/quizzes/add-questions?id=%d&email=%s", quiz.ID, email), http.StatusSeeOther)
}

// UpdateQuiz - For course staff to update existing quizzes
func UpdateQuiz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Get quiz ID from URL
		quizIDStr := r.URL.Query().Get("id")
		quizID, err := strconv.ParseUint(quizIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Quiz ID", http.StatusBadRequest)
			return
		}

		// Get quiz details
		var quiz models.Quiz
		if err := db.DB.First(&quiz, quizID).Error; err != nil {
			http.Error(w, "Quiz not found", http.StatusNotFound)
			return
		}

		// Render the update quiz form
		template.Must(template.ParseFiles("static/update_quiz.html")).Execute(w, map[string]interface{}{
			"Quiz": quiz,
		})
		return
	}

	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get quiz ID from form
	quizIDStr := r.FormValue("id")
	quizID, err := strconv.ParseUint(quizIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Quiz ID", http.StatusBadRequest)
		return
	}

	// Get quiz from database
	var quiz models.Quiz
	if err := db.DB.First(&quiz, quizID).Error; err != nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(quiz.CourseID) && !user.IsTAOf(quiz.CourseID) {
		http.Error(w, "Unauthorized: Only course staff can update quizzes", http.StatusForbidden)
		return
	}

	// Parse form data
	r.ParseForm()

	// Update quiz details
	quiz.Title = r.FormValue("title")
	quiz.Description = r.FormValue("description")

	// Parse points value
	if pointsValue, err := strconv.ParseFloat(r.FormValue("points_value"), 64); err == nil {
		quiz.PointsValue = pointsValue
	}

	// Parse time limit
	if timeLimit, err := strconv.Atoi(r.FormValue("time_limit")); err == nil {
		quiz.TimeLimit = timeLimit
	}

	// Parse attempts
	if attempts, err := strconv.Atoi(r.FormValue("attempts")); err == nil {
		quiz.Attempts = attempts
	}

	// Parse due date
	if dueDate, err := time.Parse("2006-01-02T15:04", r.FormValue("due_date")); err == nil {
		quiz.DueDate = dueDate
	}

	// Parse visible dates if provided
	if visibleFrom := r.FormValue("visible_from"); visibleFrom != "" {
		if date, err := time.Parse("2006-01-02", visibleFrom); err == nil {
			quiz.VisibleFrom = date
		}
	}

	if visibleTo := r.FormValue("visible_to"); visibleTo != "" {
		if date, err := time.Parse("2006-01-02", visibleTo); err == nil {
			quiz.VisibleTo = date
		}
	}

	// Save changes to database
	if err := db.DB.Save(&quiz).Error; err != nil {
		http.Error(w, "Error updating quiz: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", quiz.CourseID, email), http.StatusSeeOther)
}

// DeleteQuiz - For course staff to delete quizzes
func DeleteQuiz(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	email := r.URL.Query().Get("email")
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Get quiz ID from URL
	quizIDStr := r.URL.Query().Get("id")
	quizID, err := strconv.ParseUint(quizIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Quiz ID", http.StatusBadRequest)
		return
	}

	// Get quiz from database
	var quiz models.Quiz
	if err := db.DB.First(&quiz, quizID).Error; err != nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	// Check if user is professor or TA for this course
	if !user.IsProfessorOf(quiz.CourseID) && !user.IsTAOf(quiz.CourseID) {
		http.Error(w, "Unauthorized: Only course staff can delete quizzes", http.StatusForbidden)
		return
	}

	// Delete quiz from database
	if err := db.DB.Delete(&quiz).Error; err != nil {
		http.Error(w, "Error deleting quiz: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, fmt.Sprintf("/course?id=%d&email=%s", quiz.CourseID, email), http.StatusSeeOther)
}

// TakeQuiz - For students to take quizzes
func TakeQuiz(w http.ResponseWriter, r *http.Request) {
	// Implementation for students to take quizzes
	// This would involve displaying quiz questions, handling answers, and saving submissions
}

// GradeQuiz - For course staff to grade quiz submissions
func GradeQuiz(w http.ResponseWriter, r *http.Request) {
	// Implementation for course staff to grade quiz submissions
	// This would involve reviewing student answers and providing feedback
}

// Communication Handlers

// CreateAnnouncement - For course staff to create announcements
func CreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	// Implementation for creating course announcements
}

// CreateDiscussion - For users to create discussion threads
func CreateDiscussion(w http.ResponseWriter, r *http.Request) {
	// Implementation for creating discussion threads
}

// ReplyToDiscussion - For users to reply to discussion threads
func ReplyToDiscussion(w http.ResponseWriter, r *http.Request) {
	// Implementation for replying to discussion threads
}

// SendMessage - For users to send direct messages
func SendMessage(w http.ResponseWriter, r *http.Request) {
	// Implementation for sending direct messages
}

// Calendar Handlers

// CreateEvent - For course staff to create calendar events
func CreateEvent(w http.ResponseWriter, r *http.Request) {
	// Implementation for creating calendar events
}

// ViewCalendar - For users to view their calendar
func ViewCalendar(w http.ResponseWriter, r *http.Request) {
	// Implementation for viewing calendar events
}
