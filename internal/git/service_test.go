package git

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	goplugin "github.com/go-git/go-git/v6/x/plugin"
	xconfig "github.com/go-git/go-git/v6/x/plugin/config"
)

// TestMain replaces the default ConfigLoader (which reads ~/.gitconfig) with an
// empty one so global settings like commit.gpgSign don't affect test repos
func TestMain(m *testing.M) {
	_ = goplugin.Register(goplugin.ConfigLoader(), func() goplugin.ConfigSource {
		return xconfig.NewEmpty()
	})
	m.Run()
}

func setupTestService(t *testing.T) (*Service, string) {
	t.Helper()
	tempDir := t.TempDir()
	if _, err := git.PlainInit(tempDir, false); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	svc := NewService(tempDir, "https://github.com/test/repo.git", "main", "token", Author{Name: "Test", Email: "test@example.com"}, slog.Default())
	return svc, tempDir
}

func TestService_RepoExists(t *testing.T) {
	tempDir := t.TempDir()

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test", Email: "test@example.com"},
		slog.Default(),
	)

	// Should not exist initially (directory exists but not a git repo)
	// Check if .git directory exists
	gitDir := filepath.Join(tempDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		t.Error("expected .git directory to not exist initially")
	}

	// Create a git repo
	if _, err := git.PlainInit(tempDir, false); err != nil {
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
	service, _ := setupTestService(t)

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
		slog.Default(),
	)

	_, err := service.Open()
	if err == nil {
		t.Error("expected error when opening non-existent repo")
	}
}

func TestService_AssignAuthor(t *testing.T) {
	tempDir := t.TempDir()

	if _, err := git.PlainInit(tempDir, false); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test User", Email: "test@example.com"},
		slog.Default(),
	)

	if err := service.AssignAuthor(); err != nil {
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
	service, tempDir := setupTestService(t)

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

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
	service, tempDir := setupTestService(t)

	files := []string{"test1.txt", "test2.txt", "test3.txt"}
	for _, filename := range files {
		testFile := filepath.Join(tempDir, filename)
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	err := service.Add(files...)
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
	service, tempDir := setupTestService(t)

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

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
	service, _ := setupTestService(t)

	err := service.Add()
	if err == nil {
		t.Error("expected error when adding with no files")
	}
}

func TestService_Add_NonExistentFile(t *testing.T) {
	service, _ := setupTestService(t)

	err := service.Add("nonexistent.txt")
	if err == nil {
		t.Error("expected error when adding non-existent file")
	}
}

func TestService_Remove(t *testing.T) {
	tempDir := t.TempDir()

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
		slog.Default(),
	)

	// Create and commit a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	if _, err := wt.Add("test.txt"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	if _, err := wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	}); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Remove file
	if err := service.Remove("test.txt"); err != nil {
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

	if _, err := git.PlainInit(tempDir, false); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test User", Email: "test@example.com"},
		slog.Default(),
	)

	// Create and add a test file
	if err := os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := service.Add("test.txt"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	if err := service.Commit("test commit"); err != nil {
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
	service, _ := setupTestService(t)

	err := service.Commit("test commit")
	if err == nil {
		t.Error("expected error when committing with no changes")
	}
}

func TestService_CommitAndPush(t *testing.T) {
	tempDir := t.TempDir()

	if _, err := git.PlainInit(tempDir, false); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	service := NewService(
		tempDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		Author{Name: "Test User", Email: "test@example.com"},
		slog.Default(),
	)

	if err := os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := service.Add("test.txt"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	// Commit should succeed; push will fail without a real remote
	err := service.CommitAndPush(t.Context(), "test commit")
	if err != nil {
		// Push error is expected without a real remote; verify commit was created
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
		slog.Default(),
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
