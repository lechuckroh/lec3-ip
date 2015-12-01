package main
import (
	"testing"
	"image/color"
	"image"
)

func testDeskew(t *testing.T, img image.Image, option DeskewOption, rotatedAngleMin, rotatedAngleMax float32) {
	// Run Filter
	result := NewDeskewFilter(option).Run(NewFilterSource(img, "filename"))
	rotatedAngle := result.(DeskewResult).rotatedAngle

	// Test result image size
	if ! InRangef32(rotatedAngle, rotatedAngleMin, rotatedAngleMax) {
		t.Errorf("angle mismatch. exepcted=(%v ~ %v), actual=%v", rotatedAngleMin, rotatedAngleMax, rotatedAngle)
	}
}

func TestDeskewCCW(t *testing.T) {
	img := CreateImage(400, 700, color.White)
	FillRect(img, 50, 50, 350, 650, color.Black)
	rotatedImg := RotateImage(img, -1.4, color.White)

	// Run Filter
	option := DeskewOption{
		maxRotation: 2,
		incrStep: 0.2,
		threshold: 220,
		emptyLineMinDotCount: 0,
	}
	testDeskew(t, rotatedImg, option, 1.2, 1.6)
}

func TestDeskewCW(t *testing.T) {
	img := CreateImage(400, 700, color.White)
	FillRect(img, 50, 50, 550, 650, color.Black)
	rotatedImg := RotateImage(img, 1.4, color.White)

	// Run Filter
	option := DeskewOption{
		maxRotation: 2,
		incrStep: 0.2,
		threshold: 220,
		emptyLineMinDotCount: 0,
	}
	testDeskew(t, rotatedImg, option, -1.6, -1.2)
}

