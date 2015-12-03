package main
import (
	"image"
	"github.com/disintegration/gift"
	"image/color"
	"image/draw"
	"github.com/mitchellh/mapstructure"
	"log"
)

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type DeskewOption struct {
	maxRotation          float32 // max rotation angle (0 <= value <= 360)
	incrStep             float32 // rotation angle increment step (0 <= value <= 360)
	emptyLineMaxDotCount int
	debugOutputDir       string
	debugMode            bool
	threshold            uint8   // min brightness of space (0~255)
}

func NewDeskewOption(m map[string]interface{}) (*DeskewOption, error) {
	option := DeskewOption{}

	err := mapstructure.Decode(m, &option)
	if err != nil {
		return nil, err
	}

	return &option, nil
}


type DeskewResult struct {
	image        image.Image
	filename     string
	rotatedAngle float32
}

func (r DeskewResult) Image() image.Image {
	return r.image
}

func (r DeskewResult) Log() {
	if r.rotatedAngle != 0 {
		log.Printf("%v : rotated angle=%v\n", r.filename, r.rotatedAngle)
	}
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

type DeskewFilter struct {
	option DeskewOption
}

// Create DeskewFilter instance
func NewDeskewFilter(option DeskewOption) *DeskewFilter {
	return &DeskewFilter{option}
}

// Implements Filter.Run()
func (f DeskewFilter) Run(s *FilterSource) FilterResult {
	resultImage, rotatedAngle := f.run(s.image, s.filename)
	return DeskewResult{resultImage, s.filename, rotatedAngle}
}

// actual deskew implementation
func (f DeskewFilter) run(src image.Image, name string) (image.Image, float32) {
	bounds := src.Bounds()
	var rgba *image.RGBA

	switch src.(type) {
	case *image.RGBA:
		rgba = src.(*image.RGBA)
	default:
		rgba = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
		draw.Draw(rgba, bounds, src, bounds.Min, draw.Src)
	}

	if angle := f.detectAngle(rgba, name); angle != 0 {
		return f.rotateImage(rgba, angle), angle
	}
	return src, 0
}

// Rotate image
func (f DeskewFilter) rotateImage(src image.Image, angle float32) image.Image {
	bounds := src.Bounds()
	width, height := CalcRotatedSize(bounds.Dx(), bounds.Dy(), angle)
	dest := image.NewRGBA(image.Rect(0, 0, width, height))
	rotateFilter := gift.New(gift.Rotate(angle, color.White, gift.CubicInterpolation))
	rotateFilter.Draw(dest, src)
	return dest
}

func (f DeskewFilter) detectAngle(src *image.RGBA, name string) float32 {
	minNonEmptyLineCount := f.calcNonEmptyLineCount(src, 0, name)

	// increase rotation angle by incrStep
	detectedAngle := float32(0)

	prevPositiveCount := minNonEmptyLineCount
	prevNegativeCount := minNonEmptyLineCount
	positiveDir := true
	negativeDir := true

	for angle := f.option.incrStep; angle <= f.option.maxRotation; angle += f.option.incrStep {
		if positiveDir {
			nonEmptyLineCount := f.calcNonEmptyLineCount(src, angle, name)

			if nonEmptyLineCount <= minNonEmptyLineCount {
				minNonEmptyLineCount = nonEmptyLineCount
				detectedAngle = angle
			} else if nonEmptyLineCount > prevPositiveCount {
				positiveDir = false
			}
			prevPositiveCount = nonEmptyLineCount
		}

		if angle > 0 && negativeDir {
			nonEmptyLineCount := f.calcNonEmptyLineCount(src, -angle, name)

			if nonEmptyLineCount <= minNonEmptyLineCount {
				minNonEmptyLineCount = nonEmptyLineCount
				detectedAngle = -angle
			} else if nonEmptyLineCount > prevNegativeCount {
				negativeDir = false
			}
			prevNegativeCount = nonEmptyLineCount
		}
	}

	return detectedAngle
}

func (f DeskewFilter) calcNonEmptyLineCount(src *image.RGBA, angle float32, name string) int {
	dy, _ := Sincosf32(angle)
	bounds := src.Bounds()

	threshold := uint32(f.option.threshold) * 256
	nonEmptyLineCount := 0
	width, height := bounds.Dx(), bounds.Dy()
	for y := 0; y < height; y++ {
		yPos := float32(y)
		dotCount := 0

		for x := 0; x < width; x++ {
			yPosInt := int(yPos)
			if yPosInt < 0 || yPosInt >= height {
				break
			}

			if r, g, b, _ := src.At(x, yPosInt).RGBA(); (r + g + b) / 3 <= threshold {
				dotCount++
			}

			yPos += dy
		}

		if f.option.emptyLineMaxDotCount < dotCount {
			nonEmptyLineCount++
		}
	}

	if (f.option.debugMode) {
		log.Printf("angle=%v, nonEmptyLineCount=%v\n", angle, nonEmptyLineCount)
	}

	return nonEmptyLineCount
}
