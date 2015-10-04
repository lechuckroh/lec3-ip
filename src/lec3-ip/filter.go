package main
import (
	"image"
)

type Filter interface {
	Run(src image.Image) image.Image
}
