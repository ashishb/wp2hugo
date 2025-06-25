package mediacache

import (
	"crypto/sha256"
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

	case 200:
		// Success
		stop = true

	case 429:
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

	case 404:
		// HTTP error 404 = not found
		// Useless to retry downloading
		stop = true

	default:
		
	}

	return timeout, stop
}

func (m MediaCache) GetReader(url string) (io.Reader, error) {
	if strings.Contains(url, "blog/blog") {
		log.Panic().
			Str("url", url).
			Msg("media url contains blog/blog")
	}

	if err := utils.CreateDirIfNotExist(m.cacheDirPath); err != nil {
		return nil, fmt.Errorf("error creating cache directory: %s", err)
	}

	key := getSHA256(url)
	file, err := os.OpenFile(path.Join(m.cacheDirPath, key), os.O_RDONLY, 0644)
	if err == nil {
		log.Info().
			Str("url", url).
			Msg("media found in cache")
		return file, nil
	}

	log.Info().
		Str("url", url).
		Msg("media will be fetched")

	var http_err error = fmt.Errorf("generic error")
	var resp *http.Response = nil

	retries := 0
	timeout := 1
	stop := false
	for retries < 5 && !stop {
		// Send at most 1 request per second
		// to avoid hammering servers and getting thresholded
		time.Sleep(time.Duration(timeout) * time.Second)
		resp, http_err = http.Get(url)
		timeout, stop = waitOrStop(resp)
		retries++
		timeout *= retries
	}

	if http_err != nil {
		return nil, fmt.Errorf("error fetching media %s: %s", url, http_err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching media %s: %s", url, resp.Status)
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	file, err = os.OpenFile(path.Join(m.cacheDirPath, key), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error creating cache file for media %s: %s", url, err)
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error writing media to cache %s: %s", url, err)
	}

	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("error closing cache file for media %s: %s", url, err)
	}

	file, err = os.OpenFile(path.Join(m.cacheDirPath, key), os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening cache file for media %s: %s", url, err)
	}
	return file, nil
}

func getSHA256(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
