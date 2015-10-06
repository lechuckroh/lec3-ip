package main
import (
	"image"
	"github.com/disintegration/gift"
)

type AutoCropOption struct {
	threshold         uint32  // threshold color value (0~0xffff)
	minRatio          float32 // min cropped ratio (height / width)
	maxRatio          float32 // max cropped ratio (height / width)
	maxWidthCropRate  float32 // max width crop rate (0 <= rate < 1.0)
	maxHeightCropRate float32 // max height crop rate (0 <= rate < 1.0)
}

type AutoCropFilter struct {
	edgeDetect *gift.GIFT
	option     AutoCropOption
}

// Create AutoCropFilter instance
func NewAutoCropFilter(option AutoCropOption) AutoCropFilter {
	edgeDetect := gift.New(
		gift.Convolution(
			[]float32{
				-1, -1, -1,
				-1, 8, -1,
				-1, -1, -1,
			},
			false, false, false, 0.0,
		))

	return AutoCropFilter{edgeDetect, option}
}

// override Filter.Run()
func (f AutoCropFilter) Run(s interface{}) interface{} {
	switch src := s.(type) {
	case image.Image:
		return f.runFilter(src)
	default:
		return nil
	}
}

// actual autoCriop implementation
func (f AutoCropFilter) runFilter(src image.Image) image.Image {
	bounds := src.Bounds()

	// Edge Detect
	edgeDetected := image.NewGray(bounds)
	f.edgeDetect.Draw(edgeDetected, src)

	// calculate boundary
	width := bounds.Dx()
	height := bounds.Dy()

	top := f.findTopEdge(edgeDetected, width, height)
	bottom := f.findBottomEdge(edgeDetected, width, height, top)
	left := f.findLeftEdge(edgeDetected, width, height, top, bottom)
	right := f.findRightEdge(edgeDetected, width, height, top, bottom, left)

	// crop image
	if top > 0 || left > 0 || right < width || bottom < height {
		cropRect := f.getCropRect(left, top, right, bottom, bounds)
		dest := image.NewRGBA(cropRect)
		crop := gift.New(gift.Crop(cropRect))
		crop.Draw(dest, src)
		return dest
	} else {
		return src
	}
}

// get constraint satisfied cropped Rectagle
func (f AutoCropFilter) getCropRect(left, top, right, bottom int, bounds image.Rectangle) image.Rectangle {
	initWidth := right - left
	initHeight := bottom - top
	width := initWidth
	height := initHeight
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// maxCropRate
	minWidth := int(float32(imgWidth) * f.option.maxWidthCropRate)
	minHeight := int(float32(imgHeight) * f.option.maxHeightCropRate)
	width = Max(width, minWidth)
	height = Max(height, minHeight)

	// ratio
	ratio := float32(height) / float32(width)
	if ratio < f.option.minRatio {
		height = int(float32(width) * f.option.minRatio)
	}
	if ratio > f.option.maxRatio {
		width = int(float32(height) / f.option.maxRatio)
	}

	// adjust border
	widthInc := width - initWidth
	heightInc := height - initHeight

	if widthInc > 0 {
		incHalf := int(float32(widthInc) / 2)
		dLeft := Min(left, incHalf)
		dRight := Min(width - right, incHalf)
		if dLeft > incHalf {
			dRight += incHalf - dLeft
		}
		if dRight < incHalf {
			dLeft += incHalf - dRight
		}

		left -= dLeft
		right += dRight
	}

	if heightInc > 0 {
		incHalf := int(float32(heightInc) / 2)
		dTop := Min(top, incHalf)
		dBottom := Min(height - bottom, incHalf)
		if dTop < incHalf {
			dBottom += incHalf - dTop
		}
		if dBottom < incHalf {
			dTop += incHalf - dBottom
		}

		top -= dTop
		bottom += dBottom
	}

	return image.Rect(left, top, right, bottom)
}

// Find top edge. 0 <= threshold <= 0xffff
func (f AutoCropFilter) findTopEdge(image *image.Gray, width, height int) int {
	threshold := f.option.threshold
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
func (f AutoCropFilter) findBottomEdge(image *image.Gray, width, height, top int) int {
	threshold := f.option.threshold
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
func (f AutoCropFilter) findLeftEdge(image *image.Gray, width, height, top, bottom int) int {
	threshold := f.option.threshold
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
func (f AutoCropFilter) findRightEdge(image *image.Gray, width, height, top, bottom, left int) int {
	threshold := f.option.threshold
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
