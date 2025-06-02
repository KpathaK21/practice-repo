package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// ZoomCredentials stores OAuth tokens for Zoom API
type ZoomCredentials struct {
	gorm.Model
	AccessToken  string    `gorm:"type:text"`
	RefreshToken string    `gorm:"type:text"`
	ExpiresAt    time.Time
}

// ZoomMeeting represents a Zoom meeting associated with a course
type ZoomMeeting struct {
	gorm.Model
	CourseID       uint
	Course         Course `gorm:"foreignkey:CourseID"`
	Title          string
	Description    string `gorm:"type:text"`
	StartTime      time.Time
	Duration       int    // in minutes
	ZoomMeetingID  string
	JoinURL        string `gorm:"type:text"`
	StartURL       string `gorm:"type:text"`
	Password       string
	Status         string // scheduled, started, finished, cancelled
}