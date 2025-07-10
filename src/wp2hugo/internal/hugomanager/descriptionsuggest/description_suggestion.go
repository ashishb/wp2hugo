package descriptionsuggest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/frontmatterhelper"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/llmhelper"
	"github.com/openai/openai-go"
	"github.com/rs/zerolog/log"
)

const _seoDescriptionSystemPrompt = "Extract a compelling and SEO-friendly meta description (under 160 characters) from the following" +
	" Markdown blog post. Summarize the key topic concisely while making it engaging for search engines" +
	" and readers. Ensure it is clear, relevant, and encourages clicks. Here is the blog post:"

var (
	ErrFrontMatterHasDescription     = errors.New("front matter already has a description")
	ErrFrontMatterMissingDescription = errors.New("front matter description is missing")
)

func ProcessFile(ctx context.Context, path string, updateInline bool) error {
	log.Trace().
		Str("path", path).
		Msg("Processing file")

	frontMatter, err := frontmatterhelper.GetSelectiveFrontMatter(path)
	if err != nil {
		return fmt.Errorf("error getting front matter: %w", err)
	}

	if frontMatter.HasDescription() {
		return ErrFrontMatterHasDescription
	}

	if !updateInline {
		return ErrFrontMatterMissingDescription
	}

	description, err := suggestDescription(ctx, path)
	if err != nil {
		return fmt.Errorf("error suggesting description: %w", err)
	}

	return frontmatterhelper.UpdateFrontmatter(path, "description", *description)
}

func suggestDescription(ctx context.Context, markdownFilePath string) (*string, error) {
	f, err := os.Open(markdownFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var v any // this is a placeholder for the front matter struct which we don't need to use
	markdownData, err := frontmatter.Parse(f, &v)
	if err != nil {
		return nil, fmt.Errorf("error parsing front matter: %w", err)
	}
	if len(markdownData) == 0 || len(strings.TrimSpace(string(markdownData))) == 0 {
		return nil, errors.New("markdown data is nil")
	}

	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}

	suggestion, err := llmhelper.CallLLM(ctx, openai.ChatModelGPT4o, _seoDescriptionSystemPrompt, string(markdownData))
	if err != nil {
		return nil, fmt.Errorf("error calling LLM: %w", err)
	}

	log.Debug().
		Str("markdownFilePath", markdownFilePath).
		Str("description", *suggestion).
		Msg("Description")
	return suggestion, nil
}
