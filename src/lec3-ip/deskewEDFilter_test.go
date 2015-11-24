package main
import (
	"testing"
	"image/color"
	"image"
)

func testDeskewED(t *testing.T, img image.Image, option DeskewEDOption, rotatedAngleMin, rotatedAngleMax float32) {
	// Run Filter
	result := NewDeskewEDFilter(option).Run(NewFilterSource(img, "filename"))
	rotatedAngle := result.(DeskewEDResult).rotatedAngle

	// Test result image size
	if ! InRange(rotatedAngle, rotatedAngleMin, rotatedAngleMax) {
		t.Errorf("angle mismatch. exepcted=(%v ~ %v), actual=%v", rotatedAngleMin, rotatedAngleMax, rotatedAngle)
	}
}

func TestDeskewEDCCW(t *testing.T) {
	img := CreateImage(800, 1200, color.White)
	FillRect(img, 50, 50, 750, 1150, color.Black)
	rotatedImg := RotateImage(img, -1.4, color.White)

	// Run Filter
	option := DeskewEDOption{
		maxRotation: 2,
		incrStep: 0.2,
		threshold: 100,
		emptyLineMinDotCount: 0,
	}
	testDeskewED(t, rotatedImg, option, 1.2, 1.6)
}

func TestDeskewEDCW(t *testing.T) {
	img := CreateImage(800, 1200, color.White)
	FillRect(img, 50, 50, 750, 1150, color.Black)
	rotatedImg := RotateImage(img, 1.4, color.White)

	// Run Filter
	option := DeskewEDOption{
		maxRotation: 2,
		incrStep: 0.2,
		threshold: 100,
		emptyLineMinDotCount: 0,
	}
	testDeskewED(t, rotatedImg, option, -1.6, -1.2)
}
