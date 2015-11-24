package main
import (
	"testing"
	"image/color"
	"log"
	"image"
)

func testAutoCrop(t *testing.T, img image.Image, option AutoCropOption, expectedWidth, expectedHeight int) {
	// Run Filter
	result := NewAutoCropFilter(option).Run(NewFilterSource(img, "filename"))

	// Test result image size
	destBounds := result.Image().Bounds()
	widthMatch := destBounds.Dx() == expectedWidth
	heightMatch := destBounds.Dy() == expectedHeight

	if !widthMatch || !heightMatch {
		resultRect := result.(AutoCropResult).rect
		log.Printf("result rect = %v\n", resultRect)
	}

	if !widthMatch {
		t.Errorf("width mismatch. exepcted=%v, actual=%v", expectedWidth, destBounds.Dx())
	}
	if !heightMatch {
		t.Errorf("height mismatch. exepcted=%v, actual=%v", expectedHeight, destBounds.Dy())
	}
}

func TestAutoCropMargin(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCrop(t, img, AutoCropOption{
		threshold: 128,
		minRatio: 1.0, maxRatio: 3.0,
		maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
		marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
	},
		120, // max(width - leftSpace - rightSpace + marginLeft + marginRight, width * maxWidthCropRate)
		270, // max(height - topSpace - bottomSpace + marginTop + marginBottom, height * maxHeightCropRate)
	)
}

func TestAutoCropMaxRatio(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCrop(t, img, AutoCropOption{
		threshold: 128,
		minRatio: 1.0, maxRatio: 2.0,
		maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
		marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
	},
		135, // max(120, expectedHeight / maxRatio)
		270, // 270
	)
}

func TestAutoCropMaxCropRatio(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCrop(t, img, AutoCropOption{
		threshold: 100,
		minRatio: 1.0, maxRatio: 2.0,
		maxWidthCropRate: 0.1, maxHeightCropRate: 0.2,
		marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
	},
		180, // max(width - leftSpace - rightSpace + marginLeft + marginRight, width * maxWidthCropRate)
		280, // max(height - topSpace - bottomSpace + marginTop + marginBottom, height * maxHeightCropRate)
	)
}

func TestAutoCropInnerDetectionPadding1(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)
	DrawLine(img, 0, 1, 200, 1, color.Black)
	DrawLine(img, 25, 0, 25, 350, color.Black)

	testAutoCrop(t, img, AutoCropOption{
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

func TestAutoCropInnerDetectionPadding2(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)
	DrawLine(img, 0, 20, 200, 20, color.Black)
	DrawLine(img, 10, 0, 10, 350, color.Black)

	testAutoCrop(t, img, AutoCropOption{
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
