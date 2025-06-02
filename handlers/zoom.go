package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/KpathaK21/practice-repo/db"
	"github.com/KpathaK21/practice-repo/models"
	"github.com/go-resty/resty/v2"
)

const (
	ZoomAPIBaseURL = "https://api.zoom.us/v2"
	ZoomOAuthURL   = "https://zoom.us/oauth/token"
)

var (
	zoomClientID     = os.Getenv("ZOOM_CLIENT_ID")
	zoomClientSecret = os.Getenv("ZOOM_CLIENT_SECRET")
	zoomRedirectURI  = os.Getenv("ZOOM_REDIRECT_URI")
)

// InitZoomEnvVars initializes Zoom environment variables
func InitZoomEnvVars() {
	// Set default values if not provided
	if zoomClientID == "" {
		zoomClientID = "your-zoom-client-id"
		os.Setenv("ZOOM_CLIENT_ID", zoomClientID)
	}

	if zoomClientSecret == "" {
		zoomClientSecret = "your-zoom-client-secret"
		os.Setenv("ZOOM_CLIENT_SECRET", zoomClientSecret)
	}

	if zoomRedirectURI == "" {
		zoomRedirectURI = "http://localhost:8080/zoom/callback"
		os.Setenv("ZOOM_REDIRECT_URI", zoomRedirectURI)
	}
}

// ZoomAuthHandler initiates the OAuth flow with Zoom
func ZoomAuthHandler(w http.ResponseWriter, r *http.Request) {
	authorizeURL := "https://zoom.us/oauth/authorize"
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", zoomClientID)
	params.Add("redirect_uri", zoomRedirectURI)

	// Comment out the scopes line temporarily to test if this resolves the 4700 error
	// params.Add("scope", "meeting:write:admin meeting:read:admin user:read:admin")

	// Redirect to Zoom authorization page
	http.Redirect(w, r, authorizeURL+"?"+params.Encode(), http.StatusFound)
}

// ZoomCallbackHandler handles the OAuth callback from Zoom
func ZoomCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get authorization code from query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for access token
	client := resty.New()
	resp, err := client.R().
		SetBasicAuth(zoomClientID, zoomClientSecret).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type":   "authorization_code",
			"code":         code,
			"redirect_uri": zoomRedirectURI,
		}).
		Post(ZoomOAuthURL)

	if err != nil {
		http.Error(w, "Failed to exchange authorization code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		http.Error(w, "Failed to exchange authorization code: "+resp.String(), http.StatusInternalServerError)
		return
	}

	// Parse response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(resp.Body(), &tokenResponse); err != nil {
		http.Error(w, "Failed to parse token response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save tokens to database
	credentials := models.ZoomCredentials{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}

	if err := db.DB.Create(&credentials).Error; err != nil {
		http.Error(w, "Failed to save credentials: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard with success message
	http.Redirect(w, r, "/dashboard?zoom_connected=true", http.StatusFound)
}

// refreshZoomToken refreshes the Zoom access token if expired
func refreshZoomToken() (string, error) {
	// Get the latest credentials from the database
	var credentials models.ZoomCredentials
	if err := db.DB.Order("created_at desc").First(&credentials).Error; err != nil {
		return "", fmt.Errorf("no zoom credentials found: %v", err)
	}

	// Check if token is expired
	if time.Now().Before(credentials.ExpiresAt) {
		return credentials.AccessToken, nil
	}

	// Refresh the token
	client := resty.New()
	resp, err := client.R().
		SetBasicAuth(zoomClientID, zoomClientSecret).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": credentials.RefreshToken,
		}).
		Post(ZoomOAuthURL)

	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("failed to refresh token: %s", resp.String())
	}

	// Parse response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(resp.Body(), &tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse token response: %v", err)
	}

	// Save new tokens to database
	newCredentials := models.ZoomCredentials{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}

	if err := db.DB.Create(&newCredentials).Error; err != nil {
		return "", fmt.Errorf("failed to save new credentials: %v", err)
	}

	return newCredentials.AccessToken, nil
}

// CreateZoomMeeting creates a new Zoom meeting for a course
func CreateZoomMeeting(w http.ResponseWriter, r *http.Request) {
	// Check if user is authorized (professor or TA)
	user, err := getUserFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse course ID from request
	courseIDStr := r.URL.Query().Get("course_id")
	if courseIDStr == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// Check if user is professor or TA for this course
	var course models.Course
	if err := db.DB.First(&course, courseID).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	if !user.IsProfessorOf(uint(courseID)) && !user.IsTAOf(uint(courseID)) {
		http.Error(w, "You are not authorized to create meetings for this course", http.StatusForbidden)
		return
	}

	// Parse meeting details from form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	startTimeStr := r.FormValue("start_time")
	durationStr := r.FormValue("duration")

	if title == "" || startTimeStr == "" || durationStr == "" {
		http.Error(w, "Title, start time, and duration are required", http.StatusBadRequest)
		return
	}

	// Parse start time
	startTime, err := time.Parse("2006-01-02T15:04", startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	// Parse duration
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		http.Error(w, "Invalid duration", http.StatusBadRequest)
		return
	}

	// Get Zoom access token
	accessToken, err := refreshZoomToken()
	if err != nil {
		http.Error(w, "Failed to get Zoom access token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create meeting in Zoom
	client := resty.New()
	meetingRequest := map[string]interface{}{
		"topic":      title,
		"type":       2, // Scheduled meeting
		"start_time": startTime.Format(time.RFC3339),
		"duration":   duration,
		"timezone":   "UTC",
		"agenda":     description,
		"settings": map[string]interface{}{
			"host_video":        true,
			"participant_video": true,
			"join_before_host":  true,
			"mute_upon_entry":   false,
			"waiting_room":      false,
		},
	}

	resp, err := client.R().
		SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(meetingRequest).
		Post(ZoomAPIBaseURL + "/users/me/meetings")

	if err != nil {
		http.Error(w, "Failed to create Zoom meeting: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		http.Error(w, "Failed to create Zoom meeting: "+resp.String(), http.StatusInternalServerError)
		return
	}

	// Parse response
	var meetingResponse struct {
		ID       int    `json:"id"`
		JoinURL  string `json:"join_url"`
		StartURL string `json:"start_url"`
		Password string `json:"password"`
	}

	if err := json.Unmarshal(resp.Body(), &meetingResponse); err != nil {
		http.Error(w, "Failed to parse meeting response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save meeting to database
	zoomMeeting := models.ZoomMeeting{
		CourseID:      uint(courseID),
		Title:         title,
		Description:   description,
		StartTime:     startTime,
		Duration:      duration,
		ZoomMeetingID: strconv.Itoa(meetingResponse.ID),
		JoinURL:       meetingResponse.JoinURL,
		StartURL:      meetingResponse.StartURL,
		Password:      meetingResponse.Password,
		Status:        "scheduled",
	}

	if err := db.DB.Create(&zoomMeeting).Error; err != nil {
		http.Error(w, "Failed to save meeting: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to course page
	http.Redirect(w, r, "/course?id="+courseIDStr, http.StatusFound)
}

// ListZoomMeetings lists all Zoom meetings for a course
func ListZoomMeetings(w http.ResponseWriter, r *http.Request) {
	// Parse course ID from request
	courseIDStr := r.URL.Query().Get("course_id")
	if courseIDStr == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// Get meetings from database
	var meetings []models.ZoomMeeting
	if err := db.DB.Where("course_id = ?", courseID).Find(&meetings).Error; err != nil {
		http.Error(w, "Failed to get meetings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return meetings as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meetings)
}

// JoinZoomMeeting redirects to the Zoom meeting join URL
func JoinZoomMeeting(w http.ResponseWriter, r *http.Request) {
	// Parse meeting ID from request
	meetingIDStr := r.URL.Query().Get("meeting_id")
	if meetingIDStr == "" {
		http.Error(w, "Meeting ID is required", http.StatusBadRequest)
		return
	}

	meetingID, err := strconv.ParseUint(meetingIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	// Get meeting from database
	var meeting models.ZoomMeeting
	if err := db.DB.First(&meeting, meetingID).Error; err != nil {
		http.Error(w, "Meeting not found", http.StatusNotFound)
		return
	}

	// Check if user is enrolled in the course
	user, err := getUserFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !user.IsProfessorOf(meeting.CourseID) && !user.IsTAOf(meeting.CourseID) && !user.IsEnrolledIn(meeting.CourseID) {
		http.Error(w, "You are not authorized to join this meeting", http.StatusForbidden)
		return
	}

	// Redirect to join URL
	http.Redirect(w, r, meeting.JoinURL, http.StatusFound)
}
