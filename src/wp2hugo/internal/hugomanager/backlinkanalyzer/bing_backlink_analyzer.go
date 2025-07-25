package backlinkanalyzer

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/rs/zerolog/log"
)

type BingBacklinkAnalyzer struct {
	backlinksFilepath string
}

type (
	WebsiteURL = string
	Result     map[WebsiteURL]map[BacklinkURL]bool
)

func (r Result) NumReferringPages() int {
	count := 0
	for _, backlinks := range r {
		count += len(backlinks)
	}
	return count
}

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

func (r Result) HostAndFrequency(limit int) ([]Pair[string, int], error) {
	domainFrequency := make(map[string]int)
	for _, backlinks := range r {
		for backlinkURL := range backlinks {
			// Extract the hostname from the backlink URL
			hostname, err := backlinkURL.Host()
			if err != nil {
				return nil, fmt.Errorf("failed to get host from backlink URL %q: %w", backlinkURL, err)
			}

			domainFrequency[*hostname]++
		}
	}

	pairs := make([]Pair[string, int], 0, len(domainFrequency))
	for domain, frequency := range domainFrequency {
		pairs = append(pairs, Pair[string, int]{Key: domain, Value: frequency})
	}

	// Sort the pairs by frequency in descending order
	if limit > 0 && limit < len(pairs) {
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Value > pairs[j].Value
		})
		pairs = pairs[:limit] // Limit the number of results
	}

	return pairs, nil
}

func NewBingBacklinkAnalyzer(bingBacklinksFilepath string) (*BingBacklinkAnalyzer, error) {
	if bingBacklinksFilepath == "" {
		return nil, errors.New("bing backlinks file path is required")
	}

	return &BingBacklinkAnalyzer{
		backlinksFilepath: bingBacklinksFilepath,
	}, nil
}

func (a BingBacklinkAnalyzer) AnalyzeBacklinks() (Result, error) {
	// The file is a CSV file with three columns:
	// 1. Source URL (the backlink)
	// 2. Anchor text (not used)
	// 3. Target URL (the website URL)
	result := make(Result)

	// Open the file and read it line by line
	// For each line, extract the source URL and target URL
	// Add the source URL to the result map under the target URL
	file, err := os.Open(a.backlinksFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backlinks file: %w", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true // Allow lazy quotes in CSV parsing
	firstRecord := true
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break // End of file
		}

		if firstRecord {
			firstRecord = false // Skip the header record
			continue
		}

		if err != nil {
			fmt.Println("Error reading record:", err)
			return nil, fmt.Errorf("failed to read record from backlinks file: %w", err)
		}

		if len(record) < 3 {
			return nil, fmt.Errorf("invalid record: %v, expected at least 3 columns", record)
		}

		if result[record[2]] == nil {
			result[record[2]] = make(map[BacklinkURL]bool, 1)
		}
		backlinkURL, err := BacklinkURL(record[0]).Normalize()
		if err != nil {
			return nil, fmt.Errorf("failed to normalize backlink URL %q: %w", record[0], err)
		}

		if skip, err := backlinkURL.ShouldSkip(); err != nil {
			return nil, fmt.Errorf("failed to check if backlink URL %q should be skipped: %w", *backlinkURL, err)
		} else if *skip {
			log.Trace().
				Str("backlinkURL", string(*backlinkURL)).
				Msg("Skipping backlink URL")
			continue // Skip this backlink URL
		}

		if !result[record[2]][*backlinkURL] {
			log.Debug().
				Any("sourceURL", *backlinkURL).
				Str("targetURL", record[2]).
				Msg("Adding backlink record")
		}
		result[record[2]][*backlinkURL] = true // Add the backlink URL to the target URL's backlinks
	}

	log.Debug().
		Str("backlinksFilepath", a.backlinksFilepath).
		Int("numURLsBacklinked", len(result)).
		Msg("Finished reading backlinks file")
	return result, nil
}
