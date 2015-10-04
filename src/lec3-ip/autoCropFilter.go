package main
import (
	"github.com/disintegration/gift"
	"image"
)

type AutoCropFilter struct {
	edgeDetect *gift.GIFT
}

func NewAutoCropFilter() AutoCropFilter {
	edgeDetect := gift.New(
		gift.Convolution(
			[]float32{
				-1, -1, -1,
				-1, 8, -1,
				-1, -1, -1,
			},
			false, false, false, 0.0,
		))

	return AutoCropFilter{edgeDetect}
}

func (f AutoCropFilter) Run(src image.Image) image.Image {
	bounds := src.Bounds()

	// Edge Detect
	edgeDetected := image.NewGray(bounds)
	f.edgeDetect.Draw(edgeDetected, src)

	// calculate boundary
	width := bounds.Dx()
	height := bounds.Dy()
	threshold := uint32(0)

	top := findTopEdge(edgeDetected, width, height, threshold)
	bottom := findBottomEdge(edgeDetected, width, height, top, threshold)
	left := findLeftEdge(edgeDetected, width, height, top, bottom, threshold)
	right := findRightEdge(edgeDetected, width, height, top, bottom, left, threshold)

	// crop image
	if top > 0 || left > 0 || right < (width - 1) || bottom < (height - 1) {
		dest := image.NewRGBA(image.Rect(left, top, right, bottom))
		crop := gift.New(gift.Crop(image.Rect(left, top, right, bottom)))
		crop.Draw(dest, src)
		return dest
	} else {
		return src
	}
}

// Find top edge. 0 <= threshold <= 0xffff
func findTopEdge(image *image.Gray, width, height int, threshold uint32) int {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := image.At(x, y)
			r, _, _, _ := color.RGBA()
			if r > threshold {
				return y
			}
		}
	}
	return height
}

// Find bottom edge. 0 <= threshold <= 0xffff
func findBottomEdge(image *image.Gray, width, height, top int, threshold uint32) int {
	for y := height - 1; y > top; y-- {
		for x := 0; x < width; x++ {
			color := image.At(x, y)
			r, _, _, _ := color.RGBA()
			if r > threshold {
				return y
			}
		}
	}
	return top
}

// Find left edge. 0 <= threshold <= 0xffff
func findLeftEdge(image *image.Gray, width, height, top, bottom int, threshold uint32) int {
	for x := 0; x < width; x++ {
		for y := top + 1; y < bottom; y++ {
			color := image.At(x, y)
			r, _, _, _ := color.RGBA()
			if r > threshold {
				return x
			}
		}
	}
	return width
}

// Find right edge. 0 <= threshold <= 0xffff
func findRightEdge(image *image.Gray, width, height, top, bottom, left int, threshold uint32) int {
	for x := width - 1; x > left; x-- {
		for y := top + 1; y < bottom; y++ {
			color := image.At(x, y)
			r, _, _, _ := color.RGBA()
			if r > threshold {
				return x
			}
		}
	}
	return left
}
