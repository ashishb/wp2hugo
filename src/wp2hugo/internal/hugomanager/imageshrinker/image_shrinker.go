package imageshrinker

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

// ResizeImage resizes the image at srcPath to the specified newWidth while maintaining aspect ratio,
func ResizeImage(srcPath string, destPath string, newWidth int) error {
	src, err := decode(srcPath)
	if err != nil {
		return fmt.Errorf("error decoding source image %s: %w", srcPath, err)
	}

	if src == nil {
		return fmt.Errorf("decoded image is nil for source image %s", srcPath)
	}

	ratio := (float64)(src.Bounds().Max.Y) / (float64)(src.Bounds().Max.X)
	newHeight := int(math.Round(float64(newWidth) * ratio))
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	err = encode(dst, destPath)
	if err != nil {
		return err
	}

	originalSize := GetFileSize(srcPath)
	newSize := GetFileSize(destPath)
	shrunkPct := 100.0 * (float64(originalSize-newSize) / float64(originalSize))
	log.Info().
		Str("srcPath", srcPath).
		Str("destPath", destPath).
		Int("newWidth", newWidth).
		Int("newHeight", newHeight).
		Str("ShrinkBy", fmt.Sprintf("%.0f%%", shrunkPct)).
		Msg("Resized image successfully")
	return nil
}

func decode(srcPath string) (image.Image, error) {
	r, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("error opening source image %s: %w", srcPath, err)
	}
	defer r.Close()

	var src image.Image
	switch strings.ToLower(path.Ext(srcPath)) {
	case ".jpg", ".jpeg":
		src, err = jpeg.Decode(r)
	case ".png":
		src, err = png.Decode(r)
	case ".gif":
		src, err = gif.Decode(r)
	case ".bmp":
		src, err = bmp.Decode(r)
	case ".tiff", ".tif":
		src, err = tiff.Decode(r)
	case ".webp":
		src, err = webp.Decode(r)
	default:
		err = fmt.Errorf("unsupported image format: %s", path.Ext(srcPath))
	}
	return src, err
}

func encode(dst *image.RGBA, destPath string) error {
	w, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating destination image %s: %w", destPath, err)
	}

	defer w.Close()
	switch strings.ToLower(path.Ext(destPath)) {
	case ".jpg", ".jpeg":
		return jpeg.Encode(w, dst, nil)
	case ".png":
		return png.Encode(w, dst)
	case ".gif":
		return gif.Encode(w, dst, nil)
	case ".bmp":
		return bmp.Encode(w, dst)
	case ".tiff", ".tif":
		return tiff.Encode(w, dst, nil)
	case ".webp":
		return errors.New("webp encoding not supported yet")
	default:
		return fmt.Errorf("unsupported image format: %s", path.Ext(destPath))
	}
}

func GetFileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("filePath", filePath).
			Msg("Error getting file info")
		return -1
	}

	return info.Size()
}
