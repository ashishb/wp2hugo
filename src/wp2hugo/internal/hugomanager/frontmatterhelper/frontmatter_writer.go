package frontmatterhelper

import (
	"github.com/adrg/frontmatter"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
)

func UpdateFrontmatter(path string, key string, value string) error {
	fullmatter, restOfTheFile, err := getFullFrontMatter(path)
	if err != nil {
		return err
	}

	fullmatter[key] = value
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte("---\n"))
	if err != nil {
		return err
	}

	yamlEncoder := yaml.NewEncoder(file)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(&fullmatter)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte("---\n"))
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
