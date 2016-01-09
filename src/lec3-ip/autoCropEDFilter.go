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
type AutoCropEDOption struct {
	Threshold            uint8   // edge strength threshold (0~255(max edge))
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

func NewAutoCropEDOption(m map[string]interface{}) (*AutoCropEDOption, error) {
	option := AutoCropEDOption{}

	err := mapstructure.Decode(m, &option)
	if err != nil {
		return nil, err
	}

	return &option, nil
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type AutoCropEDResult struct {
	image image.Image
	rect  image.Rectangle
}

func (r AutoCropEDResult) Image() image.Image {
	return r.image
}

func (r AutoCropEDResult) Log() {
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type AutoCropEDFilter struct {
	edgeDetect *gift.GIFT
	option     AutoCropEDOption
}

// Create AutoCropEDFilter instance
func NewAutoCropEDFilter(option AutoCropEDOption) *AutoCropEDFilter {
	edgeDetect := gift.New(
		gift.Convolution(
			[]float32{
				-1, -1, -1,
				-1, 8, -1,
				-1, -1, -1,
			},
			false, false, false, 0.0,
		))
	return &AutoCropEDFilter{
		edgeDetect: edgeDetect,
		option:     option,
	}
}

// Implements Filter.Run()
func (f AutoCropEDFilter) Run(s *FilterSource) FilterResult {
	img, rect := f.run(s.image)
	return AutoCropEDResult{img, rect}
}

// actual autoCrop implementation
func (f AutoCropEDFilter) run(src image.Image) (image.Image, image.Rectangle) {
	bounds := src.Bounds()
	o := f.option

	// Edge Detect
	edgeDetected := image.NewGray(bounds)
	f.edgeDetect.Draw(edgeDetected, src)

	//	SaveJpeg(src, "./", "autoCropSrc.jpg", 80)
	//	SaveJpeg(edgeDetected, "./", "autoCropED.jpg", 80)

	// calculate boundary
	width, height := bounds.Dx(), bounds.Dy()

	top := f.findTopEdge(edgeDetected, width, height) + 1
	bottom := f.findBottomEdge(edgeDetected, width, height, top)
	left := f.findLeftEdge(edgeDetected, width, height, top, bottom) + 1
	right := f.findRightEdge(edgeDetected, width, height, top, bottom, left)

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
func (f AutoCropEDFilter) findTopEdge(image *image.Gray, width, height int) int {
	threshold := uint32(f.option.Threshold) * 256
	yEnd := height - f.option.PaddingBottom
	xEnd := width - f.option.PaddingRight
	maxDotCount := f.option.EmptyLineMaxDotCount
	for y := f.option.PaddingTop; y < yEnd; y++ {
		dotCount := 0
		for x := f.option.PaddingLeft; x < xEnd; x++ {
			if r, _, _, _ := image.At(x, y).RGBA(); r > threshold {
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
func (f AutoCropEDFilter) findBottomEdge(image *image.Gray, width, height, top int) int {
	threshold := uint32(f.option.Threshold) * 256
	xEnd := width - f.option.PaddingRight
	maxDotCount := f.option.EmptyLineMaxDotCount
	for y := height - f.option.PaddingBottom - 1; y > top; y-- {
		dotCount := 0
		for x := f.option.PaddingLeft; x < xEnd; x++ {
			if r, _, _, _ := image.At(x, y).RGBA(); r > threshold {
				dotCount++
				if dotCount > maxDotCount {
					return Min(height - 1, y - 1 + f.option.MarginBottom)
				}
			}
		}
	}
	return top
}

// Find left edge. 0 <= threshold <= 0xffff
func (f AutoCropEDFilter) findLeftEdge(image *image.Gray, width, height, top, bottom int) int {
	threshold := uint32(f.option.Threshold) * 256
	yEnd := height - f.option.PaddingBottom
	xEnd := width - f.option.PaddingRight
	maxDotCount := f.option.EmptyLineMaxDotCount
	for x := f.option.PaddingLeft; x < xEnd; x++ {
		dotCount := 0
		for y := top + 1; y < yEnd; y++ {
			if r, _, _, _ := image.At(x, y).RGBA(); r > threshold {
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
func (f AutoCropEDFilter) findRightEdge(image *image.Gray, width, height, top, bottom, left int) int {
	threshold := uint32(f.option.Threshold) * 256
	maxDotCount := f.option.EmptyLineMaxDotCount
	for x := width - f.option.PaddingRight - 1; x > left; x-- {
		dotCount := 0
		for y := top + 1; y < bottom; y++ {
			if r, _, _, _ := image.At(x, y).RGBA(); r > threshold {
				dotCount++
				if dotCount > maxDotCount {
					return Min(width - 1, x - 1 + f.option.MarginRight)
				}
			}
		}
	}
	return left
}
