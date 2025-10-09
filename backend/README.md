# Backend Project

This project is a simple backend application built using the Gin framework in Go.

## Get Started

1. **Install dependencies:**

   ```
   go mod tidy
   ```

2. **Run the application:**
   ```
   go run .
   ```

Once the application is running, visit the `:8080` port to access endpoints.

## Coding Style

- Follow [Go Style](https://google.github.io/styleguide/go/).
- Use Golang best practices.
- Use Gin best practices.

## File/Folder Structure

```
backend
├── controllers      # Contains the controller logic
│   └── controller.go
├── routes           # Defines the application routes
│   └── routes.go
├── models           # Contains data structures and methods
│   └── model.go
├── middleware       # Middleware functions for the application
|   └── middleware.go
├── go.mod               # Module definition file
├── go.sum               # Dependency checksums
├── main.go          # Entry point of the application
└── README.md            # Project documentation
```
