package zola

import (
	"errors"
	"log/slog"

	"github.com/en9inerd/postpal/internal/git"
)

// ErrNoPostsDeleted is returned when no posts matched the given IDs.
var ErrNoPostsDeleted = errors.New("no posts deleted")

// Service handles Zola blog post creation and management
type Service struct {
	postsDir    string
	relPostsDir string
	repoDir     string
	channelID   string
	gitService  *git.Service
	logger      *slog.Logger
}

// NewService creates a new Zola post service
func NewService(postsDir, relPostsDir, repoDir, channelID string, gitService *git.Service, logger *slog.Logger) *Service {
	return &Service{
		postsDir:    postsDir,
		relPostsDir: relPostsDir,
		repoDir:     repoDir,
		channelID:   channelID,
		gitService:  gitService,
		logger:      logger,
	}
}
