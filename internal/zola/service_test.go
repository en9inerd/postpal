package zola

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/en9inerd/postpal/internal/git"
	gogit "github.com/go-git/go-git/v6"
)

func setupTestService(t *testing.T) (*Service, string) {
	tempDir := t.TempDir()
	postsDir := filepath.Join(tempDir, "content", "posts")
	relPostsDir := "content/posts"
	repoDir := tempDir
	channelID := "@testchannel"

	// Create a git repo for the service
	_, err := gogit.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	gitSvc := git.NewService(
		repoDir,
		"https://github.com/test/repo.git",
		"main",
		"token",
		git.Author{Name: "Test", Email: "test@example.com"},
	)

	service := NewService(postsDir, relPostsDir, repoDir, channelID, gitSvc, "")

	return service, tempDir
}

func createJPEGBytes() []byte {
	return []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
}

func createPNGBytes() []byte {
	return []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
}

func TestService_CreatePost_TextOnly(t *testing.T) {
	service, tempDir := setupTestService(t)
	ctx := context.Background()

	post := Post{
		ID:      123,
		Title:   "@testchannel",
		Content: "Test content",
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	err := service.CreatePost(ctx, post, nil)
	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "content", "posts", "123.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected post file to exist at %s", expectedPath)
	}

	// Verify file content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read post file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "title = \"@testchannel\"") {
		t.Errorf("expected front matter to contain title, got: %s", contentStr)
	}
	if !contains(contentStr, "date = 2024-01-15T10:30:00Z") {
		t.Errorf("expected front matter to contain date, got: %s", contentStr)
	}
	if !contains(contentStr, "Test content") {
		t.Errorf("expected content to contain 'Test content', got: %s", contentStr)
	}
}

func TestService_CreatePost_WithMedia(t *testing.T) {
	service, tempDir := setupTestService(t)
	ctx := context.Background()

	post := Post{
		ID:      456,
		Title:   "@testchannel",
		Content: "Test content with images",
		Date:    time.Date(2024, 2, 20, 15, 45, 0, 0, time.UTC),
	}

	mediaFiles := [][]byte{
		createJPEGBytes(),
		createPNGBytes(),
	}

	err := service.CreatePost(ctx, post, mediaFiles)
	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	postDir := filepath.Join(tempDir, "content", "posts", "456")
	if _, err := os.Stat(postDir); os.IsNotExist(err) {
		t.Errorf("expected post directory to exist at %s", postDir)
	}

	indexPath := filepath.Join(postDir, "index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("expected index.md to exist at %s", indexPath)
	}

	image0Path := filepath.Join(postDir, "image_0.jpg")
	image1Path := filepath.Join(postDir, "image_1.png")
	if _, err := os.Stat(image0Path); os.IsNotExist(err) {
		t.Errorf("expected image_0.jpg to exist")
	}
	if _, err := os.Stat(image1Path); os.IsNotExist(err) {
		t.Errorf("expected image_1.png to exist")
	}

	// Verify front matter contains images
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read post file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, `images = ["image_0.jpg", "image_1.png"]`) {
		t.Errorf("expected front matter to contain images array, got: %s", contentStr)
	}
}

func TestService_EditPost_FindClosestID(t *testing.T) {
	service, tempDir := setupTestService(t)
	_ = context.Background()

	postsDir := filepath.Join(tempDir, "content", "posts")
	// Ensure directory exists
	err := os.MkdirAll(postsDir, 0755)
	if err != nil {
		t.Fatalf("failed to create posts directory: %v", err)
	}

	// Create existing posts: 100, 105, 110
	for _, id := range []int64{100, 105, 110} {
		postPath := filepath.Join(postsDir, fmt.Sprintf("%d.md", id))
		err := os.WriteFile(postPath, []byte("existing post"), 0644)
		if err != nil {
			t.Fatalf("failed to create test post: %v", err)
		}
	}

	editableID, err := service.getEditablePostID(103)
	if err != nil {
		t.Fatalf("getEditablePostID failed: %v", err)
	}

	if editableID != 105 {
		t.Errorf("expected closest post ID to be 105, got %d", editableID)
	}

	editableID, err = service.getEditablePostID(107)
	if err != nil {
		t.Fatalf("getEditablePostID failed: %v", err)
	}

	if editableID != 105 {
		t.Errorf("expected closest post ID to be 105, got %d", editableID)
	}
}

func TestService_EditPost_WithExistingMedia(t *testing.T) {
	service, tempDir := setupTestService(t)
	ctx := context.Background()

	postsDir := filepath.Join(tempDir, "content", "posts")
	postDir := filepath.Join(postsDir, "200")

	if err := os.MkdirAll(postDir, 0755); err != nil {
		t.Fatalf("failed to create post directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(postDir, "image_0.jpg"), createJPEGBytes(), 0644); err != nil {
		t.Fatalf("failed to create image: %v", err)
	}
	if err := os.WriteFile(filepath.Join(postDir, "image_1.jpg"), createJPEGBytes(), 0644); err != nil {
		t.Fatalf("failed to create image: %v", err)
	}

	editPost := Post{
		ID:      200,
		Content: "Updated content",
		Date:    time.Now(),
	}

	if err := service.EditPost(ctx, editPost, nil); err != nil {
		t.Fatalf("EditPost failed: %v", err)
	}

	indexPath := filepath.Join(postDir, "index.md")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read post file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "image_0.jpg") || !contains(contentStr, "image_1.jpg") {
		t.Errorf("expected front matter to contain existing images, got: %s", contentStr)
	}
}

func TestService_EditPost_WithNewMedia(t *testing.T) {
	service, tempDir := setupTestService(t)
	ctx := context.Background()

	postsDir := filepath.Join(tempDir, "content", "posts")
	postDir := filepath.Join(postsDir, "300")

	if err := os.MkdirAll(postDir, 0755); err != nil {
		t.Fatalf("failed to create post directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(postDir, "image_0.jpg"), createJPEGBytes(), 0644); err != nil {
		t.Fatalf("failed to create image: %v", err)
	}

	post := Post{
		ID:      305,
		Content: "Updated content with new image",
		Date:    time.Now(),
	}

	newMedia := createPNGBytes()
	if err := service.EditPost(ctx, post, newMedia); err != nil {
		t.Fatalf("EditPost failed: %v", err)
	}

	newImagePath := filepath.Join(postDir, "image_5.png")
	if _, err := os.Stat(newImagePath); os.IsNotExist(err) {
		t.Errorf("expected new image image_5.png to exist (index = 305 - 300 = 5)")
	}
}

func TestService_DeletePost_TextOnly(t *testing.T) {
	service, tempDir := setupTestService(t)
	ctx := context.Background()

	postsDir := filepath.Join(tempDir, "content", "posts")
	// Ensure directory exists
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatalf("failed to create posts directory: %v", err)
	}

	// Create a text-only post
	postPath := filepath.Join(postsDir, "400.md")
	if err := os.WriteFile(postPath, []byte("test post"), 0644); err != nil {
		t.Fatalf("failed to create test post: %v", err)
	}

	// DeletePost will try to commit, but since files aren't in git, it will fail
	// We'll just verify the files are deleted, not the git operations
	if err := service.DeletePost(ctx, "400"); err != nil && !strings.Contains(err.Error(), "no changes to commit") {
		t.Fatalf("DeletePost failed with unexpected error: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(postPath); err == nil {
		t.Error("expected post file to be deleted")
	}
}

func TestService_DeletePost_WithMedia(t *testing.T) {
	service, tempDir := setupTestService(t)
	ctx := context.Background()

	postsDir := filepath.Join(tempDir, "content", "posts")
	postDir := filepath.Join(postsDir, "500")

	if err := os.MkdirAll(postDir, 0755); err != nil {
		t.Fatalf("failed to create post directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(postDir, "index.md"), []byte("test post"), 0644); err != nil {
		t.Fatalf("failed to create index.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(postDir, "image_0.jpg"), createJPEGBytes(), 0644); err != nil {
		t.Fatalf("failed to create image: %v", err)
	}
	if err := os.WriteFile(filepath.Join(postDir, "image_1.png"), createPNGBytes(), 0644); err != nil {
		t.Fatalf("failed to create image: %v", err)
	}

	if err := service.DeletePost(ctx, "500"); err != nil && !strings.Contains(err.Error(), "no changes to commit") {
		t.Fatalf("DeletePost failed with unexpected error: %v", err)
	}

	if _, err := os.Stat(postDir); err == nil {
		t.Error("expected post directory to be deleted")
	}
}

func TestService_DeletePost_MultipleIDs(t *testing.T) {
	service, tempDir := setupTestService(t)
	ctx := context.Background()

	postsDir := filepath.Join(tempDir, "content", "posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatalf("failed to create posts directory: %v", err)
	}

	for _, id := range []string{"600", "601", "602"} {
		postPath := filepath.Join(postsDir, id+".md")
		if err := os.WriteFile(postPath, []byte("test post"), 0644); err != nil {
			t.Fatalf("failed to create test post: %v", err)
		}
	}

	if err := service.DeletePost(ctx, "600, 601, 602"); err != nil && !strings.Contains(err.Error(), "no changes to commit") {
		t.Fatalf("DeletePost failed with unexpected error: %v", err)
	}

	for _, id := range []string{"600", "601", "602"} {
		postPath := filepath.Join(postsDir, id+".md")
		if _, err := os.Stat(postPath); err == nil {
			t.Errorf("expected post %s to be deleted", id)
		}
	}
}

func TestService_getEditablePostID_FiltersIndex(t *testing.T) {
	service, tempDir := setupTestService(t)

	postsDir := filepath.Join(tempDir, "content", "posts")
	postDir := filepath.Join(postsDir, "700")

	if err := os.MkdirAll(postDir, 0755); err != nil {
		t.Fatalf("failed to create post directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(postDir, "index.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create index.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(postsDir, "701.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create post file: %v", err)
	}

	editableID, err := service.getEditablePostID(702)
	if err != nil {
		t.Fatalf("getEditablePostID failed: %v", err)
	}

	if editableID != 701 {
		t.Errorf("expected closest post ID to be 701, got %d", editableID)
	}
}

func TestService_getPostImageNames(t *testing.T) {
	service, tempDir := setupTestService(t)

	postsDir := filepath.Join(tempDir, "content", "posts")
	postDir := filepath.Join(postsDir, "800")

	if err := os.MkdirAll(postDir, 0755); err != nil {
		t.Fatalf("failed to create post directory: %v", err)
	}

	images := []string{"image_0.jpg", "image_1.png", "image_2.gif"}
	for _, img := range images {
		if err := os.WriteFile(filepath.Join(postDir, img), []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create image: %v", err)
		}
	}

	if err := os.WriteFile(filepath.Join(postDir, "index.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create index.md: %v", err)
	}

	imageNames, err := service.getPostImageNames(800)
	if err != nil {
		t.Fatalf("getPostImageNames failed: %v", err)
	}

	if len(imageNames) != 3 {
		t.Errorf("expected 3 image names, got %d", len(imageNames))
	}

	for _, expected := range images {
		if !slices.Contains(imageNames, expected) {
			t.Errorf("expected image '%s' to be found", expected)
		}
	}
}

func TestService_getImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "JPEG",
			data:     createJPEGBytes(),
			expected: "jpg",
		},
		{
			name:     "PNG",
			data:     createPNGBytes(),
			expected: "png",
		},
		{
			name:     "GIF",
			data:     []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61},
			expected: "gif",
		},
		{
			name:     "WebP",
			data:     []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			expected: "webp",
		},
		{
			name:     "Default for empty",
			data:     []byte{},
			expected: "jpg",
		},
		{
			name:     "Default for unknown",
			data:     []byte{0x00, 0x01, 0x02, 0x03},
			expected: "jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getImageFormat(tt.data)
			if result != tt.expected {
				t.Errorf("expected format '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
