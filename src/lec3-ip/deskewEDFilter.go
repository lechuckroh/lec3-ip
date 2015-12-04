package main

import (
	"github.com/disintegration/gift"
	"github.com/mitchellh/mapstructure"
	"image"
	"image/color"
	"log"
)

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type DeskewEDOption struct {
	MaxRotation          float32 // max rotation angle (0 <= value <= 360)
	IncrStep             float32 // rotation angle increment step (0 <= value <= 360)
	EmptyLineMaxDotCount int
	DebugMode            bool
	Threshold            uint8 // edge strength threshold (0~255(max edge))
}

func NewDeskewEDOption(m map[string]interface{}) (*DeskewEDOption, error) {
	option := DeskewEDOption{}

	err := mapstructure.Decode(m, &option)
	if err != nil {
		return nil, err
	}

	return &option, nil
}

type DeskewEDResult struct {
	image        image.Image
	filename     string
	rotatedAngle float32
}

func (r DeskewEDResult) Image() image.Image {
	return r.image
}

func (r DeskewEDResult) Log() {
	if r.rotatedAngle != 0 {
		log.Printf("%v : rotated angle=%v", r.filename, r.rotatedAngle)
	}
}

// ----------------------------------------------------------------------------
// EdgeDetectedDeskewFilter
// ----------------------------------------------------------------------------

type DeskewEDFilter struct {
	option     DeskewEDOption
	edgeDetect *gift.GIFT
}

// Create DeskewEDFilter instance
func NewDeskewEDFilter(option DeskewEDOption) *DeskewEDFilter {
	edgeDetect := gift.New(
		gift.Convolution(
			[]float32{
				-1, -1, -1,
				-1, 8, -1,
				-1, -1, -1,
			},
			false, false, false, 0.0,
		))

	return &DeskewEDFilter{
		option:     option,
		edgeDetect: edgeDetect,
	}
}

// Implements Filter.Run()
func (f DeskewEDFilter) Run(s *FilterSource) FilterResult {
	resultImage, rotatedAngle := f.run(s.image, s.filename)
	return DeskewEDResult{resultImage, s.filename, rotatedAngle}
}

// actual deskew implementation
func (f DeskewEDFilter) run(src image.Image, name string) (image.Image, float32) {
	// Edge Detect Image
	edImg := image.NewGray(src.Bounds())
	f.edgeDetect.Draw(edImg, src)

	// Find preferred rotation angle
	if angle := f.detectAngle(edImg, name); angle != 0 {
		return f.rotateImage(src, angle), angle
	}
	return src, 0
}

// Rotate image
func (f DeskewEDFilter) rotateImage(src image.Image, angle float32) image.Image {
	bounds := src.Bounds()
	width, height := CalcRotatedSize(bounds.Dx(), bounds.Dy(), angle)
	dest := image.NewRGBA(image.Rect(0, 0, width, height))
	rotateFilter := gift.New(gift.Rotate(angle, color.White, gift.CubicInterpolation))
	rotateFilter.Draw(dest, src)
	return dest
}

// Detect rotation angle
func (f DeskewEDFilter) detectAngle(edImg *image.Gray, name string) float32 {
	minNonEmptyLineCount := f.calcNonEmptyLineCount(edImg, 0, name)

	// increase rotation angle by incrStep
	detectedAngle := float32(0)

	prevPositiveCount := minNonEmptyLineCount
	prevNegativeCount := minNonEmptyLineCount
	positiveDir := true
	negativeDir := true

	incrStep := f.option.IncrStep
	if incrStep > 0 {
		for angle := incrStep; angle <= f.option.MaxRotation; angle += incrStep {
			if positiveDir {
				nonEmptyLineCount := f.calcNonEmptyLineCount(edImg, angle, name)

				if nonEmptyLineCount <= minNonEmptyLineCount {
					minNonEmptyLineCount = nonEmptyLineCount
					detectedAngle = angle
				} else if nonEmptyLineCount > prevPositiveCount {
					positiveDir = false
				}
				prevPositiveCount = nonEmptyLineCount
			}

			if angle > 0 && negativeDir {
				nonEmptyLineCount := f.calcNonEmptyLineCount(edImg, -angle, name)

				if nonEmptyLineCount <= minNonEmptyLineCount {
					minNonEmptyLineCount = nonEmptyLineCount
					detectedAngle = -angle
				} else if nonEmptyLineCount > prevNegativeCount {
					negativeDir = false
				}
				prevNegativeCount = nonEmptyLineCount
			}
		}
	}

	if detectedAngle != 0 {
		log.Printf("detected angle %v\n", detectedAngle)
	}

	return detectedAngle
}

func (f DeskewEDFilter) calcNonEmptyLineCount(edImg *image.Gray, angle float32, name string) int {
	dy, _ := Sincosf32(angle)
	bounds := edImg.Bounds()

	threshold := uint32(f.option.Threshold) * 256
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

			if r, _, _, _ := edImg.At(x, yPosInt).RGBA(); r >= threshold {
				dotCount++
			}

			yPos += dy
		}

		if f.option.EmptyLineMaxDotCount < dotCount {
			nonEmptyLineCount++
		}
	}

	if f.option.DebugMode {
		log.Printf("angle=%v, nonEmptyLineCount=%v\n", angle, nonEmptyLineCount)
	}

	return nonEmptyLineCount
}
