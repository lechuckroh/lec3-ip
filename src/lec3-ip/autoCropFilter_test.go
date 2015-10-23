package main
import (
	"testing"
	"image"
	"image/color"
	"fmt"
)

// create image with colored rectangle
func createImageWithRect(width, height, x1, y1, x2, y2 int, color color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			img.Set(x, y, color)
		}
	}

	return img
}

func TestRun(t *testing.T) {
	cases := []struct {
		option         AutoCropOption
		expectedWidth  int
		expectedHeight int
	}{
		{
			AutoCropOption{
				threshold: 128,
				minRatio: 1.0,
				maxRatio: 3.0,
				maxWidthCropRate: 0.5,
				maxHeightCropRate: 0.5,
				marginTop: 10,
				marginBottom: 10,
				marginLeft: 10,
				marginRight: 10,
			}, 120, 270,
		},
		{	// constraint maxRatio
			AutoCropOption{
				threshold: 128,
				minRatio: 1.0,
				maxRatio: 2.0,
				maxWidthCropRate: 0.5,
				maxHeightCropRate: 0.5,
				marginTop: 10,
				marginBottom: 10,
				marginLeft: 10,
				marginRight: 10,
			}, 134, 270,
		},
	}

	for idx, c := range cases {
		// Create Image
		width, height, margin := 200, 350, 50
		col := color.RGBA{200, 200, 200, 255}
		srcImg := createImageWithRect(width, height, margin, margin, width - margin, height - margin, col)

		// Run Filter
		option := c.option
		result := NewAutoCropFilter(option).Run(NewFilterSource(srcImg, "filename"))
		resultRect := result.(AutoCropResult).rect
		fmt.Println(resultRect)

		// Test result image size
		destBounds := result.Image().Bounds()
		if destBounds.Dx() != c.expectedWidth {
			t.Errorf("[%v] width mismatch. exepcted=%v, actual=%v", idx, c.expectedWidth, destBounds.Dx())
		}
		if destBounds.Dy() != c.expectedHeight {
			t.Errorf("[%v] height mismatch. exepcted=%v, actual=%v", idx, c.expectedHeight, destBounds.Dy())
		}
	}
}
