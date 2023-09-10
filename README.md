# Pet Project Readme

## Introduction

This readme provides an overview of my pet project, a simple Go web application for user registration, login, and session management. The application utilizes PostgreSQL database, and bcrypt for password hashing. It also uses Gorilla Sessions to manage user sessions.

## Getting Started

To run the project, you'll need Go and PostgreSQL installed on your system.

1. **Database Setup**: You must have a PostgreSQL database ready for use. Ensure that you have created a database and specified the database connection details in the `main` function where the `db` variable is initialized.

   ```go
   db, err = sql.Open("postgres", "host=localhost user=postgres password=yourpassword dbname=yourdatabase sslmode=disable")
2. **Dependencies:**: To install the project dependencies, run:

3. **Build and Run:** Build and run the project using the following command:
   ```console
   go build ./your-project-name
   ```
  The application will start on port 8080 by default. You can change the port in the main function.

  ## API Endpoints
  POST /sign-up: Register a new user with a username and password.
  POST /log-in: Log in with a registered username and password.
  GET /wellcome: Greet the logged-in user.
  POST /log-out: Log out and end the user session.
