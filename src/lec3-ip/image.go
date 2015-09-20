package main
import (
	"image"
	"strings"
	"path/filepath"
	"errors"
	"os"
	"image/jpeg"
	"image/gif"
	"image/png"
	"io"
	"path"
)

func getExt(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// Load Image
func LoadImage(filename string) (image.Image, error) {
	var decoder func(io.Reader) (image.Image, error) = nil

	ext := getExt(filename)
	switch ext {
	case ".jpg", ".jpeg":
		decoder = jpeg.Decode
	case ".gif":
		decoder = gif.Decode
	case ".png":
		decoder = png.Decode
	}

	if decoder == nil {
		return nil, errors.New("Unsupported file format : " + ext)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	img, err := decoder(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// save image to jpeg file
func SaveJpeg(img image.Image, dir string, filename string, quality int) error {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	file, err := os.Create(path.Join(dir, filename))
	if err != nil {
		return err
	}

	return jpeg.Encode(file, img, &jpeg.Options{quality })
}
