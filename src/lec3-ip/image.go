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
	"image/color"
	"github.com/disintegration/gift"
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

// create image with colored rectangle
func CreateImageWithRect(width, height, x1, y1, x2, y2 int, bgColor color.Color, rectColor color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		rectArea := x >= x1 && x < x2
		for y := 0; y < height; y++ {
			if rectArea && y >= y1 && y < y2 {
				img.Set(x, y, rectColor)
			} else {
				img.Set(x, y, bgColor)
			}
		}
	}

	return img
}

// calculate width/height after rotation
func CalcRotatedSize(w, h int, angle float32) (int, int) {
	if w <= 0 || h <= 0 {
		return 0, 0
	}

	xoff := float32(w) / 2 - 0.5
	yoff := float32(h) / 2 - 0.5

	asin, acos := Sincosf32(angle)
	x1, y1 := RotatePoint(0 - xoff, 0 - yoff, asin, acos)
	x2, y2 := RotatePoint(float32(w - 1) - xoff, 0 - yoff, asin, acos)
	x3, y3 := RotatePoint(float32(w - 1) - xoff, float32(h - 1) - yoff, asin, acos)
	x4, y4 := RotatePoint(0 - xoff, float32(h - 1) - yoff, asin, acos)

	minx := Minf32(x1, Minf32(x2, Minf32(x3, x4)))
	maxx := Maxf32(x1, Maxf32(x2, Maxf32(x3, x4)))
	miny := Minf32(y1, Minf32(y2, Minf32(y3, y4)))
	maxy := Maxf32(y1, Maxf32(y2, Maxf32(y3, y4)))

	neww := maxx - minx + 1
	if neww - Floorf32(neww) > 0.01 {
		neww += 2
	}
	newh := maxy - miny + 1
	if newh - Floorf32(newh) > 0.01 {
		newh += 2
	}
	return int(neww), int(newh)
}

func RotatePoint(x, y, asin, acos float32) (float32, float32) {
	newx := x * acos - y * asin
	newy := x * asin + y * acos
	return newx, newy
}

// Rotate image
func RotateImage(src image.Image, angle float32, bgColor color.Color) image.Image {
	bounds := src.Bounds()
	width, height := CalcRotatedSize(bounds.Dx(), bounds.Dy(), angle)
	dest := image.NewRGBA(image.Rect(0, 0, width, height))
	gift.New(gift.Rotate(angle, bgColor, gift.CubicInterpolation)).Draw(dest, src)
	return dest
}
