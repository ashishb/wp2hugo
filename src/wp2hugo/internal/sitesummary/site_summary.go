package sitesummary

import (
	"os"
	"path/filepath"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/urlsuggest"
	"github.com/rs/zerolog/log"
)

type SiteSummary struct {
	Posts  int // Number of articles including drafts
	Drafts int // Number of drafts
	Future int // Number of posts with future dates
	// Add to be published posts
}

func ScanDir(dir string) (*SiteSummary, error) {
	var summary SiteSummary

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Process file
		matter, err := urlsuggest.GetSelectiveFrontMatter(path)
		if err != nil {
			return err
		}
		summary.Posts++
		if matter.IsDraft() {
			summary.Drafts++
		}
		if inFuture, err := matter.IsInFuture(); err == nil && inFuture {
			summary.Future++
		} else if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("Error checking if post is in future")
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &summary, nil
}
