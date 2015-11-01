package main
import (
	"testing"
	"image/color"
	"fmt"
)

func TestDeskew(t *testing.T) {
	cases := []struct {
		option          DeskewOption
		rotation        float32
		rotatedAngleMin float32
		rotatedAngleMax float32
	}{
		{
			DeskewOption{
				maxRotation: 2,
				incrStep: 0.2,
				threshold: 220,
				emptyLineMinDotCount: 0,
//				debugOutputDir: "test",
//				debugMode: true,
			}, -1.4, 1.2, 1.6,
		},
		{
			DeskewOption{
				maxRotation: 2,
				incrStep: 0.2,
				threshold: 220,
				emptyLineMinDotCount: 0,
//				debugMode: true,
			}, 1.4, -1.6, -1.2,
		},
	}

	epsilon := float32(0.00001)
	for idx, c := range cases {
		// Create Image (800x1200)
		width, height, margin := 800, 1200, 50
		bgColor := color.White
		rectColor := color.Black
		srcImg := CreateImageWithRect(width, height, margin, margin, width - margin, height - margin, bgColor, rectColor)

		// Rotate Image
		rotatedImg := RotateImage(srcImg, c.rotation, color.White)

		// Run Filter
		option := c.option
		result := NewDeskewFilter(option).Run(NewFilterSource(rotatedImg, fmt.Sprintf("filename%v", idx)))
		rotatedAngle := result.(DeskewResult).rotatedAngle

		// Test result image size
		if rotatedAngle + epsilon < c.rotatedAngleMin || rotatedAngle - epsilon > c.rotatedAngleMax {
			t.Errorf("[%v] angle mismatch. exepcted=(%v ~ %v), actual=%v",
				idx, c.rotatedAngleMin, c.rotatedAngleMax, rotatedAngle)
		}
	}
}
