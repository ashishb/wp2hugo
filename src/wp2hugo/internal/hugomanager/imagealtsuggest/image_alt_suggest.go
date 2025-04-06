package imagealtsuggest

import (
	"context"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
)

// Parse "{{< figure align=aligncenter width=768 src="Cedar_trail_waterfall-768x1024.jpg" alt="" >}}"
// and extract "src" and "alt" attributes using regular expressions
var _figureShortCodeRegEx = regexp.MustCompile(`{{<\s*?figure.*?>\s*}}`)

type Result struct {
	numImageWithAlt    int
	numImageMissingAlt int
	numImageUpdated    int
}

func (r Result) NumImageWithAlt() int {
	return r.numImageWithAlt
}

func (r Result) NumImageMissingAlt() int {
	return r.numImageMissingAlt
}

func (r Result) NumImageUpdated() int {
	return r.numImageUpdated
}

func ProcessFile(ctx context.Context, path string, updateInline bool) (*Result, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	defer func() {
		_ = f.Close()
	}()
	var frontmatterData any
	log.Debug().
		Str("path", path).
		Msg("Parsing frontmatter")
	mdBody, err := frontmatter.Parse(f, &frontmatterData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in file %s: %w", path, err)
	}
	if len(mdBody) == 0 {
		log.Debug().
			Str("path", path).
			Msg("No markdown body found, skipping")
		return &Result{}, nil
	}

	// Check if the body contains images
	// and if they are missing alt text
	matches := _figureShortCodeRegEx.FindAll(mdBody, -1)
	if len(matches) == 0 {
		return &Result{
			numImageWithAlt:    0,
			numImageMissingAlt: 0,
			numImageUpdated:    0,
		}, nil
	}

	numImageWithAlt := 0
	for _, match := range matches {
		if len(_figureShortCodeRegEx.FindSubmatch(match)) > 2 &&
			len(_figureShortCodeRegEx.FindSubmatch(match)[2]) > 0 {
			numImageWithAlt++
		}
	}
	numImageMissingAlt := len(matches) - numImageWithAlt
	if updateInline {
		return nil, fmt.Errorf("inline update not supported yet")
	}

	return &Result{
		numImageWithAlt:    numImageWithAlt,
		numImageMissingAlt: numImageMissingAlt,
		numImageUpdated:    0,
	}, nil
}
