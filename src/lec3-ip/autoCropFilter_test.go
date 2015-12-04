package main

import (
	"image"
	"image/color"
	"log"
	"testing"
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
		Threshold: 128,
		MinRatio:  1.0, MaxRatio: 3.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
	},
		120, // max(width - leftSpace - rightSpace + marginLeft + marginRight, width * maxWidthCropRate)
		270, // max(height - topSpace - bottomSpace + marginTop + marginBottom, height * maxHeightCropRate)
	)
}

func TestAutoCropMaxRatio(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCrop(t, img, AutoCropOption{
		Threshold: 128,
		MinRatio:  1.0, MaxRatio: 2.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
	},
		135, // max(120, expectedHeight / maxRatio)
		270, // 270
	)
}

func TestAutoCropMaxCropRatio(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCrop(t, img, AutoCropOption{
		Threshold: 100,
		MinRatio:  1.0, MaxRatio: 2.0,
		MaxWidthCropRate: 0.1, MaxHeightCropRate: 0.2,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
	},
		180, // max(width - leftSpace - rightSpace + marginLeft + marginRight, width * maxWidthCropRate)
		280, // max(height - topSpace - bottomSpace + marginTop + marginBottom, height * maxHeightCropRate)
	)
}

func TestAutoCropInnerDetectionPadding1(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)
	DrawLine(img, 0, 0, 200, 0, color.Black)
	DrawLine(img, 25, 0, 25, 350, color.Black)

	testAutoCrop(t, img, AutoCropOption{
		Threshold: 100,
		MinRatio:  1.0, MaxRatio: 10.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
		PaddingTop: 10, PaddingLeft: 25,
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
		Threshold: 100,
		MinRatio:  1.0, MaxRatio: 10.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
		PaddingTop: 21, PaddingLeft: 11,
	},
		120,
		270,
	)
}

func TestAutoCropMaxCrop(t *testing.T) {
	img := CreateImage(200, 350, color.White)
	FillRect(img, 50, 50, 150, 300, color.Black)

	testAutoCrop(t, img, AutoCropOption{
		Threshold: 128,
		MinRatio:  1.0, MaxRatio: 3.0,
		MaxWidthCropRate: 0.5, MaxHeightCropRate: 0.5,
		MarginTop: 10, MarginBottom: 10, MarginLeft: 10, MarginRight: 10,
		MaxCropTop: 0, MaxCropBottom: 100, MaxCropLeft: 0, MaxCropRight: 100,
	},
		160, // max(width - rightSpace + marginRight, width * maxWidthCropRate)
		310, // max(height - bottomSpace + marginBottom, height * maxHeightCropRate)
	)
}
