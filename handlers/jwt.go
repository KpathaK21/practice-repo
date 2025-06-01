package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/KpathaK21/practice-repo/db"
	"github.com/KpathaK21/practice-repo/models"
	"github.com/golang-jwt/jwt/v5"
)

// JWT secret keys - in production, these should be environment variables
var (
	accessTokenSecret  = []byte("access_secret_key")  // Change this in production
	refreshTokenSecret = []byte("refresh_secret_key") // Change this in production
)

// TokenDetails contains the metadata of both access and refresh tokens
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

// AccessTokenClaims represents the claims in the access token
type AccessTokenClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	UUID     string `json:"uuid"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims represents the claims in the refresh token
type RefreshTokenClaims struct {
	UserID uint   `json:"user_id"`
	UUID   string `json:"uuid"`
	jwt.RegisteredClaims
}

// CreateToken creates a new token for a user
func CreateToken(user models.User) (*TokenDetails, error) {
	td := &TokenDetails{}

	// Set expiration times
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()   // 15 minutes
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix() // 7 days

	// Generate UUIDs for both tokens
	td.AccessUUID = fmt.Sprintf("%d_%d", user.ID, time.Now().Unix())
	td.RefreshUUID = fmt.Sprintf("%d_%d_refresh", user.ID, time.Now().Unix())

	// Create Access Token
	atClaims := AccessTokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		UUID:     td.AccessUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(td.AtExpires, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "learning-management-system",
		},
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	var err error
	td.AccessToken, err = at.SignedString(accessTokenSecret)
	if err != nil {
		return nil, err
	}

	// Create Refresh Token
	rtClaims := RefreshTokenClaims{
		UserID: user.ID,
		UUID:   td.RefreshUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(td.RtExpires, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "learning-management-system",
		},
	}

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString(refreshTokenSecret)
	if err != nil {
		return nil, err
	}

	return td, nil
}

// VerifyAccessToken verifies the access token
func VerifyAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return accessTokenSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// VerifyRefreshToken verifies the refresh token
func VerifyRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return refreshTokenSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ExtractTokenFromRequest extracts the token from the request header
func ExtractTokenFromRequest(r *http.Request) string {
	// Get token from Authorization header
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && bearer[:7] == "Bearer " {
		return bearer[7:]
	}

	// If not in header, try to get from cookie
	cookie, err := r.Cookie("access_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

// SetTokenCookies sets the token cookies
func SetTokenCookies(w http.ResponseWriter, td *TokenDetails) {
	// Set access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    td.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		MaxAge:   int(td.AtExpires - time.Now().Unix()),
	})

	// Set refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    td.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		MaxAge:   int(td.RtExpires - time.Now().Unix()),
	})
}

// 5. Create JWT Middleware

// JWTMiddleware is a middleware that checks for a valid JWT token
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from request
		tokenString := ExtractTokenFromRequest(r)
		if tokenString == "" {
			// No token found, redirect to login
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		// Verify the token
		claims, err := VerifyAccessToken(tokenString)
		if err != nil {
			// Token is invalid, try to refresh it
			http.Redirect(w, r, "/refresh", http.StatusSeeOther)
			return
		}

		// Set user information in request context
		ctx := context.WithValue(r.Context(), "user", claims)
		next(w, r.WithContext(ctx))
	}
}

// ProfessorJWTMiddleware checks if the user is a professor
func ProfessorJWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Get user claims from context
		claims, ok := r.Context().Value("user").(*AccessTokenClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user is a professor
		if claims.Role != "professor" {
			http.Error(w, "Unauthorized: Professors only", http.StatusForbidden)
			return
		}

		next(w, r)
	})
}

// TAJWTMiddleware checks if the user is a TA
func TAJWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Get user claims from context
		claims, ok := r.Context().Value("user").(*AccessTokenClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user is a TA
		if claims.Role != "ta" {
			http.Error(w, "Unauthorized: TAs only", http.StatusForbidden)
			return
		}

		next(w, r)
	})
}

// CourseStaffJWTMiddleware checks if the user is a professor or TA for a specific course
func CourseStaffJWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Get user claims from context
		claims, ok := r.Context().Value("user").(*AccessTokenClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

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

		// Get user from database to check course relationships
		var user models.User
		if err := db.DB.First(&user, claims.UserID).Error; err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Check if user is professor or TA for this course
		if !user.IsProfessorOf(uint(courseID)) && !user.IsTAOf(uint(courseID)) {
			http.Error(w, "Unauthorized: You are not staff for this course", http.StatusForbidden)
			return
		}

		next(w, r)
	})
}

// EnrolledJWTMiddleware checks if the user is enrolled in a specific course
func EnrolledJWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Get user claims from context
		claims, ok := r.Context().Value("user").(*AccessTokenClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

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

		// Get user from database to check enrollment
		var user models.User
		if err := db.DB.First(&user, claims.UserID).Error; err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Check if user is enrolled, a professor, or a TA for this course
		if !user.IsEnrolledIn(uint(courseID)) && !user.IsProfessorOf(uint(courseID)) && !user.IsTAOf(uint(courseID)) {
			http.Error(w, "Unauthorized: You are not enrolled in this course", http.StatusForbidden)
			return
		}

		next(w, r)
	})
}
