package mediacache

import (
	"crypto/sha256"
	"fmt"
	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"path"
)

type MediaCache struct {
	cacheDirPath string
}

func New(cacheDirPath string) MediaCache {
	return MediaCache{cacheDirPath: cacheDirPath}
}

func (m MediaCache) GetReader(url string) (io.Reader, error) {
	if err := utils.CreateDirIfNotExist(m.cacheDirPath); err != nil {
		return nil, fmt.Errorf("error creating cache directory: %s", err)
	}

	key := getSHA256(url)
	file, err := os.OpenFile(path.Join(m.cacheDirPath, key), os.O_RDONLY, 0644)
	if err == nil {
		log.Info().Msgf("media %s found in cache", url)
		return file, nil
	}
	log.Info().Msgf("media %s will be fetched", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching media %s: %s", url, err)
	}
	defer resp.Body.Close()
	file, err = os.OpenFile(path.Join(m.cacheDirPath, key), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error creating cache file for media %s: %s", url, err)
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error writing media to cache %s: %s", url, err)
	}
	file.Close()
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
