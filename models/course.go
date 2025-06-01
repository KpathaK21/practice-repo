package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Course struct {
	gorm.Model
	Title         string `gorm:"not null"`
	Description   string `gorm:"type:text"`
	Term          string `gorm:"not null"` // e.g., "Fall 2023"
	Syllabus      string `gorm:"type:text"`
	IsPublic      bool   `gorm:"default:false"`
	ProfessorID   uint
	Professor     User   `gorm:"foreignkey:ProfessorID"`
	Students      []User `gorm:"many2many:user_courses;"`
	Assistants    []User `gorm:"many2many:course_assistants;"`
	Materials     []Material
	Assignments   []Assignment
	Announcements []Announcement
}

type Material struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description string
	FileType    string // "pdf", "video", "link", etc.
	FilePath    string
	CourseID    uint
}

type Assignment struct {
	gorm.Model
	Title          string `gorm:"not null"`
	Description    string `gorm:"type:text"`
	DueDate        time.Time
	PointsValue    float64
	SubmissionType string // "text", "file", "quiz", etc.
	AllowLate      bool
	LatePenalty    float64 // percentage penalty
	CourseID       uint
	Submissions    []Submission
}

type Submission struct {
	gorm.Model
	AssignmentID uint
	UserID       uint
	Content      string `gorm:"type:text"`
	FilePath     string
	SubmittedAt  time.Time
	IsLate       bool
	Grade        float64
	Feedback     string `gorm:"type:text"`
}

type Announcement struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Content     string `gorm:"type:text"`
	SendEmail   bool
	CourseID    uint
	PublisherID uint
	Publisher   User `gorm:"foreignkey:PublisherID"`
}

type Discussion struct {
	gorm.Model
	Title    string `gorm:"not null"`
	Content  string `gorm:"type:text"`
	CourseID uint
	UserID   uint
	User     User
	Replies  []DiscussionReply
}

type DiscussionReply struct {
	gorm.Model
	Content      string `gorm:"type:text"`
	DiscussionID uint
	UserID       uint
	User         User
}

type Message struct {
	gorm.Model
	Subject    string `gorm:"not null"`
	Content    string `gorm:"type:text"`
	SenderID   uint
	Sender     User `gorm:"foreignkey:SenderID"`
	ReceiverID uint
	Receiver   User `gorm:"foreignkey:ReceiverID"`
	IsRead     bool `gorm:"default:false"`
}
