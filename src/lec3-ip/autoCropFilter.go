package main
import (
	"image"
	"github.com/disintegration/gift"
	"github.com/mitchellh/mapstructure"
)

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type AutoCropOption struct {
	threshold         uint8   // min brightness of space (0~255)
	minRatio          float32 // min cropped ratio (height / width)
	maxRatio          float32 // max cropped ratio (height / width)
	maxWidthCropRate  float32 // max width crop rate (0 <= rate < 1.0)
	maxHeightCropRate float32 // max height crop rate (0 <= rate < 1.0)
	marginTop         int
	marginBottom      int
	marginLeft        int
	marginRight       int
}

func NewAutoCropOption(m map[string]interface{}) (*AutoCropOption, error) {
	option := AutoCropOption{}

	err := mapstructure.Decode(m, &option)
	if err != nil {
		return nil, err
	}

	return &option, nil
}

type AutoCropResult struct {
	image image.Image
	rect  image.Rectangle
}

func (r AutoCropResult) Image() image.Image {
	return r.image
}

func (r AutoCropResult) Log() {
}


// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type AutoCropFilter struct {
	option AutoCropOption
}

// Create AutoCropFilter instance
func NewAutoCropFilter(option AutoCropOption) *AutoCropFilter {
	return &AutoCropFilter{option: option}
}

// Implements Filter.Run()
func (f AutoCropFilter) Run(s *FilterSource) FilterResult {
	img, rect := f.run(s.image)
	return AutoCropResult{img, rect}
}

// actual autoCrop implementation
func (f AutoCropFilter) run(src image.Image) (image.Image, image.Rectangle) {
	bounds := src.Bounds()

	// calculate boundary
	width, height := bounds.Dx(), bounds.Dy()

	top := f.findTopEdge(src, width, height)
	bottom := f.findBottomEdge(src, width, height, top)
	left := f.findLeftEdge(src, width, height, top, bottom)
	right := f.findRightEdge(src, width, height, top, bottom, left)

	// crop image
	if top > 0 || left > 0 || right + 1 < width || bottom + 1 < height {
		cropRect := f.getCropRect(left, top, right + 1, bottom + 1, bounds)
		dest := image.NewRGBA(cropRect)
		crop := gift.New(gift.Crop(cropRect))
		crop.Draw(dest, src)
		return dest, cropRect
	} else {
		return src, bounds
	}
}

// get constraint satisfied cropped Rectagle
func (f AutoCropFilter) getCropRect(left, top, right, bottom int, bounds image.Rectangle) image.Rectangle {
	initWidth, initHeight := right - left, bottom - top
	width, height := initWidth, initHeight
	imgWidth, imgHeight := bounds.Dx(), bounds.Dy()

	// maxCropRate
	minWidth := int(float32(imgWidth) * f.option.maxWidthCropRate)
	minHeight := int(float32(imgHeight) * f.option.maxHeightCropRate)
	width, height = Max(width, minWidth), Max(height, minHeight)


	// ratio
	ratio := float32(height) / float32(width)
	if ratio < f.option.minRatio {
		height = int(float32(width) * f.option.minRatio)
	}
	if ratio > f.option.maxRatio {
		width = int(float32(height) / f.option.maxRatio)
	}

	// adjust border
	widthInc, heightInc := width - initWidth, height - initHeight
	widthMargin, heightMargin := width - initWidth, height - initHeight

	if widthInc > 0 {
		widthHalfMargin := int(float32(widthMargin) / 2)
		leftMargin := Min(left, widthHalfMargin)
		rightMargin := Min(imgWidth - right, widthMargin - leftMargin)
		left -= leftMargin
		right += rightMargin
	}

	if heightInc > 0 {
		heightHalfMargin := int(float32(heightMargin) / 2)
		topMargin := Min(top, heightHalfMargin)
		bottomMargin := Min(imgHeight - bottom, heightMargin - topMargin)
		top -= topMargin
		bottom += bottomMargin
	}

	return image.Rect(left, top, right, bottom)
}

// Find top edge. 0 <= threshold <= 0xffff
func (f AutoCropFilter) findTopEdge(image image.Image, width, height int) int {
	threshold := uint32(f.option.threshold) * 256
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if r, g, b, _ := image.At(x, y).RGBA(); (r + g + b) / 3 < threshold {
				return Max(0, y - f.option.marginTop)
			}
		}
	}
	return height
}

// Find bottom edge. 0 <= threshold <= 0xffff
func (f AutoCropFilter) findBottomEdge(image image.Image, width, height, top int) int {
	threshold := uint32(f.option.threshold) * 256
	for y := height - 1; y > top; y-- {
		for x := 0; x < width; x++ {
			if r, g, b, _ := image.At(x, y).RGBA(); (r + g + b) / 3 < threshold {
				return Min(height - 1, y + f.option.marginBottom)
			}
		}
	}
	return top
}

// Find left edge. 0 <= threshold <= 0xffff
func (f AutoCropFilter) findLeftEdge(image image.Image, width, height, top, bottom int) int {
	threshold := uint32(f.option.threshold) * 256
	for x := 0; x < width; x++ {
		for y := top + 1; y < bottom; y++ {
			if r, g, b, _ := image.At(x, y).RGBA(); (r + g + b) / 3 < threshold {
				return Max(0, x - f.option.marginLeft)
			}
		}
	}
	return width
}

// Find right edge. 0 <= threshold <= 0xffff
func (f AutoCropFilter) findRightEdge(image image.Image, width, height, top, bottom, left int) int {
	threshold := uint32(f.option.threshold) * 256
	for x := width - 1; x > left; x-- {
		for y := top + 1; y < bottom; y++ {
			if r, g, b, _ := image.At(x, y).RGBA(); (r + g + b) / 3 < threshold {
				return Min(width - 1, x + f.option.marginRight)
			}
		}
	}
	return left
}
