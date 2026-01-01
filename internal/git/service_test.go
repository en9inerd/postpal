package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func TestService_RepoExists(t *testing.T) {
	tempDir := t.TempDir()

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	// Should not exist initially (directory exists but not a git repo)
	// Check if .git directory exists
	gitDir := filepath.Join(tempDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		t.Error("expected .git directory to not exist initially")
	}

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	// Should exist now
	if !service.RepoExists() {
		t.Error("expected repo to exist")
	}

	// Verify .git directory exists
	if _, err := os.Stat(gitDir); err != nil {
		t.Error("expected .git directory to exist after init")
	}
}

func TestService_Open(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	repo, err := service.Open()
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	if repo == nil {
		t.Error("expected repo to be non-nil")
	}
}

func TestService_Open_NonExistent(t *testing.T) {
	tempDir := t.TempDir()

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	_, err := service.Open()
	if err == nil {
		t.Error("expected error when opening non-existent repo")
	}
}

func TestService_AssignAuthor(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test User", Email: "test@example.com"},
	)

	err = service.AssignAuthor()
	if err != nil {
		t.Fatalf("failed to assign author: %v", err)
	}

	// Verify author was set
	repo, err := service.Open()
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	cfg, err := repo.Config()
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}

	if cfg.User.Name != "Test User" {
		t.Errorf("expected name to be 'Test User', got '%s'", cfg.User.Name)
	}

	if cfg.User.Email != "test@example.com" {
		t.Errorf("expected email to be 'test@example.com', got '%s'", cfg.User.Email)
	}
}

func TestService_Add_SingleFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Add file using relative path
	err = service.Add("test.txt")
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	// Verify file is staged
	repo, err := service.Open()
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	status, err := wt.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if status.File("test.txt").Staging != git.Added {
		t.Error("expected file to be staged")
	}
}

func TestService_Add_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	// Create multiple test files
	files := []string{"test1.txt", "test2.txt", "test3.txt"}
	for _, filename := range files {
		testFile := filepath.Join(tempDir, filename)
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Add multiple files at once using variadic arguments
	err = service.Add(files...)
	if err != nil {
		t.Fatalf("failed to add files: %v", err)
	}

	// Verify all files are staged
	repo, err := service.Open()
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	status, err := wt.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	for _, filename := range files {
		if status.File(filename).Staging != git.Added {
			t.Errorf("expected file %s to be staged", filename)
		}
	}
}

func TestService_Add_AbsolutePath(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Add file using absolute path
	err = service.Add(testFile)
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	// Verify file is staged
	repo, err := service.Open()
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	status, err := wt.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if status.File("test.txt").Staging != git.Added {
		t.Error("expected file to be staged")
	}
}

func TestService_Add_NoFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	// Try to add with no files
	err = service.Add()
	if err == nil {
		t.Error("expected error when adding with no files")
	}
}

func TestService_Add_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	err = service.Add("nonexistent.txt")
	if err == nil {
		t.Error("expected error when adding non-existent file")
	}
}

func TestService_Remove(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	repo, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	// Create and commit a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	_, err = wt.Add("test.txt")
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	_, err = wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Remove file
	err = service.Remove("test.txt")
	if err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}

	// Verify file is staged for removal
	status, err := wt.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if status.File("test.txt").Staging != git.Deleted {
		t.Error("expected file to be staged for deletion")
	}
}

func TestService_Commit(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test User", Email: "test@example.com"},
	)

	// Create and add a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err = service.Add("test.txt")
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	// Commit
	err = service.Commit("test commit")
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Verify commit was created
	repo, err := service.Open()
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("failed to get head: %v", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		t.Fatalf("failed to get commit: %v", err)
	}

	if commit.Message != "test commit" {
		t.Errorf("expected commit message to be 'test commit', got '%s'", commit.Message)
	}

	if commit.Author.Name != "Test User" {
		t.Errorf("expected author name to be 'Test User', got '%s'", commit.Author.Name)
	}

	if commit.Author.Email != "test@example.com" {
		t.Errorf("expected author email to be 'test@example.com', got '%s'", commit.Author.Email)
	}
}

func TestService_Commit_NoChanges(t *testing.T) {
	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	// Try to commit with no changes
	err = service.Commit("test commit")
	if err == nil {
		t.Error("expected error when committing with no changes")
	}
}

func TestService_CommitAndPush(t *testing.T) {
	// This test requires a real repository or mocking
	// For now, we'll test the commit part and skip push
	// In a real scenario, you'd use a test repository or mock the push

	tempDir := t.TempDir()

	// Create a git repo
	_, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test User", Email: "test@example.com"},
	)

	// Create and add a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err = service.Add("test.txt")
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	// Test commit part (push will fail without remote, which is expected)
	ctx := context.Background()
	err = service.CommitAndPush(ctx, "test commit")
	// We expect push to fail, but commit should succeed
	// In a real test, you'd set up a test remote or mock it
	if err != nil {
		// Push error is expected without a real remote
		// Verify commit was created
		repo, err2 := service.Open()
		if err2 != nil {
			t.Fatalf("failed to open repo: %v", err2)
		}

		head, err2 := repo.Head()
		if err2 != nil {
			t.Fatalf("failed to get head: %v", err2)
		}

		if head.Hash().IsZero() {
			t.Error("expected commit to be created")
		}
	}
}

func TestService_NewService(t *testing.T) {
	service := NewService(
		"/tmp/repo",
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
	)

	if service.repoDir != "/tmp/repo" {
		t.Errorf("expected repoDir to be '/tmp/repo', got '%s'", service.repoDir)
	}

	if service.repoURL != "https://github.com/test/repo.git" {
		t.Errorf("expected repoURL to be 'https://github.com/test/repo.git', got '%s'", service.repoURL)
	}

	if service.branch != "main" {
		t.Errorf("expected branch to be 'main', got '%s'", service.branch)
	}

	if service.authToken != "token" {
		t.Errorf("expected authToken to be 'token', got '%s'", service.authToken)
	}

	if service.author.Name != "Test" {
		t.Errorf("expected author name to be 'Test', got '%s'", service.author.Name)
	}

	if service.author.Email != "test@example.com" {
		t.Errorf("expected author email to be 'test@example.com', got '%s'", service.author.Email)
	}
}
