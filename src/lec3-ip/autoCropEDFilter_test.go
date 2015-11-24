package main
import (
	"testing"
	"image/color"
	"log"
	"image"
)

//func TestAutoCropED(t *testing.T) {
//	cases := []struct {
//		width          int
//		height         int
//		x1             int
//		y1             int
//		x2             int
//		y2             int
//		option         AutoCropEDOption
//		expectedWidth  int
//		expectedHeight int
//	}{
//		{
//			200, 350, 50, 50, 200 - 50, 350 - 50,
//			AutoCropEDOption{
//				threshold: 100,
//				minRatio: 1.0, maxRatio: 3.0,
//				maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
//				marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
//			},
//			120, // 200 - (50 - 10) * 2
//			270, // 350 - (50 - 10) * 2
//		},
//		{
//			// constraint maxRatio
//			200, 350, 50, 50, 200 - 50, 350 - 50,
//			AutoCropEDOption{
//				threshold: 100,
//				minRatio: 1.0, maxRatio: 2.0,
//				maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
//				marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
//			},
//			135, // max (200 - (50 - 10) * 2, 200 * 0.5)
//			270, // 350 - (50 - 10) * 2
//		},
//		{
//			// constraint maxCropRatio
//			200, 350, 50, 250, 200 - 50, 350 - 50,
//			AutoCropEDOption{
//				threshold: 100,
//				minRatio: 1.0, maxRatio: 2.0,
//				maxWidthCropRate: 0.1, maxHeightCropRate: 0.1,
//				marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
//			},
//			180, // max (200 - (50 - 10) * 2, 200 * 0.9)
//			315, // max (350 - (250 - 10 + 50 - 10), 350 * 0.9)
//		},
//		{
//			// constraint maxCropRatio
//			200, 350, 140, 50, 200 - 50, 350 - 50,
//			AutoCropEDOption{
//				threshold: 100,
//				minRatio: 1.0, maxRatio: 2.0,
//				maxWidthCropRate: 0.1, maxHeightCropRate: 0.2,
//				marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
//			},
//			180, // max (200 - (140 - 10 + 50 - 10), 200 * 0.9)
//			280, // max (350 - (50 - 10) * 2, 350 * 0.8)
//		},
//	}
//
//	for idx, c := range cases {
//		// Create Image (200x350)
//		// Rectangle (100x250)
//		bgColor := color.White
//		rectColor := color.Black
//		srcImg := CreateImageWithRect(c.width, c.height, c.x1, c.y1, c.x2, c.y2, bgColor, rectColor)
//
//		// Run Filter
//		option := c.option
//		result := NewAutoCropEDFilter(option).Run(NewFilterSource(srcImg, "filename"))
//		resultRect := result.(AutoCropEDResult).rect
//		log.Printf("result rect = %v\n", resultRect)
//
//		// Test result image size
//		destBounds := result.Image().Bounds()
//		if destBounds.Dx() != c.expectedWidth {
//			t.Errorf("[%v] width mismatch. exepcted=%v, actual=%v", idx, c.expectedWidth, destBounds.Dx())
//		}
//		if destBounds.Dy() != c.expectedHeight {
//			t.Errorf("[%v] height mismatch. exepcted=%v, actual=%v", idx, c.expectedHeight, destBounds.Dy())
//		}
//	}
//}

func testAutoCropED(t *testing.T, img image.Image, option AutoCropEDOption, expectedWidth, expectedHeight int) {
	// Run Filter
	result := NewAutoCropEDFilter(option).Run(NewFilterSource(img, "filename"))

	// Test result image size
	destBounds := result.Image().Bounds()
	widthMatch := destBounds.Dx() == expectedWidth
	heightMatch := destBounds.Dy() == expectedHeight

	if !widthMatch || !heightMatch {
		resultRect := result.(AutoCropEDResult).rect
		log.Printf("result rect = %v\n", resultRect)
	}

	if !widthMatch {
		t.Errorf("width mismatch. exepcted=%v, actual=%v", expectedWidth, destBounds.Dx())
	}
	if !heightMatch {
		t.Errorf("height mismatch. exepcted=%v, actual=%v", expectedHeight, destBounds.Dy())
	}
}

func TestAutoCropEDMargin(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCropED(t, img, AutoCropEDOption{
		threshold: 100,
		minRatio: 1.0, maxRatio: 3.0,
		maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
		marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
	},
		120, // max(width - leftSpace - rightSpace + marginLeft + marginRight, width * maxWidthCropRate)
		270, // max(height - topSpace - bottomSpace + marginTop + marginBottom, height * maxHeightCropRate)
	)
}

func TestAutoCropEDMaxRatio(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCropED(t, img, AutoCropEDOption{
		threshold: 100,
		minRatio: 1.0, maxRatio: 2.0,
		maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
		marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
	},
		135, // max(120, expectedHeight / maxRatio)
		270, // 270
	)
}

func TestAutoCropEDMaxCropRatio(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCropED(t, img, AutoCropEDOption{
		threshold: 100,
		minRatio: 1.0, maxRatio: 2.0,
		maxWidthCropRate: 0.1, maxHeightCropRate: 0.2,
		marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
	},
		180, // max(width - leftSpace - rightSpace + marginLeft + marginRight, width * maxWidthCropRate)
		280, // max(height - topSpace - bottomSpace + marginTop + marginBottom, height * maxHeightCropRate)
	)
}

func TestAutoCropEDInnerDetectionPadding1(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)
	DrawLine(img, 0, 1, 200, 1, color.Black)
	DrawLine(img, 25, 0, 25, 350, color.Black)

	testAutoCropED(t, img, AutoCropEDOption{
		threshold: 100,
		minRatio: 1.0, maxRatio: 10.0,
		maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
		marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
		paddingTop: 25,
	},
		145,
		350,
	)
}

func TestAutoCropEDInnerDetectionPadding2(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)
	DrawLine(img, 0, 20, 200, 20, color.Black)
	DrawLine(img, 10, 0, 10, 350, color.Black)

	testAutoCropED(t, img, AutoCropEDOption{
		threshold: 100,
		minRatio: 1.0, maxRatio: 10.0,
		maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
		marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
		paddingTop: 21, paddingLeft: 11,
	},
		120,
		270,
	)
}
