package backlinkanalyzer

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var _randomAggregators = []string{
	"aerowisatahotels.com", // spam site
	"android.libhunt.com",
	"biggo.com",
	"brandonkaz.com",
	"brutalist.report",
	"cipha.net",
	"discovery.cooperpress.com",
	"duesenklipper.de",
	"geek.ds3783.com",
	"gitea.treehouse.systems",
	"gitlibrary.club",
	"goblgobl.com",
	"hackurls.com",
	"hckrnews.com", // Hacker News clone
	"hnsummary.ai",
	"isthistechdead.com",
	"jimmyr.com",
	"href.ninja",
	"libhunt.com",
	"news.routley.io",
	"news.starmorph.com",
	"news.mcan.sh",        // Hacker News clone
	"old.programming.dev", // Reddit aggregator
	"pandia.org",
	"progscrape.com",
	"programming.dev",       // Reddit aggregator
	"r.genit.al",            // Reddit aggregator
	"reddit.ny4.dev",        // Reddit aggregator
	"ser2.dastresanart.com", // Hacker News clone
	"telegra.ph",
	"uocat.com",
	"upstract.com",
	"vercel.app",
	"webtagr.com",
	"xn--r1a.website",
}

type BacklinkURL string

func (b BacklinkURL) Normalize() (*BacklinkURL, error) {
	u1, err := url.Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("failed to parse backlink URL %q: %w", b, err)
	}

	// Remove "host language" parameter used by Google
	if strings.HasSuffix(u1.Host, "google.com") && u1.Query().Get("hl") != "" {
		q1 := u1.Query()
		q1.Del("hl") // Remove the "hl" parameter
		u1.RawQuery = q1.Encode()
	}

	// Remove "geolocation" parameter used by Google
	if strings.HasSuffix(u1.Host, "google.com") && u1.Query().Get("gl") != "" {
		q1 := u1.Query()
		q1.Del("gl") // Remove the "hl" parameter
		u1.RawQuery = q1.Encode()
	}

	// Normalize the language for "f-droid.org/<langCode>/..." -> "f-droid.org/..."
	if strings.HasSuffix(u1.Host, "f-droid.org") {
		// Regex to match a-Z and underscore
		r1 := regexp.MustCompile("[a-zA-Z_]{2,100}")
		fields := strings.Split(strings.TrimPrefix(u1.Path, "/"), "/")
		log.Trace().
			Str("path", u1.Path).
			Str("fields[0]", fields[0]).
			Msg("BacklinkURL: fields")
		if len(fields) > 1 && r1.MatchString(fields[0]) {
			// If the first part of the path is a language code, remove it
			u1.Path = strings.TrimPrefix(u1.Path, "/"+fields[0])
		}
	}

	if u1.Host == "signup.tldr.tech" {
		u1.Host = "tldr.tech"
	}

	return lo.ToPtr(BacklinkURL(u1.String())), nil
}

func (b BacklinkURL) ShouldSkip() (*bool, error) {
	u1, err := url.Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("failed to parse backlink URL %q: %w", b, err)
	}

	// Skip bad domains
	if slices.Contains(_randomAggregators, strings.TrimPrefix(u1.Host, "www.")) {
		return lo.ToPtr(true), nil
	}

	// low-quality Hacker news copies
	if strings.HasSuffix(u1.Host, ".netlify.app") {
		return lo.ToPtr(true), nil
	}

	// Skip if path seems to imply some sort of list page
	if strings.Contains(u1.Path, "/tag/") ||
		strings.Contains(u1.Path, "/tags/") ||
		strings.Contains(u1.Path, "/page/") ||
		strings.Contains(u1.Path, "/k/") {
		return lo.ToPtr(true), nil
	}

	if u1.Query().Has("page") {
		value := u1.Query().Get("page")
		_, err := strconv.Atoi(value) // Check if the value is a number
		if err == nil {
			// If the value is a number then this is a random page of blog listings
			// and not a blog post, skip this URL
			return lo.ToPtr(true), nil
		}
	}

	// Alternatively, p%d
	r1 := regexp.MustCompile(`/p\d+/`)
	if r1.MatchString(u1.Path) {
		// If the path contains /p<number>/, this is likely a paginated list of blog posts
		// and not a blog post itself, skip this URL
		log.Trace().
			Str("path", u1.Path).
			Msg("BacklinkURL: skipping paginated list URL")
		return lo.ToPtr(true), nil
	}

	// Skip if path contains "/dislike_post/", this seems to be some sort of a low-quality aggregator
	if strings.Contains(u1.Path, "/like_post/") || strings.Contains(u1.Path, "/dislike_post/") {
		return lo.ToPtr(true), nil
	}

	// Skip Lobste.rs comments
	if strings.HasPrefix(u1.Path, "/c/") && strings.HasSuffix(u1.Host, "lobste.rs") {
		return lo.ToPtr(true), nil
	}

	// Skip https://lobste.rs/domains/<domain>
	if strings.HasPrefix(u1.Path, "/domains/") && strings.HasSuffix(u1.Host, "lobste.rs") {
		return lo.ToPtr(true), nil
	}

	// Skip rss feeds
	if strings.Contains(u1.Path, "/feed.xml") {
		return lo.ToPtr(true), nil
	}

	// Skip rss feeds
	if strings.HasSuffix(u1.Path, ".rss") {
		return lo.ToPtr(true), nil
	}

	// Low-quality links like hn.cho.sh, https://hn-tldr.com/posts/44214835 and https://hn-grep.rednafi.com/dev/test.php
	if strings.Contains(u1.Host, "hn.") ||
		strings.Contains(u1.Host, "hn-") ||
		strings.Contains(u1.Host, "links-yc.") {
		return lo.ToPtr(true), nil
	}

	if u1.Host == "play.google.com" || u1.Path == "/store/apps/details" {
		// Skip Google Play Store links
		return lo.ToPtr(true), nil
	}

	return lo.ToPtr(false), nil
}

func (b BacklinkURL) Host() (*string, error) {
	u1, err := url.Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("failed to parse backlink URL %q: %w", b, err)
	}

	if u1.Host == "" {
		return nil, fmt.Errorf("backlink URL %q has no host", b)
	}

	return lo.ToPtr(strings.TrimPrefix(u1.Host, "www.")), nil
}
