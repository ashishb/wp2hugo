package nginxgenerator

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
)

//go:embed data/ngnix_base_config.conf
var _baseTemplate string

const _redirectHomepageQueryStringTemplate = `
            if ($query_string = "%s") { return 302 %s; }`

type (
	_SourceQueryString string
	_DestinationPath   string
)

type Config struct {
	redirects map[_SourceQueryString]_DestinationPath
}

func NewConfig() *Config {
	return &Config{
		redirects: make(map[_SourceQueryString]_DestinationPath),
	}
}

func (c *Config) AddRedirect(source string, destination string) error {
	if !strings.HasPrefix(source, "/?") {
		// No strong reason, just that this suffices for most of the use cases
		return errors.New("only source path starting with /? are supported for now")
	}
	// Sanity check to ensure that we are redirecting to a relative path
	if !strings.HasPrefix(destination, "/") {
		return errors.New("destination path must start with /")
	}
	source = strings.TrimPrefix(source, "/?")
	c.redirects[_SourceQueryString(source)] = _DestinationPath(destination)
	return nil
}

func (c *Config) generateRedirects() string {
	var sb strings.Builder
	for sourceQueryString, destination := range c.redirects {
		fmt.Fprintf(&sb, _redirectHomepageQueryStringTemplate, sourceQueryString, destination)
	}
	return sb.String()
}

func (c *Config) Generate() string {
	return fmt.Sprintf(_baseTemplate, c.generateRedirects())
}
