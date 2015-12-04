package main

import (
	"image"
	"image/color"
	"testing"
)

func testDeskewED(t *testing.T, img image.Image, option DeskewEDOption, rotatedAngleMin, rotatedAngleMax float32) {
	// Run Filter
	result := NewDeskewEDFilter(option).Run(NewFilterSource(img, "filename"))
	rotatedAngle := result.(DeskewEDResult).rotatedAngle

	// Test result image size
	if !InRangef32(rotatedAngle, rotatedAngleMin, rotatedAngleMax) {
		t.Errorf("angle mismatch. exepcted=(%v ~ %v), actual=%v", rotatedAngleMin, rotatedAngleMax, rotatedAngle)
	}
}

func TestDeskewEDCCW(t *testing.T) {
	img := CreateImage(400, 700, color.White)
	FillRect(img, 50, 50, 350, 650, color.Black)
	rotatedImg := RotateImage(img, -1.4, color.White)

	// Run Filter
	option := DeskewEDOption{
		MaxRotation:          2,
		IncrStep:             0.2,
		Threshold:            100,
		EmptyLineMaxDotCount: 0,
	}
	testDeskewED(t, rotatedImg, option, 1.2, 1.6)
}

func TestDeskewEDCW(t *testing.T) {
	img := CreateImage(400, 700, color.White)
	FillRect(img, 50, 50, 350, 650, color.Black)
	rotatedImg := RotateImage(img, 1.4, color.White)

	// Run Filter
	option := DeskewEDOption{
		MaxRotation:          2,
		IncrStep:             0.2,
		Threshold:            100,
		EmptyLineMaxDotCount: 0,
	}
	testDeskewED(t, rotatedImg, option, -1.6, -1.2)
}
