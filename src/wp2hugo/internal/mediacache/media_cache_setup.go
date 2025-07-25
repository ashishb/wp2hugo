package mediacache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
)

type MediaCache struct {
	cacheDirPath string
}

func New(cacheDirPath string) MediaCache {
	return MediaCache{cacheDirPath: cacheDirPath}
}

func waitOrStop(resp *http.Response) (int, bool) {
	timeout := 1
	stop := false
	switch resp.StatusCode {
	case http.StatusOK:
		// Success
		stop = true

	case http.StatusTooManyRequests:
		// HTTP error 429 = too many requests,
		// aka we are getting rate-thresholded.
		// Some servers may tell us when we are allowed to retry:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil {
				timeout = int(seconds.Seconds())
			}
		} else {
			timeout = 2
		}

	case http.StatusNotFound:
		// HTTP error 404 = not found
		// Useless to retry downloading
		stop = true

	default:
	}

	return timeout, stop
}

func (m MediaCache) GetReader(ctx context.Context, url string) (io.Reader, error) {
	if strings.Contains(url, "blog/blog") {
		log.Panic().
			Str("url", url).
			Msg("media url contains blog/blog")
	}

	if err := utils.CreateDirIfNotExist(m.cacheDirPath); err != nil {
		return nil, fmt.Errorf("error creating cache directory: %w", err)
	}

	key := getSHA256(url)
	file, err := os.OpenFile(path.Join(m.cacheDirPath, key), os.O_RDONLY, 0o644)
	if err == nil {
		log.Info().
			Str("url", url).
			Msg("media found in cache")
		return file, nil
	}

	log.Info().
		Str("url", url).
		Msg("media will be fetched")

	retries := 0
	timeout := 1
	stop := false
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for media %s: %w", url, err)
	}

	var httpErr error
	var resp *http.Response

	for retries < 5 && !stop {
		// Send at most 1 request per second
		// to avoid hammering servers and getting rate-limited.
		time.Sleep(time.Duration(timeout) * time.Second)
		resp, httpErr = http.DefaultClient.Do(req)
		timeout, stop = waitOrStop(resp)
		retries++
		timeout *= retries
	}

	if httpErr != nil {
		return nil, fmt.Errorf("error fetching media %s: %w", url, httpErr)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching media %s: %s", url, resp.Status)
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	file, err = os.OpenFile(path.Join(m.cacheDirPath, key), os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("error creating cache file for media %s: %w", url, err)
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error writing media to cache %s: %w", url, err)
	}

	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("error closing cache file for media %s: %w", url, err)
	}

	file, err = os.OpenFile(path.Join(m.cacheDirPath, key), os.O_RDONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("error opening cache file for media %s: %w", url, err)
	}
	return file, nil
}

func getSHA256(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	return hex.EncodeToString(hasher.Sum(nil))
}
