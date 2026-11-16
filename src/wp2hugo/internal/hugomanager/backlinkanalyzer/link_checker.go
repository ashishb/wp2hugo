package backlinkanalyzer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	_linkCheckTimeout     = 30 * time.Second
	_linkCheckConcurrency = 16
	_linkCheckUserAgent   = "ashishb/wp2hugo"
	_maxRedirects         = 10
)

var (
	errRedirectLoop     = errors.New("redirect loop detected")
	errTooManyRedirects = errors.New("too many redirects")
)

// _noRedirectClient does not follow redirects automatically so that we can
// follow them manually (up to _maxRedirects) and record the final URL.
var _noRedirectClient = &http.Client{
	Timeout: _linkCheckTimeout,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type linkCheckResult struct {
	finalURL   string
	statusCode int
	redirected bool
}

// CheckLinksAlive checks every backlink in the result and logs a warning
// for every link that does not return HTTP 200.
// Redirects are followed (up to _maxRedirects) until a non-redirect response is received,
// and the final URL is logged.
func (r Result) CheckLinksAlive(ctx context.Context) {
	// Collect unique backlinks along with the target URLs they point to
	backlinks := make(map[BacklinkURL][]WebsiteURL)
	for targetURL, sources := range r {
		for backlink := range sources {
			backlinks[backlink] = append(backlinks[backlink], targetURL)
		}
	}

	log.Info().
		Int("numBacklinks", len(backlinks)).
		Msg("Checking whether backlinks are alive")

	work := make(chan BacklinkURL)
	var wg sync.WaitGroup
	for range _linkCheckConcurrency {
		wg.Go(func() {
			for backlink := range work {
				checkAndLogLink(ctx, backlink, backlinks[backlink])
			}
		})
	}

	for backlink := range backlinks {
		work <- backlink
	}
	close(work)
	wg.Wait()
}

func checkAndLogLink(ctx context.Context, backlink BacklinkURL, targetURLs []WebsiteURL) {
	result, err := checkLink(ctx, string(backlink))
	if err != nil {
		event := log.Warn().
			Err(err).
			Str("backlinkURL", string(backlink)).
			Strs("targetURLs", targetURLs)
		if result != nil && result.redirected {
			event = event.Str("finalURL", result.finalURL)
		}
		event.Msg("Backlink is not reachable")
		return
	}

	if result.statusCode == http.StatusOK && !result.redirected {
		log.Trace().
			Str("backlinkURL", string(backlink)).
			Msg("Backlink is alive")
		return
	}

	event := log.Warn().
		Str("backlinkURL", string(backlink)).
		Strs("targetURLs", targetURLs).
		Int("statusCode", result.statusCode)
	if result.redirected {
		event.
			Str("finalURL", result.finalURL).
			Msg("Backlink redirects")
	} else {
		event.Msg("Backlink did not return 200")
	}
}

// checkLink fetches the given URL and follows up to _maxRedirects redirects
// (stopping early if a redirect loop is detected).
func checkLink(ctx context.Context, link string) (*linkCheckResult, error) {
	result := &linkCheckResult{finalURL: link}
	visited := make(map[string]bool)
	for numRedirects := 0; ; numRedirects++ {
		if visited[result.finalURL] {
			return result, fmt.Errorf("%w at %q", errRedirectLoop, result.finalURL)
		}
		visited[result.finalURL] = true

		statusCode, location, err := fetchOnce(ctx, result.finalURL)
		if err != nil {
			return result, err
		}

		result.statusCode = statusCode
		if !isRedirect(statusCode) {
			return result, nil
		}

		if location == "" {
			return result, fmt.Errorf("redirect status %d from %q without Location header", statusCode, result.finalURL)
		}

		if numRedirects >= _maxRedirects {
			return result, fmt.Errorf("%w: stopped after %d redirects at %q", errTooManyRedirects, _maxRedirects, result.finalURL)
		}

		nextURL, err := resolveLocation(result.finalURL, location)
		if err != nil {
			return result, err
		}

		log.Trace().
			Str("from", result.finalURL).
			Str("to", nextURL).
			Int("statusCode", statusCode).
			Msg("Following redirect")
		result.finalURL = nextURL
		result.redirected = true
	}
}

func fetchOnce(ctx context.Context, link string) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request for %q: %w", link, err)
	}

	req.Header.Set("User-Agent", _linkCheckUserAgent)
	resp, err := _noRedirectClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to fetch %q: %w", link, err)
	}

	defer resp.Body.Close()
	// Drain (a bounded amount of) the body so the connection can be reused
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, resp.Header.Get("Location"), nil
}

func isRedirect(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

func resolveLocation(base string, location string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %q: %w", base, err)
	}

	locationURL, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("failed to parse redirect location %q: %w", location, err)
	}

	return baseURL.ResolveReference(locationURL).String(), nil
}
