package sitesummary

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/hugomanager/urlsuggest"
	"github.com/mergestat/timediff"
	"github.com/rs/zerolog/log"
)

type PostAndTime struct {
	Path string
	Time string
}

func (p PostAndTime) String() string {
	if p.Time == "" {
		return p.Path
	}
	return fmt.Sprintf("%s (%s)", p.Path, p.Time)
}

func (p PostAndTime) RelativeTime() string {
	if p.Time == "" {
		return ""
	}
	t1, err := time.Parse(time.RFC3339, p.Time)
	if err != nil {
		panic(err)
	}
	return timediff.TimeDiff(t1)
}

type SiteSummary struct {
	numPosts        int // Number of articles including drafts
	numDrafts       int // Number of drafts
	numFuture       int // Number of posts with future dates
	draftPostPaths  []PostAndTime
	futurePostPaths []PostAndTime
}

func (s *SiteSummary) Posts() int {
	return s.numPosts
}

func (s *SiteSummary) Drafts() int {
	return s.numDrafts
}

func (s *SiteSummary) Future() int {
	return s.numFuture
}

func (s *SiteSummary) DraftPostPaths(limit int) []PostAndTime {
	sort.Slice(s.draftPostPaths, func(i, j int) bool {
		return s.draftPostPaths[i].Time < s.draftPostPaths[j].Time
	})
	if limit > 0 && limit < len(s.draftPostPaths) {
		return s.draftPostPaths[:limit]
	}
	return s.draftPostPaths
}

func (s *SiteSummary) FuturePostPaths(limit int) []PostAndTime {
	sort.Slice(s.futurePostPaths, func(i, j int) bool {
		return s.futurePostPaths[i].Time < s.futurePostPaths[j].Time
	})
	if limit > 0 && limit < len(s.futurePostPaths) {
		return s.futurePostPaths[:limit]
	}
	return s.futurePostPaths
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
		postInfo, err := urlsuggest.GetSelectiveFrontMatter(path)
		if err != nil {
			return err
		}
		summary.numPosts++
		if postInfo.IsDraft() {
			summary.numDrafts++
			summary.draftPostPaths = append(summary.draftPostPaths, PostAndTime{Path: path})
		}
		inFuture, err := postInfo.IsInFuture()
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("Error checking if post is in future")
			return err
		}
		if inFuture {
			summary.numFuture++
			summary.futurePostPaths = append(summary.futurePostPaths,
				PostAndTime{Path: path, Time: postInfo.PublishDate})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &summary, nil
}
