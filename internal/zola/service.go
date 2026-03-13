package zola

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

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

// CreatePost creates a new Zola blog post from a Post struct and media files
func (s *Service) CreatePost(ctx context.Context, post Post, mediaFiles [][]byte) error {
	imageNames := make([]string, len(mediaFiles))
	for i := range mediaFiles {
		format := getImageFormat(mediaFiles[i])
		imageNames[i] = fmt.Sprintf("image_%d.%s", i, format)
	}
	post.ImageNames = imageNames

	postIDStr := strconv.FormatInt(post.ID, 10)

	var filename string
	var postFilePath string

	if len(post.ImageNames) > 0 {
		postDir := filepath.Join(s.postsDir, postIDStr)
		if err := os.MkdirAll(postDir, 0755); err != nil {
			return fmt.Errorf("failed to create post directory: %w", err)
		}
		filename = filepath.Join(postIDStr, "index.md")
		postFilePath = filepath.Join(s.postsDir, filename)
	} else {
		if err := os.MkdirAll(s.postsDir, 0755); err != nil {
			return fmt.Errorf("failed to create posts directory: %w", err)
		}
		filename = postIDStr + ".md"
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
		imagePath := filepath.Join(s.postsDir, postIDStr, imageFilename)
		relImagePath := filepath.Join(s.relPostsDir, postIDStr, imageFilename)

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

	editablePostIDStr := strconv.FormatInt(editablePostID, 10)

	imageNames, err := s.getPostImageNames(editablePostID)
	if err != nil {
		return fmt.Errorf("failed to get post image names: %w", err)
	}

	numOfMediaFiles := len(imageNames)
	post.ID = editablePostID

	// Check if we're converting from text-only to album (adding media to a post that had none)
	convertingToAlbum := numOfMediaFiles == 0 && mediaFile != nil

	s.logger.Debug("edit post state",
		"original_post_id", originalPostID,
		"editable_post_id", editablePostID,
		"num_media_files", numOfMediaFiles,
		"has_media_file", mediaFile != nil,
		"has_content", post.Content != "",
		"converting_to_album", convertingToAlbum)

	// If converting to album without content update, read existing content first.
	// Content from file is already processed, so we skip ProcessContent later.
	contentFromFile := false
	if convertingToAlbum && post.Content == "" {
		oldFilename := editablePostIDStr + ".md"
		oldFilePath := filepath.Join(s.postsDir, oldFilename)
		existingContent, err := os.ReadFile(oldFilePath)
		if err == nil {
			// Extract content after front matter (after second +++)
			content := string(existingContent)
			parts := strings.SplitN(content, "+++", 3)
			if len(parts) >= 3 {
				post.Content = strings.TrimSpace(parts[2])
				contentFromFile = true
			}
		}
	}

	if post.Content != "" || convertingToAlbum {
		if numOfMediaFiles > 0 || convertingToAlbum {
			if len(post.ImageNames) > 0 {
				firstImageName := post.ImageNames[0]
				parts := strings.Split(firstImageName, ".")
				format := "jpg"
				if len(parts) > 1 {
					format = parts[len(parts)-1]
				}
				if convertingToAlbum {
					// New album with one image
					post.ImageNames = []string{fmt.Sprintf("image_0.%s", format)}
				} else {
					post.ImageNames = make([]string, numOfMediaFiles)
					for i := range post.ImageNames {
						post.ImageNames[i] = fmt.Sprintf("image_%d.%s", i, format)
					}
				}
			} else if convertingToAlbum {
				format := getImageFormat(mediaFile)
				post.ImageNames = []string{fmt.Sprintf("image_0.%s", format)}
			} else {
				post.ImageNames = imageNames
			}
		}

		processedContent := post.Content
		if post.Title == "" {
			post.Title = ExtractTitle(post.Content, s.channelID)
		}
		if !contentFromFile {
			processedContent = ProcessContent(post.Content)
			processedContent = RemoveAddressPattern(processedContent)
		}

		var filename string
		var postFilePath string

		if numOfMediaFiles > 0 || convertingToAlbum {
			postDir := filepath.Join(s.postsDir, editablePostIDStr)
			if err := os.MkdirAll(postDir, 0755); err != nil {
				return fmt.Errorf("failed to create post directory: %w", err)
			}
			filename = filepath.Join(editablePostIDStr, "index.md")
			postFilePath = filepath.Join(s.postsDir, filename)

			// If converting to album, remove the old .md file
			if convertingToAlbum {
				oldFilename := editablePostIDStr + ".md"
				oldFilePath := filepath.Join(s.postsDir, oldFilename)
				_ = os.Remove(oldFilePath)
				relOldPath := filepath.Join(s.relPostsDir, oldFilename)
				_ = s.gitService.Remove(relOldPath)
			}
		} else {
			filename = editablePostIDStr + ".md"
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
		imagePath := filepath.Join(s.postsDir, editablePostIDStr, imageFilename)
		relImagePath := filepath.Join(s.relPostsDir, editablePostIDStr, imageFilename)

		postDir := filepath.Join(s.postsDir, editablePostIDStr)
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
	deleted := false

	for idStr := range strings.SplitSeq(ids, ",") {
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

		postIDStr := strconv.FormatInt(postID, 10)

		if len(imageNames) > 0 {
			// Album post - delete directory
			postDir := filepath.Join(s.postsDir, postIDStr)
			if err := os.RemoveAll(postDir); err != nil {
				return fmt.Errorf("failed to remove post directory: %w", err)
			}

			relPostDir := filepath.Join(s.relPostsDir, postIDStr)
			_ = s.gitService.Remove(filepath.Join(relPostDir, "index.md"))

			for _, imageName := range imageNames {
				_ = s.gitService.Remove(filepath.Join(relPostDir, imageName))
			}
			deleted = true
		} else {
			// Text-only post - delete .md file
			filename := postIDStr + ".md"
			postFilePath := filepath.Join(s.postsDir, filename)
			if _, err := os.Stat(postFilePath); err == nil {
				_ = os.Remove(postFilePath)
				relPostPath := filepath.Join(s.relPostsDir, filename)
				_ = s.gitService.Remove(relPostPath)
				deleted = true
			}
		}
	}

	if !deleted {
		return ErrNoPostsDeleted
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

		idStr, _, _ := strings.Cut(name, ".")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		postIDs = append(postIDs, id)
	}

	if len(postIDs) == 0 {
		return postID, nil
	}

	// Prefer the largest ID <= postID. Album message IDs are always >= the
	// first message's ID, so looking backward finds the correct album post
	// even when a newer post is numerically closer.
	slices.Sort(postIDs)
	closestID := postIDs[0]
	for _, id := range postIDs {
		if id <= postID {
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

	slices.Sort(imageNames)

	return imageNames, nil
}

func getImageFormat(data []byte) string {
	if len(data) < 4 {
		return "jpg"
	}

	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "gif"
	}
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 && string(data[8:12]) == "WEBP" {
		return "webp"
	}

	return "jpg"
}

// GetLatestAddress returns the most recent hex address found in posts
// and the next incremented address.
func (s *Service) GetLatestAddress() (current, next string, err error) {
	entries, err := os.ReadDir(s.postsDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to read posts directory: %w", err)
	}

	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		name := entry.Name()

		var idStr string
		if entry.IsDir() {
			idStr = name
		} else if before, ok := strings.CutSuffix(name, ".md"); ok {
			idStr = before
		} else {
			continue
		}

		if strings.HasPrefix(idStr, "_") {
			continue
		}

		if _, err := strconv.ParseInt(idStr, 10, 64); err != nil {
			continue
		}

		var filePath string
		if entry.IsDir() {
			filePath = filepath.Join(s.postsDir, name, "index.md")
		} else {
			filePath = filepath.Join(s.postsDir, name)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		content := string(data)
		parts := strings.SplitN(content, "+++", 3)
		if len(parts) < 3 {
			continue
		}

		frontMatter := parts[1]
		for line := range strings.SplitSeq(frontMatter, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "title") {
				_, value, found := strings.Cut(line, "=")
				if !found {
					continue
				}
				value = strings.TrimSpace(value)
				value = strings.Trim(value, "\"")

				start := strings.LastIndex(value, "[0x")
				end := strings.LastIndex(value, "]")
				if start >= 0 && end > start {
					current = value[start+1 : end]
				}
				break
			}
		}

		if current != "" {
			break
		}
	}

	if current == "" {
		return "", "", fmt.Errorf("no address found in any post")
	}

	hexStr := current[2:]
	addrVal, err := strconv.ParseUint(hexStr, 16, 64)
	if err != nil {
		return current, "", fmt.Errorf("failed to parse address %s: %w", current, err)
	}

	nextVal := addrVal + 1
	format := "0x%0*x"
	if strings.ToUpper(hexStr) == hexStr {
		format = "0x%0*X"
	}
	next = fmt.Sprintf(format, len(hexStr), nextVal)

	return current, next, nil
}

// SaveChannelLogo saves the channel logo to the static directory
func (s *Service) SaveChannelLogo(logoData []byte) error {
	format := getImageFormat(logoData)
	logoPath := filepath.Join(s.repoDir, "static", "logo."+format)

	staticDir := filepath.Join(s.repoDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("failed to create static directory: %w", err)
	}

	if err := os.WriteFile(logoPath, logoData, 0644); err != nil {
		return fmt.Errorf("failed to write logo file: %w", err)
	}

	relLogoPath := filepath.Join("static", "logo."+format)
	if err := s.gitService.Add(relLogoPath); err != nil {
		return fmt.Errorf("failed to add logo to git: %w", err)
	}

	return nil
}
