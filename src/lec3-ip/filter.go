package main
import "image"


// ----------------------------------------------------------------------------
// Filter source
// ----------------------------------------------------------------------------
type FilterSource struct {
	image    image.Image
	filename string
}

func NewFilterSource(image image.Image, filename string) FilterSource {
	return FilterSource{image, filename}
}


// ----------------------------------------------------------------------------
// Filter result
// ----------------------------------------------------------------------------
type FilterResult interface {
	Image() image.Image
	Print()
}

// ----------------------------------------------------------------------------
// Filter interface
// ----------------------------------------------------------------------------
type Filter interface {
	Run(src FilterSource) FilterResult
}
