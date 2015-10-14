package main
import (
	"image"
	"math"
	"github.com/disintegration/gift"
	"image/color"
	"fmt"
)

type DeskewOption struct {
	maxRotation          float32     // max rotation angle (0 <= value <= 360)
	incrStep             float32     // rotation angle increment step (0 <= value <= 360)
	bgColor              color.Color // background color
	threshold            uint32      // threshold color value (0~0xffff)
	emptyLineMinDotCount int
}

type DeskewFilter struct {
	edgeDetect *gift.GIFT
	rotateMap  map[float32]*gift.GIFT
	option     DeskewOption
}

// Create DeskewFilter instance
func NewDeskewFilter(option DeskewOption) DeskewFilter {
	edgeDetect := gift.New(
		gift.Convolution(
			[]float32{
				-1, -1, -1,
				-1, 8, -1,
				-1, -1, -1,
			},
			false, false, false, 0.0,
		))

	rotateMap := make(map[float32]*gift.GIFT)
	for angle := float32(0); angle < option.maxRotation; angle += option.incrStep {
		rotateMap[angle] = gift.New(gift.Rotate(angle, option.bgColor, gift.CubicInterpolation))
		rotateMap[-angle] = gift.New(gift.Rotate(-angle, option.bgColor, gift.CubicInterpolation))
	}

	return DeskewFilter{edgeDetect, rotateMap, option}
}

// Implements Filter.Run()
func (f DeskewFilter) Run(s interface{}) interface{} {
	switch src := s.(type) {
	case image.Image:
		return f.run(src)
	default:
		return nil
	}
}

// actual deskew implementation
func (f DeskewFilter) run(src image.Image) image.Image {
	bounds := src.Bounds()

	// Edge Detect
	edgeDetected := image.NewGray(bounds)
	f.edgeDetect.Draw(edgeDetected, src)

	// Find preferred rotation angle
	if angle := f.detectRotationAngle(edgeDetected); angle != 0 {
		return f.rotateImage(src, angle)
	}
	return src
}

func (f DeskewFilter) detectRotationAngle(src image.Image) float32 {
	// increase rotation angle by incrStep
	maxEmptyLineCount := 0
	detectedAngle := 0

	prevPositiveCount := 0
	prevNegativeCount := 0
	positiveDir := true
	negativeDir := true

	for angle := 0; angle <= f.option.maxRotation; angle += f.option.incrStep {
		if positiveDir {
			rotatedImg := f.rotateImage(src, angle)
			if rotatedImg != nil {
				emptyLineCount := f.getEmptyLineCount(rotatedImg)
				if emptyLineCount > maxEmptyLineCount {
					maxEmptyLineCount = emptyLineCount
					detectedAngle = angle
				}

				if emptyLineCount < prevPositiveCount {
					positiveDir = false
				}
			}
		}

		if negativeDir {
			rotatedImg := f.rotateImage(src, -angle)
			if rotatedImg != nil {
				emptyLineCount := f.getEmptyLineCount(rotatedImg)
				if emptyLineCount > maxEmptyLineCount {
					maxEmptyLineCount = emptyLineCount
					detectedAngle = angle
				}

				if emptyLineCount < prevNegativeCount {
					negativeDir = false
				}
			}
		}
	}

	return detectedAngle
}

// Rotate image
func (f DeskewFilter) rotateImage(src image.Image, angle float32) image.Image {
	bounds := src.Bounds()
	width, height := calcRotatedSize(bounds.Dx(), bounds.Dy(), angle)
	dest := image.NewRGBA(image.Rect(0, 0, width, height))
	rotateFilter := f.getRotateFilter(angle)
	if rotateFilter == nil {
		return nil
	}

	rotateFilter.Draw(dest, src)
	return dest
}

// Get rotate filter by angle
func (f DeskewFilter) getRotateFilter(angle float32) *gift.GIFT {
	if rotate, ok := f.rotateMap[angle]; ok {
		return rotate
	} else {
		fmt.Errorf("Cannot find rotate filter of angle %v\n", angle)
		return nil
	}
}

// calculate dot count of each horizontal line
func calcDotCounts(img image.Image, threshold uint32) []uint32 {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	dotCounts := make([]uint32, h)
	for y := 0; y < h; y++ {
		dotCount := 0
		for x := 0; x < w; x++ {
			if r, _, _, _ := img.At(x, y).RGBA(); r > threshold {
				dotCount++
			}
		}
		dotCounts[y] = dotCount
	}
	return dotCounts
}

// get empty horizontal line count
func (f DeskewFilter) getEmptyLineCount(img image.Image) uint32 {
	dotCounts := calcDotCounts(img, f.option.threshold)

	emptyLineCount := 0
	for dotCount := range dotCounts {
		if dotCount <= f.option.emptyLineMinDotCount {
			emptyLineCount++
		}
	}
	return emptyLineCount
}

// calculate width/height after rotation
func calcRotatedSize(w, h int, angle float32) (int, int) {
	if w <= 0 || h <= 0 {
		return 0, 0
	}

	xoff := float32(w) / 2 - 0.5
	yoff := float32(h) / 2 - 0.5

	asin, acos := sincosf32(angle)
	x1, y1 := rotatePoint(0 - xoff, 0 - yoff, asin, acos)
	x2, y2 := rotatePoint(float32(w - 1) - xoff, 0 - yoff, asin, acos)
	x3, y3 := rotatePoint(float32(w - 1) - xoff, float32(h - 1) - yoff, asin, acos)
	x4, y4 := rotatePoint(0 - xoff, float32(h - 1) - yoff, asin, acos)

	minx := minf32(x1, minf32(x2, minf32(x3, x4)))
	maxx := maxf32(x1, maxf32(x2, maxf32(x3, x4)))
	miny := minf32(y1, minf32(y2, minf32(y3, y4)))
	maxy := maxf32(y1, maxf32(y2, maxf32(y3, y4)))

	neww := maxx - minx + 1
	if neww - floorf32(neww) > 0.01 {
		neww += 2
	}
	newh := maxy - miny + 1
	if newh - floorf32(newh) > 0.01 {
		newh += 2
	}
	return int(neww), int(newh)
}

func rotatePoint(x, y, asin, acos float32) (float32, float32) {
	newx := x * acos - y * asin
	newy := x * asin + y * acos
	return newx, newy
}

func minf32(x, y float32) float32 {
	if x < y {
		return x
	}
	return y
}
func maxf32(x, y float32) float32 {
	if x > y {
		return x
	}
	return y
}
func sincosf32(a float32) (float32, float32) {
	sin, cos := math.Sincos(math.Pi * float64(a) / 180)
	return float32(sin), float32(cos)
}
func floorf32(x float32) float32 {
	return float32(math.Floor(float64(x)))
}

