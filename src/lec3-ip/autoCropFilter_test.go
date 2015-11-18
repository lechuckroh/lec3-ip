package main
import (
	"testing"
	"image/color"
	"log"
)

func TestAutoCrop(t *testing.T) {
	cases := []struct {
		option         AutoCropOption
		expectedWidth  int
		expectedHeight int
	}{
		{
			AutoCropOption{
				threshold: 128,
				minRatio: 1.0, maxRatio: 3.0,
				maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
				marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
			},
			120, // 200 - (50 - 10) * 2
			270, // 350 - (50 - 10) * 2
		},
		{
			// constraint maxRatio
			AutoCropOption{
				threshold: 128,
				minRatio: 1.0, maxRatio: 2.0,
				maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
				marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
			},
			135, // max (200 - (50 - 10) * 2, 270 * 0.5)
			270, // 350 - (50 - 10) * 2
		},
	}

	for idx, c := range cases {
		// Create Image (200x350)
		// Rectangle (100x250)
		width, height, margin := 200, 350, 50
		bgColor := color.White
		rectColor := color.Black
		srcImg := CreateImageWithRect(width, height, margin, margin, width - margin, height - margin, bgColor, rectColor)

		// Run Filter
		option := c.option
		result := NewAutoCropFilter(option).Run(NewFilterSource(srcImg, "filename"))
		resultRect := result.(AutoCropResult).rect
		log.Printf("result rect = %v\n", resultRect)

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
