# Learning Management System

A comprehensive Learning Management System (LMS) built with Go, providing features for course management, assignments, quizzes, discussions, and more.

## Tech Stack

### Backend
- **Language**: Go (version 1.23.0)
- **Web Framework**: Standard Go `net/http` package for HTTP server
- **Database**: PostgreSQL with GORM (v1.9.16) as the ORM
- **Authentication**: Custom JWT implementation using `golang-jwt/jwt/v5` (v5.2.2)
- **Password Hashing**: `golang.org/x/crypto` (v0.38.0)
- **Email Service**: SendGrid via `sendgrid-go` (v3.16.1)

### Frontend
- **HTML/CSS/JavaScript**: Pure vanilla implementation without external frameworks
- **Architecture**: Server-side rendered HTML with minimal client-side JavaScript
- **Authentication**: JWT-based authentication with cookies for session management

## Project Structure
Trae

I'll create a comprehensive README.md file that documents the tech stack and important information about this Learning Management System project.

models/course.go :( 1 - 50 )

# Learning Management System
A comprehensive Learning Management System (LMS) built with Go, providing features for course management, assignments, quizzes, discussions, and more.

## Tech Stack
### Backend
- Language : Go (version 1.23.0)
- Web Framework : Standard Go net/http package for HTTP server
- Database : PostgreSQL with GORM (v1.9.16) as the ORM
- Authentication : Custom JWT implementation using golang-jwt/jwt/v5 (v5.2.2)
- Password Hashing : golang.org/x/crypto (v0.38.0)
- Email Service : SendGrid via sendgrid-go (v3.16.1)
### Frontend
- HTML/CSS/JavaScript : Pure vanilla implementation without external frameworks
- Architecture : Server-side rendered HTML with minimal client-side JavaScript
- Authentication : JWT-based authentication with cookies for session management
## Project Structure
```
├── db/                 # Database 
configuration and initialization
├── handlers/           # HTTP request 
handlers (controllers)
│   ├── auth.go         # Authentication 
handlers
│   ├── course.go       # Course management 
handlers
│   └── jwt.go          # JWT authentication 
middleware
├── main.go             # Application entry 
point
├── models/             # Data models and 
database interactions
│   ├── course.go       # Course-related models
│   └── user.go         # User model and 
authentication
└── static/             # Frontend assets
    ├── js/             # JavaScript files
    └── *.html          # HTML templates
```
## Features
- User Management
  
  - Role-based access control (Student, Professor, TA)
  - Email verification
  - Secure authentication with JWT
- Course Management
  
  - Course creation and enrollment
  - Material management
  - Assignment creation and submission
  - Grading system
- Communication
  
  - Announcements
  - Discussion forums
  - Messaging system
- Assessment
  
  - Quiz creation and submission
  - Assignment grading
  - Grade export
- Calendar
  
  - Event management
  - Course schedule
## Setup and Installation
### Prerequisites
- Go 1.23.0 or higher
- PostgreSQL database
- SendGrid API key (for email functionality)
### Database Configuration
The application uses PostgreSQL. Update the database connection string in db/db.go with your credentials:

```
dsn := "host=localhost port=5432 
user=yourusername dbname=yourdbname 
password=yourpassword sslmode=disable"
```
### Environment Variables
Set up the following environment variables or update them in the code:

- JWT secret keys
- SendGrid API key
- Application URL
### Running the Application
1. Clone the repository
2. Install dependencies: go mod download
3. Run the application: go run main.go
4. Access the application at http://localhost:8080 (or configured port)
## API Endpoints
### Authentication
- /signup - User registration
- /signin - User login
- /verify - Email verification
- /refresh - Refresh JWT token
- /logout - User logout
### Courses
- /courses - List available courses
- /course/create - Create a new course (Professor only)
- /course/update - Update course details (Professor only)
- /course/assign-ta - Assign a TA to a course (Professor only)
- /course/enroll-student - Enroll a student in a course (Professor only)
### Content Management
- /course/materials/create - Add course materials (Professor/TA only)
- /course/materials/update - Update course materials (Professor/TA only)
- /course/materials/delete - Delete course materials (Professor/TA only)
### Assignments
- /course/assignments/create - Create assignments (Professor/TA only)
- /course/assignments/submit - Submit assignments (Students only)
- /course/assignments/grade - Grade submissions (Professor/TA only)
## License
[Add your license information here]

## Contributors
[Add contributor information here]

KpathaK21

Copy

Delete

make a README.md file in the repo

Trae

I'll create a README.md file for your Learning Management System project. Here's the content you can save to a README.md file in your repository root:

```
# Learning Management System

A comprehensive Learning Management System 
(LMS) built with Go, providing features for 
course management, assignments, quizzes, 
discussions, and more.

## Tech Stack

### Backend
- **Language**: Go (version 1.23.0)
- **Web Framework**: Standard Go `net/http` 
package for HTTP server
- **Database**: PostgreSQL with GORM (v1.9.16) 
as the ORM
- **Authentication**: Custom JWT 
implementation using `golang-jwt/jwt/v5` (v5.2.
2)
- **Password Hashing**: `golang.org/x/crypto` 
(v0.38.0)
- **Email Service**: SendGrid via 
`sendgrid-go` (v3.16.1)

### Frontend
- **HTML/CSS/JavaScript**: Pure vanilla 
implementation without external frameworks
- **Architecture**: Server-side rendered HTML 
with minimal client-side JavaScript
- **Authentication**: JWT-based authentication 
with cookies for session management

## Project Structure

```
├── db/                 # Database configuration and initialization
├── handlers/           # HTTP request handlers (controllers)
│   ├── auth.go         # Authentication handlers
│   ├── course.go       # Course management handlers
│   └── jwt.go          # JWT authentication middleware
├── main.go             # Application entry point
├── models/             # Data models and database interactions
│   ├── course.go       # Course-related models
│   └── user.go         # User model and authentication
└── static/             # Frontend assets
├── js/             # JavaScript files
└── *.html          # HTML templates
```
## Features
- User Management
  - Role-based access control (Student, Professor, TA)
  - Email verification
  - Secure authentication with JWT
- Course Management
  - Course creation and enrollment
  - Material management
  - Assignment creation and submission
  - Grading system
- Communication
  - Announcements
  - Discussion forums
  - Messaging system
- Assessment
  - Quiz creation and submission
  - Assignment grading
  - Grade export
- Calendar
  - Event management
  - Course schedule
## Setup and Installation
### Prerequisites
- Go 1.23.0 or higher
- PostgreSQL database
- SendGrid API key (for email functionality)
### Database Configuration
The application uses PostgreSQL. Update the database connection string in db/db.go with your credentials:
```
dsn := "host=localhost port=5432 user=yourusername dbname=yourdbname password=yourpassword sslmode=disable"
```
### Environment Variables
Set up the following environment variables or update them in the code:
- JWT secret keys
- SendGrid API key
- Application URL
### Running the Application
1. Clone the repository
2. Install dependencies: go mod download    
3. Run the application: go run main.go
4. Access the application at
3. Run the application: go run main.go
4. Access the application at URL_ADDRESS:8080 (or configured port)