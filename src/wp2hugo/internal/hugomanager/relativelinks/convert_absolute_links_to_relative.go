package relativelinks

import (
	"fmt"
	"os"
	"regexp"

	"github.com/rs/zerolog/log"
)

func ConvertAbsoluteLinksToRelative(path string, updateInline bool, hostname string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	re1 := regexp.MustCompile(fmt.Sprintf(`(http|https)://%s/([^\?]+)`, hostname))
	if re1.Match(data) {
		log.Warn().
			Str("path", path).
			Msg("Absolute link found in file")
		data = re1.ReplaceAll(data, []byte(`/$2`))
		log.Info().
			Str("path", path).
			Msg("File updated")
		if updateInline {
			err := os.WriteFile(path, data, 0o600)
			if err != nil {
				return fmt.Errorf("failed to write file %s: %w", path, err)
			}
		}
	}

	// if re1.MatchString(matter.Summary) {
	// 	log.Warn().
	// 		Str("path", path).
	// 		Msg("Absolute link found in summary")
	// 	matter.Summary = re1.ReplaceAllString(matter.Summary, `/$2`)
	// 	log.Info().
	// 		Str("path", path).
	// 		Str("summary", matter.Summary).
	// 		Msg("Summary updated")
	// 	if updateInline {
	// 		urlsuggest.UpdateFrontmatter(path, "summary", matter.Summary)
	// 	}
	// }
	return nil
}
