package frontmatterhelper

import (
	"fmt"
	"os"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func UpdateFrontmatter(path string, key string, value string) error {
	fullmatter, restOfTheFile, err := getFullFrontMatter(path)
	if err != nil {
		return err
	}

	if !strings.Contains(key, ".") {
		fullmatter[key] = value
	} else if strings.Count(key, ".") == 1 {
		key0 := strings.Split(key, ".")[0]
		key1 := strings.Split(key, ".")[1]
		if fullmatter[key0] == nil {
			log.Fatal().
				Str("key", key).
				Msg("Key is nil")
		}
		map1, ok := fullmatter[key0].(map[any]any)
		if !ok {
			log.Fatal().
				Str("key", key).
				Any("map1", map1).
				Any("fullmatter", fullmatter).
				Msg("Key is not a map")
		}
		map1[key1] = value
	} else {
		return fmt.Errorf("key '%s' is not supported", key)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	_, err = file.WriteString("---\n")
	if err != nil {
		return err
	}

	yamlEncoder := yaml.NewEncoder(file)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(&fullmatter)
	if err != nil {
		return err
	}
	_, err = file.WriteString("---\n")
	if err != nil {
		return err
	}
	_, err = file.Write(restOfTheFile)
	if err != nil {
		return err
	}
	return nil
}

func getFullFrontMatter(path string) (map[string]any, []byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var matter map[string]any
	restOfTheFile, err := frontmatter.Parse(file, &matter)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("Error parsing front matter")
		return nil, nil, err
	}
	return matter, restOfTheFile, nil
}
