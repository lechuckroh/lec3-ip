package main
import (
	"testing"
	"image/color"
	"log"
	"image"
)

func testAutoCropED(t *testing.T, img image.Image, option AutoCropEDOption, expectedWidth, expectedHeight, allowedDelta int) {
	// Run Filter
	result := NewAutoCropEDFilter(option).Run(NewFilterSource(img, "filename"))

	// Test result image size
	destBounds := result.Image().Bounds()
	widthMatch := InRange(destBounds.Dx(), expectedWidth - allowedDelta, expectedWidth + allowedDelta)
	heightMatch := InRange(destBounds.Dy(), expectedHeight - allowedDelta, expectedHeight + allowedDelta)

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
		0,
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
		0,
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
		0,
	)
}

func TestAutoCropEDInnerDetectionPadding1(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)
	DrawLine(img, 0, 0, 200, 0, color.Black)
	DrawLine(img, 25, 0, 25, 350, color.Black)

	testAutoCropED(t, img, AutoCropEDOption{
		threshold: 100,
		minRatio: 1.0, maxRatio: 10.0,
		maxWidthCropRate: 0.5, maxHeightCropRate: 0.5,
		marginTop: 10, marginBottom: 10, marginLeft: 10, marginRight: 10,
		paddingTop: 10, paddingLeft: 25,
	},
		145,
		350,
		2,
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
		2,
	)
}
