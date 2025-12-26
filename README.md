# Go Web Application Template

This is a starter template for building web applications in Go. It includes the common pieces you'll need for a production-ready application, so you can focus on building your features instead of setting up infrastructure.

## What's Included

The template comes with a clean project structure that separates your application code from infrastructure concerns. You get an HTTP server with graceful shutdown, structured logging, configuration management, request validation utilities, security headers middleware, Docker support, and CI/CD workflows.

The project follows a standard Go layout with your application entry point in `cmd/app`, internal packages in `internal`, and UI assets in `ui`.

## Project Structure

Here's how the project is organized:

```
.
├── cmd/
│   └── app/              # Application entry point
│       └── main.go
├── internal/
│   ├── config/           # Configuration parsing
│   ├── log/              # Logging utilities
│   ├── server/           # HTTP server setup and handlers
│   └── validator/        # Validation utilities
├── ui/
│   ├── static/           # Static assets (CSS, JS)
│   └── templates/        # HTML templates
├── scripts/              # Build and utility scripts
├── .github/
│   └── workflows/        # CI/CD workflows
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Getting Started

### Step 1: Copy the Template

First, copy this template to wherever you want your new project to live:

```bash
cp -r go-template /path/to/your-new-project
cd /path/to/your-new-project
```

### Step 2: Update the Module Name

You'll need to replace the placeholder module name with your actual module path. The template uses `github.com/yourusername/yourproject` as a placeholder.

Update `go.mod` first, then update all the import statements in your Go files. You can do this with find and replace in your editor, or use sed if you prefer:

```bash
# Update go.mod
sed -i '' 's|github.com/yourusername/yourproject|github.com/yourusername/yourproject|g' go.mod

# Update all Go files
find . -name "*.go" -type f -exec sed -i '' 's|github.com/yourusername/yourproject|github.com/yourusername/yourproject|g' {} +
```

### Step 3: Customize Your Configuration

Open `internal/config/config.go` and add the configuration fields your application needs. The template includes a basic `Port` field, but you'll likely want to add things like database URLs, API keys, timeouts, and other settings.

```go
type Config struct {
    Port        string
    DatabaseURL string
    APIKey      string
    // Add your fields here
}
```

### Step 4: Add Your Handlers

The template includes a basic server setup, but you'll need to add your own routes and handlers. Edit `internal/server/server.go` and `internal/server/handlers.go` to add your application logic.

For example, you might add routes like this:

```go
// In registerAPIRoutes
apiGroup.HandleFunc("GET /users", getUsersHandler(logger, cfg))
apiGroup.HandleFunc("POST /users", createUserHandler(logger, cfg))
```

### Step 5: Initialize Go Modules

Once you've updated the module name, run:

```bash
go mod tidy
```

This will download dependencies and update your go.mod file.

### Step 6: Build and Run

You can build the application using the Makefile:

```bash
# Build
make build

# Run
./dist/yourproject --port 8000 --verbose

# Or run directly during development
go run ./cmd/app --port 8000 --verbose
```

## Features

### Configuration

The configuration system supports both command-line flags and environment variables. Flags take precedence, but environment variables are useful for deployment scenarios.

You can set the port either way:

```bash
# Via flag
./app --port 8080

# Via environment variable
APP_PORT=8080 ./app
```

### Logging

Logging is controlled by the `--verbose` or `-v` flag. When verbose mode is enabled, you'll see debug-level logs. Without it, only errors are logged.

```bash
# Verbose logging (debug level)
./app --verbose

# Silent mode (errors only)
./app
```

### Docker

The template includes Docker support for containerized deployments. You can build and run the application in Docker:

```bash
# Build
docker build -t yourproject:latest .

# Run
docker run -p 8000:8000 yourproject:latest

# Or use docker-compose
docker-compose up
```

### Multi-Architecture Builds

If you need to build for multiple platforms, use the production build target:

```bash
make build-prod
```

This creates binaries in the `dist/` directory for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)

### CI/CD

The template includes a GitHub Actions workflow that handles semantic versioning and Docker builds. It automatically determines version bumps based on your commit messages and builds Docker images when you create a release.

The workflow uses semantic versioning based on commit message prefixes:
- `feat:` - Minor version bump
- `fix:` - Patch version bump
- `feat!:` or `BREAKING CHANGE:` - Major version bump
- `chore:` - No release (Docker build is skipped)

To use the CI/CD workflow, you'll need to set up these secrets in your GitHub repository:
- `DOCKERHUB_USERNAME` - Your Docker Hub username
- `DOCKERHUB_TOKEN` - Your Docker Hub access token

## Makefile Targets

The Makefile includes several common tasks:

- `make build` - Build binary locally
- `make build-prod` - Build for multiple platforms
- `make clean` - Remove build artifacts
- `make format` - Format Go code
- `make test` - Run tests
- `make docker-build` - Build Docker image
- `make docker-clean` - Clean up Docker resources
- `make docker-clean-all` - Clean Docker resources and build cache

## UI Templates

The template includes a basic HTML template structure that you can use if you're building a web application with server-rendered pages. The templates are in `ui/templates/` and static assets go in `ui/static/`.

To use the templates, you'll need to uncomment the embed directive in `ui/efs.go`:

```go
//go:embed "templates/*" "static/*"
var Files embed.FS
```

Then you can use the template cache in your handlers. There's an example in `internal/server/server.go` that shows how to set this up.

## Validation

The `internal/validator` package provides utilities for validating request data. You can embed the validator in your request structs and implement a `Validate` method:

```go
import "github.com/yourusername/yourproject/internal/validator"

type UserRequest struct {
    Email string `json:"email"`
    validator.Validator
}

func (r *UserRequest) Validate(v *validator.Validator) {
    v.CheckField(validator.NotBlank(r.Email), "email", "email is required")
    v.CheckField(validator.Matches(r.Email, emailRegex), "email", "invalid email format")
}
```

## Security

The template includes several security features out of the box:
- Security headers middleware (CSP, X-Frame-Options, etc.)
- Request throttling
- Request size limits
- Graceful shutdown
- Health check endpoint

You can customize these in `internal/server/middleware.go` and `internal/server/server.go`.

## Dependencies

The template uses `github.com/en9inerd/go-pkgs` for router and middleware utilities. If you prefer a different router or middleware library, you can replace it. The code is structured to make this straightforward.

## Customization Checklist

Before you start building, here's a checklist of things to customize:

- Update module name in `go.mod` and all Go files
- Add your configuration fields in `internal/config/config.go`
- Implement your handlers in `internal/server/handlers.go`
- Register your routes in `internal/server/server.go`
- Update Docker image name in `.github/workflows/release-and-docker.yml`
- Customize UI templates in `ui/templates/` if you're using them
- Add your static assets to `ui/static/` if needed
- Update `.env.example` with your environment variables
- Update `docker-compose.yml` with your service configuration
- Add any additional dependencies: `go get <package>`
- Write tests for your handlers
- Update this README with your project-specific information

## License

This template is provided as-is. Feel free to customize it for your needs.
