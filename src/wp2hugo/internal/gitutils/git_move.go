package gitutils

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/ashishb/wp2hugo/src/wp2hugo/internal/utils"
	"github.com/rs/zerolog/log"
)

func GitMove(ctx context.Context, path string, newFilePath string) error {
	if !utils.FileExists(path) {
		return fmt.Errorf("file does not exist: '%s'", path)
	}

	// Convert path to absolute path
	path, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error getting absolute path for '%s': %w", path, err)
	}

	newFilePath, err = filepath.Abs(newFilePath)
	if err != nil {
		return fmt.Errorf("error getting absolute path for '%s': %w", newFilePath, err)
	}

	if utils.FileExists(newFilePath) {
		return fmt.Errorf("file already exists: '%s'", newFilePath)
	}

	cmd := exec.CommandContext(ctx, "git", "mv", path, newFilePath)
	// Run the command in the directory of the file for "git" to work
	cmd.Dir = filepath.Dir(path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error moving file '%s' to '%s': %w\n%s", path, newFilePath, err, output)
	}
	log.Debug().
		Str("path", path).
		Str("newFilePath", newFilePath).
		Str("output", string(output)).
		Msg("Moved file using git mv")
	return nil
}
