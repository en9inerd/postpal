package zola

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/en9inerd/postpal/internal/git"
)

// Service handles Zola blog post creation and management
type Service struct {
	postsDir        string
	relPostsDir     string
	repoDir         string
	channelID       string
	gitService      *git.Service
	exportedDataDir string
}

// NewService creates a new Zola post service
func NewService(postsDir, relPostsDir, repoDir, channelID string, gitService *git.Service, exportedDataDir string) *Service {
	return &Service{
		postsDir:        postsDir,
		relPostsDir:     relPostsDir,
		repoDir:         repoDir,
		channelID:       channelID,
		gitService:      gitService,
		exportedDataDir: exportedDataDir,
	}
}

// CreatePost creates a new Zola blog post from a Post struct and media files
func (s *Service) CreatePost(ctx context.Context, post Post, mediaFiles [][]byte) error {
	imageNames := make([]string, len(mediaFiles))
	for i := range mediaFiles {
		format := getImageFormat(mediaFiles[i])
		imageNames[i] = fmt.Sprintf("image_%d.%s", i, format)
	}
	post.ImageNames = imageNames

	var filename string
	var postFilePath string

	if len(post.ImageNames) > 0 {
		// Create directory for post with media
		postDir := filepath.Join(s.postsDir, strconv.FormatInt(post.ID, 10))
		if err := os.MkdirAll(postDir, 0755); err != nil {
			return fmt.Errorf("failed to create post directory: %w", err)
		}
		filename = filepath.Join(strconv.FormatInt(post.ID, 10), "index.md")
		postFilePath = filepath.Join(s.postsDir, filename)
	} else {
		if err := os.MkdirAll(s.postsDir, 0755); err != nil {
			return fmt.Errorf("failed to create posts directory: %w", err)
		}
		filename = strconv.FormatInt(post.ID, 10) + ".md"
		postFilePath = filepath.Join(s.postsDir, filename)
	}

	processedContent := ProcessContent(post.Content)
	if post.Title == "" {
		post.Title = ExtractTitle(post.Content, s.channelID)
	}
	processedContent = RemoveAddressPattern(processedContent)

	frontMatter := BuildFrontMatter(post)
	postContent := frontMatter + processedContent + "\n"

	if err := os.WriteFile(postFilePath, []byte(postContent), 0644); err != nil {
		return fmt.Errorf("failed to write post file: %w", err)
	}

	relPostPath := filepath.Join(s.relPostsDir, filename)
	if err := s.gitService.Add(relPostPath); err != nil {
		return fmt.Errorf("failed to add post file to git: %w", err)
	}

	for i, mediaFile := range mediaFiles {
		imageFilename := post.ImageNames[i]
		imagePath := filepath.Join(s.postsDir, strconv.FormatInt(post.ID, 10), imageFilename)
		relImagePath := filepath.Join(s.relPostsDir, strconv.FormatInt(post.ID, 10), imageFilename)

		if err := os.WriteFile(imagePath, mediaFile, 0644); err != nil {
			return fmt.Errorf("failed to write image file: %w", err)
		}

		if err := s.gitService.Add(relImagePath); err != nil {
			return fmt.Errorf("failed to add image file to git: %w", err)
		}
	}

	return nil
}

// EditPost edits an existing post, finding the closest post ID
func (s *Service) EditPost(ctx context.Context, post Post, mediaFile []byte) error {
	originalPostID := post.ID

	editablePostID, err := s.getEditablePostID(post.ID)
	if err != nil {
		return fmt.Errorf("failed to find editable post: %w", err)
	}

	imageNames, err := s.getPostImageNames(editablePostID)
	if err != nil {
		return fmt.Errorf("failed to get post image names: %w", err)
	}

	numOfMediaFiles := len(imageNames)
	post.ID = editablePostID

	if post.Content != "" {
		if numOfMediaFiles > 0 {
			if len(post.ImageNames) > 0 {
				firstImageName := post.ImageNames[0]
				parts := strings.Split(firstImageName, ".")
				format := "jpg"
				if len(parts) > 1 {
					format = parts[len(parts)-1]
				}
				post.ImageNames = make([]string, numOfMediaFiles)
				for i := range post.ImageNames {
					post.ImageNames[i] = fmt.Sprintf("image_%d.%s", i, format)
				}
			} else {
				post.ImageNames = imageNames
			}
		}

		processedContent := ProcessContent(post.Content)
		if post.Title == "" {
			post.Title = ExtractTitle(post.Content, s.channelID)
		}
		processedContent = RemoveAddressPattern(processedContent)

		var filename string
		var postFilePath string

		if numOfMediaFiles > 0 {
			postDir := filepath.Join(s.postsDir, strconv.FormatInt(editablePostID, 10))
			if err := os.MkdirAll(postDir, 0755); err != nil {
				return fmt.Errorf("failed to create post directory: %w", err)
			}
			filename = filepath.Join(strconv.FormatInt(editablePostID, 10), "index.md")
			postFilePath = filepath.Join(s.postsDir, filename)
		} else {
			filename = strconv.FormatInt(editablePostID, 10) + ".md"
			postFilePath = filepath.Join(s.postsDir, filename)
		}

		frontMatter := BuildFrontMatter(post)
		postContent := frontMatter + processedContent + "\n"

		if err := os.WriteFile(postFilePath, []byte(postContent), 0644); err != nil {
			return fmt.Errorf("failed to write post file: %w", err)
		}

		relPostPath := filepath.Join(s.relPostsDir, filename)
		if err := s.gitService.Add(relPostPath); err != nil {
			return fmt.Errorf("failed to add post file to git: %w", err)
		}
	}

	if mediaFile != nil {
		index := originalPostID - editablePostID
		format := getImageFormat(mediaFile)
		imageFilename := fmt.Sprintf("image_%d.%s", index, format)
		imagePath := filepath.Join(s.postsDir, strconv.FormatInt(editablePostID, 10), imageFilename)
		relImagePath := filepath.Join(s.relPostsDir, strconv.FormatInt(editablePostID, 10), imageFilename)

		postDir := filepath.Join(s.postsDir, strconv.FormatInt(editablePostID, 10))
		if err := os.MkdirAll(postDir, 0755); err != nil {
			return fmt.Errorf("failed to create post directory: %w", err)
		}

		if err := os.WriteFile(imagePath, mediaFile, 0644); err != nil {
			return fmt.Errorf("failed to write image file: %w", err)
		}

		if err := s.gitService.Add(relImagePath); err != nil {
			return fmt.Errorf("failed to add image file to git: %w", err)
		}
	}

	return nil
}

// DeletePost deletes one or more posts (comma-separated IDs)
func (s *Service) DeletePost(ctx context.Context, ids string) error {
	idList := strings.Split(ids, ",")

	for _, idStr := range idList {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		postID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid post ID: %s", idStr)
		}

		imageNames, err := s.getPostImageNames(postID)
		if err != nil {
			continue
		}

		if len(imageNames) > 0 {
			postDir := filepath.Join(s.postsDir, strconv.FormatInt(postID, 10))
			if err := os.RemoveAll(postDir); err != nil {
				return fmt.Errorf("failed to remove post directory: %w", err)
			}

			relPostDir := filepath.Join(s.relPostsDir, strconv.FormatInt(postID, 10))
			_ = s.gitService.Remove(filepath.Join(relPostDir, "index.md"))

			for _, imageName := range imageNames {
				_ = s.gitService.Remove(filepath.Join(relPostDir, imageName))
			}
		} else {
			filename := strconv.FormatInt(postID, 10) + ".md"
			postFilePath := filepath.Join(s.postsDir, filename)
			_ = os.Remove(postFilePath)

			relPostPath := filepath.Join(s.relPostsDir, filename)
			_ = s.gitService.Remove(relPostPath)
		}
	}

	commitMsg := fmt.Sprintf("Delete post(s): %s", ids)
	if err := s.gitService.CommitAndPush(ctx, commitMsg); err != nil {
		return fmt.Errorf("failed to commit and push deletion: %w", err)
	}

	return nil
}

// getEditablePostID finds the closest existing post ID to the given ID
func (s *Service) getEditablePostID(postID int64) (int64, error) {
	entries, err := os.ReadDir(s.postsDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read posts directory: %w", err)
	}

	var postIDs []int64
	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, "index") {
			continue
		}

		var id int64
		if strings.Contains(name, ".") {
			idStr := strings.Split(name, ".")[0]
			var err error
			id, err = strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				continue
			}
		} else {
			var err error
			id, err = strconv.ParseInt(name, 10, 64)
			if err != nil {
				continue
			}
		}
		postIDs = append(postIDs, id)
	}

	if len(postIDs) == 0 {
		return postID, nil
	}

	closestID := postIDs[0]
	minDiff := abs(postID - closestID)

	for _, id := range postIDs[1:] {
		diff := abs(postID - id)
		if diff < minDiff {
			minDiff = diff
			closestID = id
		}
	}

	return closestID, nil
}

// getPostImageNames returns the list of image file names for a post
func (s *Service) getPostImageNames(postID int64) ([]string, error) {
	postDir := filepath.Join(s.postsDir, strconv.FormatInt(postID, 10))

	entries, err := os.ReadDir(postDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read post directory: %w", err)
	}

	var imageNames []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "image_") {
			imageNames = append(imageNames, name)
		}
	}

	sort.Slice(imageNames, func(i, j int) bool {
		return imageNames[i] < imageNames[j]
	})

	return imageNames, nil
}

func getImageFormat(data []byte) string {
	if len(data) < 4 {
		return "jpg"
	}

	if len(data) >= 4 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpg"
	}
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}
	if len(data) >= 6 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "gif"
	}
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		if len(data) >= 12 && string(data[8:12]) == "WEBP" {
			return "webp"
		}
	}

	return "jpg"
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
