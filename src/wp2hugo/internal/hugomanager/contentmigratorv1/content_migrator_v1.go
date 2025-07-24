package contentmigratorv1

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/gitutils"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func ProcessFile(ctx context.Context, path string, updateInline bool) (*bool, error) {
	// Ignore all non-markdown files
	if !strings.HasSuffix(path, ".md") {
		return lo.ToPtr(false), nil
	}

	// Ignore all non-content paths
	if !strings.Contains(path, "content") {
		return lo.ToPtr(false), nil
	}

	// Ignore files that are already in the right location that is named "index.md" under some directory
	if strings.HasSuffix(path, "/index.md") {
		return lo.ToPtr(false), nil
	}

	// Ignore all files without attachments
	hasAttachments, err := hasAttachments(path)
	if err != nil {
		return lo.ToPtr(false), fmt.Errorf("error checking attachments for file '%s': %w", path, err)
	}

	if !*hasAttachments {
		return lo.ToPtr(false), nil
	}

	log.Debug().
		Str("path", path).
		Bool("updateInline", updateInline).
		Msg("Processing file")

	if updateInline {
		return lo.ToPtr(true), moveFileAndAttachmentIntoSameDir(ctx, path)
	}

	return lo.ToPtr(false), nil
}

func moveFileAndAttachmentIntoSameDir(ctx context.Context, path string) error {
	images, err := getAllImageAttachmentURLs(path)
	if err != nil {
		return fmt.Errorf("error getting image attachment URLs for file '%s': %w", path, err)
	}

	if len(images) == 0 {
		return nil
	}

	log.Debug().
		Str("path", path).
		Strs("images", images).
		Msg("Moving file and attachments into same directory")

	// Step 1: move file "./content/posts/filename.md" -> "./content/posts/filename/index.md"
	blogPostDirPath, err := moveFileToIndexMd(ctx, path)
	if err != nil {
		return fmt.Errorf("error moving file '%s' to index: %w", path, err)
	}
	newBlogPostMdPath := filepath.Join(*blogPostDirPath, "index.md")

	// Step 2: move attachments into "blogPostDirPath"
	for _, image := range images {
		if !strings.HasPrefix(image, "/") {
			log.Debug().
				Str("image", image).
				Msg("Ignoring image attachment which is not an absolute path on the domain")
			continue
		}

		imgFilePath, err := findStaticAttachmentFilePath(path, image)
		if err != nil {
			return fmt.Errorf("error finding image attachment file '%s': %w", image, err)
		}

		newImgFilePath := filepath.Join(*blogPostDirPath, filepath.Base(*imgFilePath))
		if err := gitutils.GitMove(ctx, *imgFilePath, newImgFilePath); err != nil {
			return fmt.Errorf("error moving image attachment file '%s': %w", *imgFilePath, err)
		}

		log.Debug().
			Str("image", image).
			Str("imgFilePath", *imgFilePath).
			Msg("Moved image attachment")
		// Replace the image URL in the markdown file
		if err := replaceImageURLInMarkdownFile(newBlogPostMdPath, image, filepath.Base(*imgFilePath)); err != nil {
			return fmt.Errorf("error replacing image URL in markdown file '%s': %w", path, err)
		}
	}

	return nil
}

func moveFileToIndexMd(ctx context.Context, path string) (*string, error) {
	// Step 1: move file "./content/posts/filename.md" -> "./content/posts/filename/index.md"
	if !strings.HasSuffix(path, ".md") {
		return nil, fmt.Errorf("path '%s' is not a markdown file", path)
	}

	// Get the directory path
	fileNameNoExt := filepath.Base(path[:len(path)-len(filepath.Ext(path))])
	dirPath := filepath.Join(filepath.Dir(path), fileNameNoExt)
	newFilePath := filepath.Join(dirPath, "index.md")
	if utils.FileExists(newFilePath) {
		return nil, fmt.Errorf("file '%s' already exists", newFilePath)
	}

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return nil, fmt.Errorf("error creating directory '%s': %w", dirPath, err)
	}
	log.Debug().
		Str("path", path).
		Str("newFilePath", newFilePath).
		Msg("Moving file to index.md")
	return &dirPath, gitutils.GitMove(ctx, path, newFilePath)
}

func replaceImageURLInMarkdownFile(markdownFilePath string, oldURL, newURL string) error {
	// Replace the image URL in the markdown file
	content, err := os.ReadFile(markdownFilePath)
	if err != nil {
		return fmt.Errorf("error reading markdown file '%s': %w", markdownFilePath, err)
	}

	newContent := strings.ReplaceAll(string(content), oldURL, newURL)
	if err := os.WriteFile(markdownFilePath, []byte(newContent), 0o600); err != nil {
		return fmt.Errorf("error writing markdown file '%s': %w", markdownFilePath, err)
	}

	return nil
}
