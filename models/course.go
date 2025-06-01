package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type CourseStatus string

const (
	Draft     CourseStatus = "draft"
	Published CourseStatus = "published"
	Closed    CourseStatus = "closed"
)

type Course struct {
	gorm.Model
	Title         string       `gorm:"not null"`
	Description   string       `gorm:"type:text"`
	Term          string       `gorm:"not null"` // e.g., "Fall 2023"
	Syllabus      string       `gorm:"type:text"`
	Status        CourseStatus `gorm:"not null;default:'draft'"` // draft, published, closed
	IsPublic      bool         `gorm:"default:false"`
	ProfessorID   uint
	Professor     User         `gorm:"foreignkey:ProfessorID"`
	Students      []User       `gorm:"many2many:user_courses;"`
	Assistants    []User       `gorm:"many2many:course_assistants;"`
	Materials     []Material
	Assignments   []Assignment
	Announcements []Announcement
	Discussions   []Discussion
	Quizzes       []Quiz
	Events        []Event
}

type Material struct {
	gorm.Model
	Title       string    `gorm:"not null"`
	Description string
	FileType    string    // "pdf", "video", "link", etc.
	FilePath    string
	VisibleFrom time.Time // When the material becomes visible
	VisibleTo   time.Time // When the material is no longer visible (optional)
	ModuleID    uint      // For organizing by module/week
	CourseID    uint
}

type Module struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description string
	Order       int       // For ordering modules
	StartDate   time.Time // When this module/week starts
	EndDate     time.Time // When this module/week ends
	CourseID    uint
	Materials   []Material
}

type Assignment struct {
	gorm.Model
	Title          string    `gorm:"not null"`
	Description    string    `gorm:"type:text"`
	DueDate        time.Time
	PointsValue    float64
	SubmissionType string    // "text", "file", "quiz", etc.
	AllowLate      bool
	LatePenalty    float64   // percentage penalty
	VisibleFrom    time.Time // When the assignment becomes visible
	VisibleTo      time.Time // When the assignment is no longer visible
	ModuleID       uint      // For organizing by module/week
	CourseID       uint
	Submissions    []Submission
}

type Submission struct {
	gorm.Model
	AssignmentID uint
	UserID       uint
	User         User
	Content      string    // For text submissions
	FilePath     string    // For file submissions
	SubmittedAt  time.Time
	IsLate       bool
	Grade        float64
	Feedback     string
	GradedBy     uint      // ID of professor or TA who graded
	GradedAt     time.Time
}

type Quiz struct {
	gorm.Model
	Title        string    `gorm:"not null"`
	Description  string    `gorm:"type:text"`
	DueDate      time.Time
	TimeLimit    int       // in minutes
	Attempts     int       // number of attempts allowed
	PointsValue  float64
	VisibleFrom  time.Time // When the quiz becomes visible
	VisibleTo    time.Time // When the quiz is no longer visible
	ModuleID     uint      // For organizing by module/week
	CourseID     uint
	Questions    []Question
	Submissions  []QuizSubmission
}

type QuestionType string

const (
	MultipleChoice QuestionType = "multiple_choice"
	ShortAnswer    QuestionType = "short_answer"
	Essay          QuestionType = "essay"
	TrueFalse      QuestionType = "true_false"
)

type Question struct {
	gorm.Model
	QuizID      uint
	Text        string       `gorm:"not null"`
	Type        QuestionType `gorm:"not null"`
	Options     string       `gorm:"type:text"` // JSON array of options for multiple choice
	CorrectAnswer string       // For auto-graded questions
	PointsValue float64
}

type QuizSubmission struct {
	gorm.Model
	QuizID      uint
	UserID      uint
	User        User
	StartedAt   time.Time
	CompletedAt time.Time
	IsCompleted bool
	Grade       float64
	Answers     []QuizAnswer
}

type QuizAnswer struct {
	gorm.Model
	QuizSubmissionID uint
	QuestionID       uint
	Answer           string `gorm:"type:text"`
	IsCorrect        bool   // For auto-graded questions
	Grade            float64
	Feedback         string
}

type Announcement struct {
	gorm.Model
	Title     string `gorm:"not null"`
	Content   string `gorm:"type:text"`
	CourseID  uint
	CreatedBy uint // ID of professor or TA who created
	Pinned    bool // Whether announcement is pinned to top
}

type Discussion struct {
	gorm.Model
	Title     string `gorm:"not null"`
	Content   string `gorm:"type:text"`
	CourseID  uint
	CreatedBy uint // ID of user who created
	IsClosed  bool // Whether discussion is closed for new replies
	Replies   []DiscussionReply
}

type DiscussionReply struct {
	gorm.Model
	Content      string `gorm:"type:text"`
	DiscussionID uint
	UserID       uint
	User         User
	ParentID     uint // For nested replies
}

type Message struct {
	gorm.Model
	Content   string `gorm:"type:text"`
	SenderID  uint
	Sender    User
	ReceiverID uint
	Receiver  User
	CourseID  uint // Optional, for course-related messages
	IsRead    bool
}

type Event struct {
	gorm.Model
	Title       string    `gorm:"not null"`
	Description string    `gorm:"type:text"`
	StartTime   time.Time `gorm:"not null"`
	EndTime     time.Time `gorm:"not null"`
	Location    string    // Physical location or online link
	CourseID    uint
	CreatedBy   uint // ID of professor or TA who created
}
