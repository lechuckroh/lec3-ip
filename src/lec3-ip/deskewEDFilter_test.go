package main
import (
	"testing"
	"image/color"
	"fmt"
	"sync"
)

type DeskewEDCase struct {
	option          DeskewEDOption
	rotation        float32
	rotatedAngleMin float32
	rotatedAngleMax float32
}

func TestDeskewED(t *testing.T) {
	cases := []DeskewEDCase{
		DeskewEDCase{
			DeskewEDOption{
				maxRotation: 2,
				incrStep: 0.2,
				threshold: 100,
				emptyLineMinDotCount: 0,
				//debugMode: true,
			}, -1.4, 1.2, 1.6,
		},
		DeskewEDCase{
			DeskewEDOption{
				maxRotation: 2,
				incrStep: 0.2,
				threshold: 100,
				emptyLineMinDotCount: 0,
				//debugMode: true,
			}, 1.4, -1.6, -1.2,
		},
	}

	wg := sync.WaitGroup{}
	for idx, c := range cases {
		wg.Add(1)
		go testCase(t, idx, c, &wg)
	}
	wg.Wait()
}

func testCase(t *testing.T, idx int, c DeskewEDCase, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	epsilon := float32(0.00001)

	// Create Image (800x1200)
	width, height, margin := 800, 1200, 50
	bgColor := color.White
	rectColor := color.Black
	srcImg := CreateImageWithRect(width, height, margin, margin, width - margin, height - margin, bgColor, rectColor)

	// Rotate Image
	rotatedImg := RotateImage(srcImg, c.rotation, color.White)

	// Run Filter
	option := c.option
	result := NewDeskewEDFilter(option).Run(NewFilterSource(rotatedImg, fmt.Sprintf("filename%v", idx)))
	rotatedAngle := result.(DeskewEDResult).rotatedAngle

	// Test result image size
	if rotatedAngle + epsilon < c.rotatedAngleMin || rotatedAngle - epsilon > c.rotatedAngleMax {
		t.Errorf("[%v] angle mismatch. exepcted=(%v ~ %v), actual=%v",
			idx, c.rotatedAngleMin, c.rotatedAngleMax, rotatedAngle)
	}
}
