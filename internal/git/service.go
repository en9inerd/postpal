package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

// Author represents Git author information
type Author struct {
	Name  string
	Email string
}

// Service handles Git repository operations
type Service struct {
	repoDir   string
	repoURL   string
	branch    string
	authToken string
	author    Author
}

// NewService creates a new Git service
func NewService(repoDir, repoURL, branch, authToken string, author Author) *Service {
	return &Service{
		repoDir:   repoDir,
		repoURL:   repoURL,
		branch:    branch,
		authToken: authToken,
		author:    author,
	}
}

// RepoExists checks if the repository directory exists
func (s *Service) RepoExists() bool {
	_, err := os.Stat(s.repoDir)
	return err == nil
}

// Clone clones the repository
func (s *Service) Clone(ctx context.Context) error {
	auth := &http.BasicAuth{
		Username: "token",
		Password: s.authToken,
	}

	_, err := git.PlainCloneContext(ctx, s.repoDir, &git.CloneOptions{
		URL:           s.repoURL,
		Auth:          auth,
		ReferenceName: plumbing.NewBranchReferenceName(s.branch),
		SingleBranch:  true,
		Depth:         1,
		Progress:      os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return s.AssignAuthor()
}

// Open opens an existing repository
func (s *Service) Open() (*git.Repository, error) {
	repo, err := git.PlainOpen(s.repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	return repo, nil
}

// AssignAuthor sets the Git author configuration
func (s *Service) AssignAuthor() error {
	repo, err := s.Open()
	if err != nil {
		return err
	}

	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	cfg.User.Name = s.author.Name
	cfg.User.Email = s.author.Email

	return repo.SetConfig(cfg)
}

// Pull pulls the latest changes from the remote
func (s *Service) Pull(ctx context.Context) error {
	repo, err := s.Open()
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	auth := &http.BasicAuth{
		Username: "token",
		Password: s.authToken,
	}

	err = wt.PullContext(ctx, &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(s.branch),
		SingleBranch:  true,
		Auth:          auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull: %w", err)
	}

	return nil
}

// Add adds files to the staging area.
func (s *Service) Add(filePaths ...string) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("no file paths provided")
	}

	repo, err := s.Open()
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	for _, filePath := range filePaths {
		absPath := filePath
		if !filepath.IsAbs(filePath) {
			absPath = filepath.Join(s.repoDir, filePath)
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}

		relPath, err := filepath.Rel(s.repoDir, absPath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		_, err = wt.Add(relPath)
		if err != nil {
			return fmt.Errorf("failed to add file %s: %w", filePath, err)
		}
	}

	return nil
}

// Remove removes files from the repository
func (s *Service) Remove(filePath string) error {
	repo, err := s.Open()
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	absPath := filePath
	if !filepath.IsAbs(filePath) {
		absPath = filepath.Join(s.repoDir, filePath)
	}

	relPath, err := filepath.Rel(s.repoDir, absPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	_, err = wt.Remove(relPath)
	if err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

// Commit commits staged changes
func (s *Service) Commit(message string) error {
	repo, err := s.Open()
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.IsClean() {
		return fmt.Errorf("no changes to commit")
	}

	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  s.author.Name,
			Email: s.author.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// Push pushes commits to the remote repository
func (s *Service) Push(ctx context.Context) error {
	repo, err := s.Open()
	if err != nil {
		return err
	}

	auth := &http.BasicAuth{
		Username: "token",
		Password: s.authToken,
	}

	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", s.branch, s.branch)),
		},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// CommitAndPush commits and pushes in one operation
func (s *Service) CommitAndPush(ctx context.Context, message string) error {
	if err := s.Commit(message); err != nil {
		return err
	}
	return s.Push(ctx)
}
