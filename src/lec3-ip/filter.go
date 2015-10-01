package main
import (
	"image"
)

type Filter interface {
	run(src image.Image) image.Image
}
