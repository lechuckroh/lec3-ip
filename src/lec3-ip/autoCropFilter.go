package main

import (
	"github.com/disintegration/gift"
	"github.com/mitchellh/mapstructure"
	"image"
	"image/draw"
	"image/color"
)

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type AutoCropOption struct {
	Threshold            uint8   // min brightness of space (0~255)
	MinRatio             float32 // min cropped ratio (height / width)
	MaxRatio             float32 // max cropped ratio (height / width)
	MaxWidthCropRate     float32 // max width crop rate (0 <= rate < 1.0)
	MaxHeightCropRate    float32 // max height crop rate (0 <= rate < 1.0)
	EmptyLineMaxDotCount int
	MarginTop            int
	MarginBottom         int
	MarginLeft           int
	MarginRight          int
	PaddingTop           int
	PaddingBottom        int
	PaddingLeft          int
	PaddingRight         int
	MaxCropTop           int
	MaxCropBottom        int
	MaxCropLeft          int
	MaxCropRight         int
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
	o := f.option

	// calculate boundary
	width, height := bounds.Dx(), bounds.Dy()

	top := f.findTopEdge(src, width, height)
	bottom := f.findBottomEdge(src, width, height, top)
	left := f.findLeftEdge(src, width, height, top, bottom)
	right := f.findRightEdge(src, width, height, top, bottom, left)

	// maxCrop
	disableMaxCrop := (o.MaxCropTop == 0 && o.MaxCropBottom == 0 && o.MaxCropLeft == 0 && o.MaxCropRight == 0)
	if !disableMaxCrop {
		if o.MaxCropTop >= 0 {
			top = Min(o.MaxCropTop, top)
		}
		if o.MaxCropBottom >= 0 {
			bottom = Max(height - o.MaxCropBottom, bottom)
		}
		if o.MaxCropLeft >= 0 {
			left = Min(o.MaxCropLeft, left)
		}
		if o.MaxCropRight >= 0 {
			right = Max(width - o.MaxCropRight, right)
		}
	}

	// crop image
	if top > 0 || left > 0 || right + 1 < width || bottom + 1 < height {
		cropRect := GetCropRect(left, top, right + 1, bottom + 1, bounds, o.MaxWidthCropRate, o.MaxHeightCropRate, o.MinRatio, o.MaxRatio)
		dest := image.NewRGBA(cropRect)
		draw.Draw(dest, dest.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)
		crop := gift.New(gift.Crop(cropRect))
		crop.Draw(dest, src)
		return dest, cropRect
	} else {
		return src, bounds
	}
}

// Find top edge. 0 <= threshold <= 0xffff
func (f AutoCropFilter) findTopEdge(image image.Image, width, height int) int {
	thresholdSum := uint32(f.option.Threshold) * 256 * 3
	yEnd := height - f.option.PaddingBottom
	xEnd := width - f.option.PaddingRight
	maxDotCount := f.option.EmptyLineMaxDotCount
	for y := f.option.PaddingTop; y < yEnd; y++ {
		dotCount := 0
		for x := f.option.PaddingLeft; x < xEnd; x++ {
			if r, g, b, _ := image.At(x, y).RGBA(); (r + g + b) < thresholdSum {
				dotCount++
				if dotCount > maxDotCount {
					return Max(0, y - f.option.MarginTop)
				}
			}
		}
	}
	return height
}

// Find bottom edge. 0 <= threshold <= 0xffff
func (f AutoCropFilter) findBottomEdge(image image.Image, width, height, top int) int {
	thresholdSum := uint32(f.option.Threshold) * 256 * 3
	xEnd := width - f.option.PaddingRight
	maxDotCount := f.option.EmptyLineMaxDotCount
	for y := height - f.option.PaddingBottom - 1; y > top; y-- {
		dotCount := 0
		for x := f.option.PaddingLeft; x < xEnd; x++ {
			if r, g, b, _ := image.At(x, y).RGBA(); (r + g + b) < thresholdSum {
				dotCount++
				if dotCount > maxDotCount {
					return Min(height - 1, y + f.option.MarginBottom)
				}
			}
		}
	}
	return top
}

// Find left edge. 0 <= threshold <= 0xffff
func (f AutoCropFilter) findLeftEdge(image image.Image, width, height, top, bottom int) int {
	thresholdSum := uint32(f.option.Threshold) * 256 * 3
	yEnd := height - f.option.PaddingBottom
	xEnd := width - f.option.PaddingRight
	maxDotCount := f.option.EmptyLineMaxDotCount
	for x := f.option.PaddingLeft; x < xEnd; x++ {
		dotCount := 0
		for y := top + 1; y < yEnd; y++ {
			if r, g, b, _ := image.At(x, y).RGBA(); (r + g + b) < thresholdSum {
				dotCount++
				if dotCount > maxDotCount {
					return Max(0, x - f.option.MarginLeft)
				}
			}
		}
	}
	return width
}

// Find right edge. 0 <= threshold <= 0xffff
func (f AutoCropFilter) findRightEdge(image image.Image, width, height, top, bottom, left int) int {
	thresholdSum := uint32(f.option.Threshold) * 256 * 3
	maxDotCount := f.option.EmptyLineMaxDotCount
	for x := width - f.option.PaddingRight - 1; x > left; x-- {
		dotCount := 0
		for y := top + 1; y < bottom; y++ {
			if r, g, b, _ := image.At(x, y).RGBA(); (r + g + b) < thresholdSum {
				dotCount++
				if dotCount > maxDotCount {
					return Min(width - 1, x + f.option.MarginRight)
				}
			}
		}
	}
	return left
}
