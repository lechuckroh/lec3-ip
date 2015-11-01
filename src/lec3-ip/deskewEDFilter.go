package main
import (
	"image"
	"github.com/disintegration/gift"
	"image/color"
	"fmt"
)

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------
type DeskewEDOption struct {
	maxRotation          float32 // max rotation angle (0 <= value <= 360)
	incrStep             float32 // rotation angle increment step (0 <= value <= 360)
	emptyLineMinDotCount int
	debugOutputDir       string
	debugMode            bool
	threshold            uint8   // edge strength threshold (0~255(max edge))
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
		fmt.Printf("%v : rotated angle=%v", r.filename, r.rotatedAngle)
	}
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

// Get rotate filter by angle
func (f DeskewEDFilter) getRotateFilter(angle float32) *gift.GIFT {
	if rotate, ok := f.rotateMap[angle]; ok {
		return rotate
	} else {
		fmt.Errorf("Cannot find rotate filter of angle %v\n", angle)
		return nil
	}
}

// Rotate image
func (f DeskewEDFilter) rotateImage(src image.Image, angle float32) image.Image {
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

// ----------------------------------------------------------------------------
// EdgeDetectedDeskewFilter
// ----------------------------------------------------------------------------

type DeskewEDFilter struct {
	rotateMap map[float32]*gift.GIFT
	option    DeskewEDOption
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
	rotateBgColor := color.Black

	rotateMap := make(map[float32]*gift.GIFT)
	for angle := float32(0); angle < option.maxRotation; angle += option.incrStep {
		rotateMap[angle] = gift.New(gift.Rotate(angle, rotateBgColor, gift.CubicInterpolation))
		rotateMap[-angle] = gift.New(gift.Rotate(-angle, rotateBgColor, gift.CubicInterpolation))
	}

	return &DeskewEDFilter{
		rotateMap: rotateMap,
		option: option,
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
	// Edge Detect
	edgeDetectedImg := image.NewGray(src.Bounds())
	f.edgeDetect.Draw(edgeDetectedImg, src)

	// Find preferred rotation angle
	if angle := f.detectRotationAngle(edgeDetectedImg, name); angle != 0 {
		return f.rotateImage(src, angle), angle
	}
	return src, 0
}

func (f DeskewEDFilter) detectRotationAngle(src image.Image, name string) float32 {
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

func (f DeskewEDFilter) calcNonEmptyLineCount(src image.Image, angle float32, name string) int {
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

// get empty horizontal line count
func (f DeskewEDFilter) getNonEmptyLineCount(img image.Image) int {
	dotCounts := f.calcDotCounts(img, uint32(f.option.threshold) * 256)
	nonEmptyLineCount := 0
	for _, dotCount := range dotCounts {
		if dotCount > f.option.emptyLineMinDotCount {
			nonEmptyLineCount++
		}
	}
	return nonEmptyLineCount
}

// calculate dot count of each horizontal line (edge detected image)
func (f DeskewEDFilter) calcDotCounts(img image.Image, threshold uint32) []int {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	dotCounts := make([]int, h)
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
