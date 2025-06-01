package models

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// DBInterface provides database access methods
type DBInterface interface {
	Where(query interface{}, args ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Table(name string) *gorm.DB
	Count(value interface{}) *gorm.DB
}

// Current DB instance - to be set by the db package
var DB DBInterface

type User struct {
	gorm.Model
	Username                string `gorm:"unique;not null"`
	Email                   string `gorm:"unique;not null"`
	Password                string `gorm:"not null"`
	Role                    string `gorm:"not null;default:'student'"` // Role can be 'student', 'professor', or 'ta'
	VerificationCode        string
	VerificationCodeCreated int64 // Unix timestamp when the code was created
	IsVerified              bool
	// Relations
	EnrolledCourses  []Course `gorm:"many2many:user_courses;"`      // For students
	TeachingCourses  []Course `gorm:"foreignkey:ProfessorID"`       // For professors
	AssistingCourses []Course `gorm:"many2many:course_assistants;"` // For TAs
}

func (u *User) SetPassword(pw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(pw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pw))
	return err == nil
}

// Permission check methods
func (u *User) IsProfessor() bool {
	return u.Role == "professor"
}

func (u *User) IsTA() bool {
	return u.Role == "ta"
}

func (u *User) IsStudent() bool {
	return u.Role == "student"
}

// Check if user is professor of a specific course
func (u *User) IsProfessorOf(courseID uint) bool {
	if !u.IsProfessor() {
		return false
	}
	// Check if user is the professor of this course
	var course Course
	if err := DB.Where("id = ? AND professor_id = ?", courseID, u.ID).First(&course).Error; err != nil {
		return false
	}
	return true
}

// Check if user is TA of a specific course
func (u *User) IsTAOf(courseID uint) bool {
	if !u.IsTA() {
		return false
	}
	// Check if user is a TA for this course
	var count int
	DB.Table("course_assistants").Where("course_id = ? AND user_id = ?", courseID, u.ID).Count(&count)
	return count > 0
}

// Check if user is enrolled in a specific course
func (u *User) IsEnrolledIn(courseID uint) bool {
	if !u.IsStudent() {
		return false
	}
	// Check if user is enrolled in this course
	var count int
	DB.Table("user_courses").Where("course_id = ? AND user_id = ?", courseID, u.ID).Count(&count)
	return count > 0
}
