package main
import (
	"github.com/disintegration/gift"
	"image"
)

type EdgeDetectionFilter struct {
	g *gift.GIFT
}

func NewEdgeDetectionFilter() *EdgeDetectionFilter {
	return &EdgeDetectionFilter{gift.New(
		gift.Convolution(
			[]float32{
				-1, -1, -1,
				-1, 8, -1,
				-1, -1, -1,
			},
			false, false, false, 0.0,
		))}
}

func (f *EdgeDetectionFilter) run(src image.Image) image.Image {
	dest := image.NewRGBA(src.Bounds())
	f.g.Draw(dest, src)
	return dest
}
