package descriptionsuggest

import (
	"errors"
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/frontmatterhelper"
	"github.com/rs/zerolog/log"
)

var (
	ErrFrontMatterMissingDescription = errors.New("front matter description is missing")
)

func ProcessFile(path string, updateInline bool) error {
	log.Trace().
		Str("path", path).
		Msg("Processing file")

	frontMatter, err := frontmatterhelper.GetSelectiveFrontMatter(path)
	if err != nil {
		return fmt.Errorf("error getting front matter: %w", err)
	}

	if frontMatter.HasDescription() {
		return nil
	}

	// TODO: Implement description suggestion logic
	// For now, we just return an error
	return ErrFrontMatterMissingDescription
}
