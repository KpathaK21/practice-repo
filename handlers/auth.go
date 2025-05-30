package handlers

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/KpathaK21/practice-repo/db"
	"github.com/KpathaK21/practice-repo/models"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper, hasNumber, hasSpecial := false, false, false
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}
	return hasUpper && hasNumber && hasSpecial
}

func generateCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	code := make([]byte, 6)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tmpl := template.Must(template.ParseFiles("static/signup.html"))
		tmpl.Execute(w, nil)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	confirmEmail := r.FormValue("confirm_email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validation checks
	switch {
	case username == "":
		renderSignupError(w, "Username is required")
	case !isValidEmail(email):
		renderSignupError(w, "Invalid email address")
	case email != confirmEmail:
		renderSignupError(w, "Emails do not match")
	case password != confirmPassword:
		renderSignupError(w, "Passwords do not match")
	case !isStrongPassword(password):
		renderSignupError(w, "Password must be at least 8 characters, with 1 uppercase, 1 number, and 1 special character")
	default:
		// Check if user already exists
		var existingUser models.User
		if db.DB.Where("email = ?", email).First(&existingUser).Error == nil {
			renderSignupError(w, "Email already registered")
			return
		}

		// Create new user
		user := models.User{
			Username: username,
			Email:    email,
		}

		// Set password
		if err := user.SetPassword(password); err != nil {
			renderSignupError(w, "Error creating account")
			return
		}

		// Generate verification code
		code := generateCode()
		user.VerificationCode = code
		user.VerificationCodeCreated = time.Now().Unix() // Set the creation time
		user.IsVerified = false

		// Save user to database
		if err := db.DB.Create(&user).Error; err != nil {
			renderSignupError(w, "Error creating account: "+err.Error())
			return
		}

		// Send verification email with improved content
		emailBody := fmt.Sprintf(`
Dear %s,

Thank you for registering with our Learning Management System!

To complete your registration, please use the following verification code:

*************************************
YOUR VERIFICATION CODE: %s
*************************************

This code will expire in 10 minutes. If you did not request this verification, please ignore this email.

Best regards,
The Learning Management System Team
		`, username, code)

		// Print the verification code to the console for debugging
		fmt.Println("Generated verification code for", email, ":", code)

		// Try to send the email
		err := sendEmail(email, "Verification Code", emailBody)
		if err != nil {
			fmt.Println("Error sending email:", err)
			// If email sending fails, render an error message
			renderSignupError(w, "We couldn't send the verification email. Please try again later or contact support.")
			return
		}

		fmt.Println("Verification email sent successfully to:", email)

		// Redirect to verification page with email pre-filled
		tmpl := template.Must(template.ParseFiles("static/verify.html"))
		tmpl.Execute(w, struct {
			Email     string
			ShowError bool
		}{
			Email:     email,
			ShowError: false,
		})
	}
}

func renderSignupError(w http.ResponseWriter, message string) {
	tmpl := template.Must(template.ParseFiles("static/signup.html"))
	tmpl.Execute(w, struct{ Error string }{Error: message})
}

func Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tmpl := template.Must(template.ParseFiles("static/verify.html"))
		tmpl.Execute(w, struct {
			Email     string
			ShowError bool
		}{
			Email:     r.URL.Query().Get("email"),
			ShowError: false,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		renderVerifyError(w, "Invalid form")
		return
	}

	email := r.FormValue("email")
	code := r.FormValue("code")

	if email == "" || code == "" {
		renderVerifyError(w, "Email and verification code are required")
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		renderVerifyError(w, "User not found")
		return
	}

	if user.VerificationCode != code {
		renderVerifyError(w, "Incorrect verification code")
		return
	}

	// Check if the verification code has expired (10 minutes = 600 seconds)
	if time.Now().Unix()-user.VerificationCodeCreated > 600 {
		renderVerifyError(w, "Verification code has expired. Please request a new one.")
		return
	}

	user.IsVerified = true
	user.VerificationCode = ""
	db.DB.Save(&user)

	// Send welcome email with improved content
	welcomeEmailBody := fmt.Sprintf(`
Dear Student,

Welcome to the Learning Management System! Your email has been successfully verified.

You can now sign in to your account and start exploring our courses and learning resources.

Best regards,
The Learning Management System Team
	`)

	sendEmail(user.Email, "Welcome", welcomeEmailBody)

	// Redirect to sign in page with success message
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func renderVerifyError(w http.ResponseWriter, message string) {
	tmpl := template.Must(template.ParseFiles("static/verify.html"))
	tmpl.Execute(w, struct {
		Email     string
		Error     string
		ShowError bool
	}{
		Email:     "",
		Error:     message,
		ShowError: true,
	})
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tmpl := template.Must(template.ParseFiles("static/signin.html"))
		tmpl.Execute(w, nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		renderSigninError(w, "Invalid form data")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		renderSigninError(w, "Invalid email or password")
		return
	}

	if !user.CheckPassword(password) {
		renderSigninError(w, "Invalid email or password")
		return
	}

	if !user.IsVerified {
		renderSigninError(w, "Email not verified. Please check your email.")
		return
	}

	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func renderSigninError(w http.ResponseWriter, message string) {
	tmpl := template.Must(template.ParseFiles("static/signin.html"))
	tmpl.Execute(w, struct{ Error string }{Error: message})
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/dashboard.html"))
	tmpl.Execute(w, nil)
}

// extractVerificationCode extracts the verification code from the email body
func extractVerificationCode(body string) string {
	// Look for a 6-character code (assuming the code is 6 characters)
	re := regexp.MustCompile(`[A-Z0-9]{6}`)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// InitEnvVars sets default environment variables if they're not set
func InitEnvVars() {
	// Check if SENDGRID_API_KEY is set
	apiKey := os.Getenv("SENDGRID_API_KEY")
	if apiKey == "" {
		fmt.Println("WARNING: SENDGRID_API_KEY environment variable is not set")
		fmt.Println("Please set a valid SendGrid API key to enable email sending")
		fmt.Println("You can get an API key by signing up at https://sendgrid.com/")
		fmt.Println("Then set the environment variable with: export SENDGRID_API_KEY=your_api_key")
		fmt.Println("Setting a temporary API key for testing (this will not work for sending real emails)")
		os.Setenv("SENDGRID_API_KEY", "SG.yourapikey")
	} else {
		fmt.Println("SENDGRID_API_KEY is set")
	}

	// Check if SENDGRID_FROM_EMAIL is set
	fromEmail := os.Getenv("SENDGRID_FROM_EMAIL")
	if fromEmail == "" {
		fmt.Println("WARNING: SENDGRID_FROM_EMAIL environment variable is not set")
		fmt.Println("Please set a valid sender email address")
		fmt.Println("Set the environment variable with: export SENDGRID_FROM_EMAIL=your_email@example.com")
		fmt.Println("Setting a temporary sender email for testing")
		os.Setenv("SENDGRID_FROM_EMAIL", "noreply@example.com")
	} else {
		fmt.Println("SENDGRID_FROM_EMAIL is set to:", fromEmail)
	}
}

func sendEmail(to, subject, body string) error {
	// Debug: Check if SendGrid API key and from email are set
	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromEmail := os.Getenv("SENDGRID_FROM_EMAIL")

	// Log email sending attempt
	fmt.Println("Attempting to send email to:", to)

	// Safely log a portion of the API key
	if len(apiKey) > 10 {
		fmt.Println("Using API key:", apiKey[:5]+"..."+apiKey[len(apiKey)-5:])
	} else {
		fmt.Println("Using API key: [REDACTED]")
	}

	fmt.Println("From email:", fromEmail)
	fmt.Println("Subject:", subject)
	// Create a more professional sender name and email
	// Use a consistent, recognizable sender name
	from := mail.NewEmail("Learning Management System", fromEmail)

	// Extract username from email for personalization
	username := to
	if atIndex := strings.Index(to, "@"); atIndex > 0 {
		username = to[:atIndex]
	}

	// Use the recipient's name for personalization
	toEmail := mail.NewEmail(username, to)

	// Modify subject to avoid spam triggers
	// Avoid excessive punctuation, all caps, and spam trigger words
	safeSubject := subject
	if !strings.HasPrefix(subject, "RE:") && !strings.HasPrefix(subject, "FW:") {
		safeSubject = subject // Use the original subject without prefix
	}

	// Create HTML version of the email for better deliverability
	// Ensure good text-to-HTML ratio and avoid spam trigger patterns
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	   <meta charset="UTF-8">
	   <title>%s</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 20px;">
	   <div style="max-width: 600px; margin: 0 auto;">
	       <h2 style="color: #35424a;">Learning Management System</h2>
	       <div style="margin: 20px 0;">
	           %s
	       </div>
	       <div style="background-color: #f0f7ff; border: 1px solid #cce5ff; padding: 15px; margin: 20px 0; text-align: center;">
	           <p style="font-weight: bold; margin-bottom: 10px;">Your Verification Code:</p>
	           <p style="font-family: monospace; font-size: 24px; letter-spacing: 2px;">%s</p>
	           <p style="font-size: 14px;">This code will expire in 10 minutes</p>
	       </div>
	       <p>To ensure you receive our emails, please add <strong>%s</strong> to your contacts.</p>
	       <p style="font-size: 12px; color: #777; margin-top: 30px;">
	           This is an automated message from the Learning Management System.<br>
	           © 2025 Learning Management System. All rights reserved.<br>
	           123 Education Street, Knowledge City, ED 12345
	       </p>
	   </div>
</body>
</html>
	`, safeSubject, strings.Replace(body, "\n", "<br>", -1), extractVerificationCode(body), fromEmail)

	// Create the email message with both plain text and HTML versions
	message := mail.NewSingleEmail(from, safeSubject, toEmail, body, htmlContent)

	// Add minimal headers to improve deliverability
	message.Headers = make(map[string]string)
	message.Headers["List-Unsubscribe"] = fmt.Sprintf("<mailto:%s?subject=unsubscribe>", fromEmail)
	message.Headers["Precedence"] = "bulk"

	// Add a single category for tracking
	message.Categories = []string{"account"}

	// Send the email
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)

	// Log the response for debugging
	if err != nil {
		fmt.Println("Error sending email:", err)
		return err
	}

	// Check if the status code indicates success (2xx)
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Println("Email sent successfully!")
	} else {
		fmt.Println("Email sending failed with status code:", response.StatusCode)
		fmt.Println("Response Body:", response.Body)
		return fmt.Errorf("failed to send email: status code %d, body: %s", response.StatusCode, response.Body)
	}

	// Log response details for debugging
	fmt.Println("Status Code:", response.StatusCode)
	fmt.Println("Response Body:", response.Body)
	fmt.Println("Response Headers:", response.Headers)

	return nil
}
