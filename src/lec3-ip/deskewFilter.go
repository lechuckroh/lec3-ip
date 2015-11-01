package main
import (
	"image"
	"github.com/disintegration/gift"
	"image/color"
	"fmt"
)

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type DeskewOption struct {
	maxRotation          float32 // max rotation angle (0 <= value <= 360)
	incrStep             float32 // rotation angle increment step (0 <= value <= 360)
	emptyLineMinDotCount int
	debugOutputDir       string
	debugMode            bool
	threshold            uint8   // min brightness of space (0~255)
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
		fmt.Printf("%v : rotated angle=%v", r.filename, r.rotatedAngle)
	}
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

type DeskewFilter struct {
	rotateMap map[float32]*gift.GIFT
	option    DeskewOption
}

// Create DeskewFilter instance
func NewDeskewFilter(option DeskewOption) DeskewFilter {
	rotateBgColor := color.White

	rotateMap := make(map[float32]*gift.GIFT)
	for angle := float32(0); angle < option.maxRotation; angle += option.incrStep {
		rotateMap[angle] = gift.New(gift.Rotate(angle, rotateBgColor, gift.CubicInterpolation))
		rotateMap[-angle] = gift.New(gift.Rotate(-angle, rotateBgColor, gift.CubicInterpolation))
	}

	return DeskewFilter{rotateMap, option}
}


// Implements Filter.Run()
func (f DeskewFilter) Run(s *FilterSource) FilterResult {
	resultImage, rotatedAngle := f.run(s.image, s.filename)
	return DeskewResult{resultImage, s.filename, rotatedAngle}
}

// actual deskew implementation
func (f DeskewFilter) run(src image.Image, name string) (image.Image, float32) {
	if angle := f.detectRotationAngle(src, name); angle != 0 {
		return f.rotateImage(src, angle), angle
	}
	return src, 0
}

func (f DeskewFilter) detectRotationAngle(src image.Image, name string) float32 {
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

	if detectedAngle != 0 {
		fmt.Printf("detected angle %v\n", detectedAngle)
	}

	return detectedAngle
}

func (f DeskewFilter) calcNonEmptyLineCount(src image.Image, angle float32, name string) int {
	rotatedImg := f.rotateImage(src, angle)
	count := -1
	if rotatedImg != nil {
		count = f.getNonEmptyLineCount(rotatedImg)
	}

	// save debug images
	if (f.option.debugMode) {
		filename := fmt.Sprintf("%v_%v.jpg", name, angle)
		fmt.Printf("angle=%v, nonEmptyLineCount=%v\n", angle, count)
		if err := SaveJpeg(rotatedImg, f.option.debugOutputDir, filename, 80); err != nil {
			fmt.Errorf("Error : %v : %v\n", "", err)
		}
	}

	return count
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

// Rotate image
func (f DeskewFilter) rotateImage(src image.Image, angle float32) image.Image {
	bounds := src.Bounds()
	width, height := CalcRotatedSize(bounds.Dx(), bounds.Dy(), angle)
	dest := image.NewRGBA(image.Rect(0, 0, width, height))
	rotateFilter := f.getRotateFilter(angle)
	if rotateFilter == nil {
		return nil
	}

	rotateFilter.Draw(dest, src)
	return dest
}

// calculate dot count of each horizontal line
func (f DeskewFilter) calcDotCounts(img image.Image) []int {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	threshold := uint32(f.option.threshold) * 256
	dotCounts := make([]int, h)
	for y := 0; y < h; y++ {
		dotCount := 0
		for x := 0; x < w; x++ {
			if r, g, b, _ := img.At(x, y).RGBA(); (r + g + b) / 3 < threshold {
				dotCount++
			}
		}
		dotCounts[y] = dotCount
	}
	return dotCounts
}

// get empty horizontal line count
func (f DeskewFilter) getNonEmptyLineCount(img image.Image) int {
	dotCounts := f.calcDotCounts(img)
	nonEmptyLineCount := 0
	for _, dotCount := range dotCounts {
		if dotCount > f.option.emptyLineMinDotCount {
			nonEmptyLineCount++
		}
	}
	return nonEmptyLineCount
}
