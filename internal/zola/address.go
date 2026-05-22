package zola

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type postEntry struct {
	id   int64
	path string
}

// GetLatestAddress returns the most recent hex address found in posts
// and the next incremented address.
func (s *Service) GetLatestAddress() (current, next string, err error) {
	entries, err := os.ReadDir(s.postsDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to read posts directory: %w", err)
	}

	var posts []postEntry
	for _, entry := range entries {
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

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}

		var filePath string
		if entry.IsDir() {
			filePath = filepath.Join(s.postsDir, name, "index.md")
		} else {
			filePath = filepath.Join(s.postsDir, name)
		}

		posts = append(posts, postEntry{id, filePath})
	}

	slices.SortFunc(posts, func(a, b postEntry) int {
		return cmp.Compare(b.id, a.id)
	})

	for _, p := range posts {
		data, err := os.ReadFile(p.path)
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
