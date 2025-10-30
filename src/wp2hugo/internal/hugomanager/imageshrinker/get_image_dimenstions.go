package imageshrinker

import (
	"image"
	"os"
)

type Size struct {
	width  int
	height int
}

func (w Size) Width() int {
	return w.width
}

func (w Size) Height() int {
	return w.height
}

// GetImageDimensions returns the width and height of the image at the given path.
func GetImageDimensions(path string) (*Size, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}
	return &Size{
		width:  cfg.Width,
		height: cfg.Height,
	}, nil
}
