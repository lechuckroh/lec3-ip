package main

import (
	"image"
	"image/color"
	"log"
	"testing"
)

func testAutoCropED(t *testing.T, img image.Image, option AutoCropEDOption, expectedWidth, expectedHeight, allowedDelta int) {
	// Run Filter
	result := NewAutoCropEDFilter(option).Run(NewFilterSource(img, "filename"))

	// Test result image size
	destBounds := result.Image().Bounds()
	widthMatch := InRange(destBounds.Dx(), expectedWidth-allowedDelta, expectedWidth+allowedDelta)
	heightMatch := InRange(destBounds.Dy(), expectedHeight-allowedDelta, expectedHeight+allowedDelta)

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
		Threshold: 100,
		MinRatio:  1.0, MaxRatio: 3.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
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
		Threshold: 100,
		MinRatio:  1.0, MaxRatio: 2.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
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
		Threshold: 100,
		MinRatio:  1.0, MaxRatio: 2.0,
		MaxWidthCropRate: 0.1, MaxHeightCropRate: 0.2,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
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
		Threshold: 100,
		MinRatio:  1.0, MaxRatio: 10.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
		PaddingTop: 10, PaddingLeft: 25,
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
		Threshold: 100,
		MinRatio:  1.0, MaxRatio: 10.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
		PaddingTop: 22, PaddingLeft: 12,
	},
		120,
		270,
		2,
	)
}

func TestAutoCropEDMaxCrop(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCropED(t, img, AutoCropEDOption{
		Threshold: 128,
		MinRatio:  1.0, MaxRatio: 3.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
		MaxCropTop: 0, MaxCropBottom: 100, MaxCropLeft: 0, MaxCropRight: 100,
	},
		160, // max(width - rightSpace + marginRight, width * maxWidthCropRate)
		310, // max(height - bottomSpace + marginBottom, height * maxHeightCropRate)
		2,
	)
}
