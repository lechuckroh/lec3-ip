package main

import (
	"errors"
	"github.com/disintegration/gift"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	} else {
		defer func() {
			file.Close()
		}()
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
	} else {
		defer func() {
			file.Close()
		}()
	}

	return jpeg.Encode(file, img, &jpeg.Options{quality})
}

// create image
func CreateImage(width, height int, bgColor color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, bgColor)
		}
	}
	return img
}

// draw filled rect
func FillRect(img *image.RGBA, x1, y1, x2, y2 int, rectColor color.Color) {
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			img.Set(x, y, rectColor)
		}
	}
}

// draw line
func DrawLine(img *image.RGBA, x1, y1, x2, y2 int, lineColor color.Color) {
	dx, dy := x2-x1, y2-y1
	if dx <= dy {
		incX := float32(dx) / float32(dy)
		x := float32(x1)
		for y := y1; y < y2; y++ {
			img.Set(int(x), y, lineColor)
			x += incX
		}
	} else {
		incY := float32(dy) / float32(dx)
		y := float32(y1)
		for x := x1; x < x2; x++ {
			img.Set(x, int(y), lineColor)
			y += incY
		}
	}
}

// calculate width/height after rotation
func CalcRotatedSize(w, h int, angle float32) (int, int) {
	if w <= 0 || h <= 0 {
		return 0, 0
	}

	xoff := float32(w)/2 - 0.5
	yoff := float32(h)/2 - 0.5

	asin, acos := Sincosf32(angle)
	x1, y1 := RotatePoint(0-xoff, 0-yoff, asin, acos)
	x2, y2 := RotatePoint(float32(w-1)-xoff, 0-yoff, asin, acos)
	x3, y3 := RotatePoint(float32(w-1)-xoff, float32(h-1)-yoff, asin, acos)
	x4, y4 := RotatePoint(0-xoff, float32(h-1)-yoff, asin, acos)

	minx := Minf32(x1, Minf32(x2, Minf32(x3, x4)))
	maxx := Maxf32(x1, Maxf32(x2, Maxf32(x3, x4)))
	miny := Minf32(y1, Minf32(y2, Minf32(y3, y4)))
	maxy := Maxf32(y1, Maxf32(y2, Maxf32(y3, y4)))

	neww := maxx - minx + 1
	if neww-Floorf32(neww) > 0.01 {
		neww += 2
	}
	newh := maxy - miny + 1
	if newh-Floorf32(newh) > 0.01 {
		newh += 2
	}
	return int(neww), int(newh)
}

func RotatePoint(x, y, asin, acos float32) (float32, float32) {
	newx := x*acos - y*asin
	newy := x*asin + y*acos
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

// get constraint satisfied cropped Rectagle
func GetCropRect(left, top, right, bottom int, bounds image.Rectangle, maxWidthCropRate, maxHeightCropRate, minRatio, maxRatio float32) image.Rectangle {
	initWidth, initHeight := right-left, bottom-top
	width, height := initWidth, initHeight
	imgWidth, imgHeight := bounds.Dx(), bounds.Dy()

	// maxCropRate
	minWidth := int(float32(imgWidth) * (1 - maxWidthCropRate))
	minHeight := int(float32(imgHeight) * (1 - maxHeightCropRate))
	width, height = Max(width, minWidth), Max(height, minHeight)

	// ratio
	ratio := float32(height) / float32(width)
	if ratio < minRatio {
		height = Max(minHeight, int(float32(width)*minRatio))
	}
	if ratio > maxRatio {
		width = Max(minWidth, int(float32(height)/maxRatio))
	}

	// adjust border
	widthInc, heightInc := width-initWidth, height-initHeight
	widthMargin, heightMargin := width-initWidth, height-initHeight

	if widthInc > 0 {
		widthHalfMargin := int(float32(widthMargin) / 2)
		leftMargin := Min(left, widthHalfMargin)
		rightMargin := Min(imgWidth-right, widthMargin-leftMargin)
		left -= leftMargin
		right += rightMargin

		w := right - left
		dx := widthInc - w + initWidth
		if dx > 0 {
			leftRatio := left / (left + imgWidth - right)
			leftSpace := dx * leftRatio
			left -= leftSpace
			right += dx - leftSpace
		}
	}

	if heightInc > 0 {
		heightHalfMargin := int(float32(heightMargin) / 2)
		topMargin := Min(top, heightHalfMargin)
		bottomMargin := Min(imgHeight-bottom, heightMargin-topMargin)
		top -= topMargin
		bottom += bottomMargin

		h := bottom - top
		dy := heightInc - h + initHeight
		if dy > 0 {
			topRatio := top / (top + imgHeight - bottom)
			topSpace := dy * topRatio
			top -= topSpace
			bottom += dy - topSpace
		}
	}

	return image.Rect(left, top, right, bottom)
}
