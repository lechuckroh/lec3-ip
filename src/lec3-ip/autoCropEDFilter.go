package main
import (
	"image"
	"github.com/disintegration/gift"
	"log"
	"github.com/mitchellh/mapstructure"
)

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type AutoCropEDOption struct {
	threshold         uint8   // edge strength threshold (0~255(max edge))
	minRatio          float32 // min cropped ratio (height / width)
	maxRatio          float32 // max cropped ratio (height / width)
	maxWidthCropRate  float32 // max width crop rate (0 <= rate < 1.0)
	maxHeightCropRate float32 // max height crop rate (0 <= rate < 1.0)
	marginTop         int
	marginBottom      int
	marginLeft        int
	marginRight       int
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
		option: option,
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

	// Edge Detect
	edgeDetected := image.NewGray(bounds)
	f.edgeDetect.Draw(edgeDetected, src)

	// calculate boundary
	width, height := bounds.Dx(), bounds.Dy()

	top := f.findTopEdge(edgeDetected, width, height) + 1
	bottom := f.findBottomEdge(edgeDetected, width, height, top)
	left := f.findLeftEdge(edgeDetected, width, height, top, bottom) + 1
	right := f.findRightEdge(edgeDetected, width, height, top, bottom, left)

	log.Printf("(%v,%v)-(%v,%v)\n", left, top, right, bottom)

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
func (f AutoCropEDFilter) getCropRect(left, top, right, bottom int, bounds image.Rectangle) image.Rectangle {
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
func (f AutoCropEDFilter) findTopEdge(image *image.Gray, width, height int) int {
	threshold := uint32(f.option.threshold) * 256
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if r, _, _, _ := image.At(x, y).RGBA(); r > threshold {
				return Max(0, y - f.option.marginTop)
			}
		}
	}
	return height
}

// Find bottom edge. 0 <= threshold <= 0xffff
func (f AutoCropEDFilter) findBottomEdge(image *image.Gray, width, height, top int) int {
	threshold := uint32(f.option.threshold) * 256
	for y := height - 1; y > top; y-- {
		for x := 0; x < width; x++ {
			if r, _, _, _ := image.At(x, y).RGBA(); r > threshold {
				return Min(height - 1, y - 1 + f.option.marginBottom)
			}
		}
	}
	return top
}

// Find left edge. 0 <= threshold <= 0xffff
func (f AutoCropEDFilter) findLeftEdge(image *image.Gray, width, height, top, bottom int) int {
	threshold := uint32(f.option.threshold) * 256
	for x := 0; x < width; x++ {
		for y := top + 1; y < bottom; y++ {
			if r, _, _, _ := image.At(x, y).RGBA(); r > threshold {
				return Max(0, x - f.option.marginLeft)
			}
		}
	}
	return width
}

// Find right edge. 0 <= threshold <= 0xffff
func (f AutoCropEDFilter) findRightEdge(image *image.Gray, width, height, top, bottom, left int) int {
	threshold := uint32(f.option.threshold) * 256
	for x := width - 1; x > left; x-- {
		for y := top + 1; y < bottom; y++ {
			if r, _, _, _ := image.At(x, y).RGBA(); r > threshold {
				return Min(width - 1, x - 1 + f.option.marginRight)
			}
		}
	}
	return left
}
