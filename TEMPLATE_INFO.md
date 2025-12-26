# Template Extraction Summary

This template was extracted from a production Go web application. Below is what was included and what needs customization.

## What's Included

### Core Application Structure
- ✅ `cmd/app/main.go` - Application entry point with graceful shutdown
- ✅ `internal/config/` - Configuration parsing (flags + env vars)
- ✅ `internal/log/` - Structured logging with slog
- ✅ `internal/server/` - HTTP server setup, middleware, and handler structure
- ✅ `internal/validator/` - Validation utilities (fully generic)

### Infrastructure
- ✅ `Dockerfile` - Multi-stage build for Go applications
- ✅ `docker-compose.yml` - Docker Compose configuration
- ✅ `.dockerignore` - Docker build exclusions
- ✅ `.gitignore` - Git exclusions for Go projects
- ✅ `Makefile` - Common build and development tasks
- ✅ `scripts/build.sh` - Multi-platform build script

### CI/CD
- ✅ `.github/workflows/release-and-docker.yml` - Semantic versioning + Docker builds

### UI Structure (Optional)
- ✅ `ui/efs.go` - Embedded filesystem setup
- ✅ `ui/templates/` - HTML template structure
- ✅ `ui/static/` - Static assets (CSS, JS)

## What Was Removed/Genericized

### Project-Specific Code
- ❌ `internal/memstore/` - Memory store implementation (project-specific)
- ❌ `internal/crypto/` - Encryption service (project-specific)
- ❌ Specific handlers (saveSecret, retrieveSecret, etc.)
- ❌ Specific config fields (MinPhraseSize, MaxPhraseSize, etc.)
- ❌ Nginx configuration files (project-specific reverse proxy setup)
- ❌ Project-specific UI templates and content

### Module References
- All `github.com/en9inerd/shhh` references changed to `github.com/yourusername/yourproject`
- Binary name changed from `shhh` to `app` (customize as needed)

## Customization Required

1. **Module Name**: Replace `github.com/yourusername/yourproject` throughout
2. **Config**: Add your application-specific configuration fields
3. **Handlers**: Implement your API and web handlers
4. **Routes**: Register your routes in `internal/server/server.go`
5. **Dependencies**: Add your required Go packages
6. **Docker**: Update image names in workflow and docker-compose
7. **UI**: Customize templates and static assets (or remove if not needed)

## Quick Start

```bash
# 1. Copy template
cp -r go-template my-new-project
cd my-new-project

# 2. Find and replace module name
find . -name "*.go" -type f -exec sed -i '' 's|github.com/yourusername/yourproject|github.com/yourusername/my-new-project|g' {} +
sed -i '' 's|github.com/yourusername/yourproject|github.com/yourusername/my-new-project|g' go.mod

# 3. Initialize modules
go mod tidy

# 4. Build and run
make build
./dist/my-new-project --verbose
```

## Notes

- The template uses `github.com/en9inerd/go-pkgs` for router/middleware. You can replace this with your preferred library.
- The UI embedding is commented out by default. Uncomment when you have files to embed.
- The Dockerfile uses a minimal Alpine base image. Adjust if you need additional tools.
- The CI/CD workflow uses semantic versioning. Adjust commit message format if needed.
