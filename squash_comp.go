package samepic

import (
	"image"
	"math"

	"github.com/nfnt/resize"
	"github.com/unixpickle/num-analysis/linalg"
)

const (
	DefaultSquashCompMinOverlap = 0.7
	DefaultSquashCompVectorSize = 150
	DefaultSquashCompThreshold  = 0.995
)

type SquashAxis int

const (
	VerticalSquash SquashAxis = iota
	HorizontalSquash
)

// SquashComp compares images by "squashing" them into
// one-dimensional lines (by scaling the image) and
// measuring the correlation between the lines.
// It tries to be resilient to cropping and scaling by
// checking many different relative scales between the
// lines and correlating the shorter line at various
// offsets along the longer line.
type SquashComp struct {
	// Axis specifies which axis is squased for the
	// comparison.
	// The default is VerticalSquash, which squashes
	// the image along the y-axis.
	Axis SquashAxis

	// MinOverlap specifies the fraction of the bigger
	// line that must be covered by the smaller line
	// during correlation checking.
	// For example, a MinOverlap value of 0.8 means that
	// one image could be cropped such that 20% of its
	// unsquashed axis is cut off.
	//
	// If this is 0, DefaultSquashCompMinOverlap is used.
	MinOverlap float64

	// VectorSize is the size to which an image is scaled
	// in the non-squashed dimension.
	// By having a constant VectorSize, scaled images will
	// still match, since they are both scaled to the
	// VectorSize.
	//
	// If this is 0, DefaultSquashCompVectorSize
	// is used.
	VectorSize int

	// Threshold is the minimum correlation to trigger
	// a positive match.
	//
	// If this is 0, DefaultSquashCompThreshold is used.
	Threshold float64
}

// Same uses squashed correlations to determine if two
// images are the same.
func (s *SquashComp) Same(img1, img2 image.Image) bool {
	return s.asymmetricalSame(img1, img2) || s.asymmetricalSame(img2, img1)
}

// asymmetricalSame keeps the first image at exactly
// s.VectorSize and scales+translates the other one.
func (s *SquashComp) asymmetricalSame(img1, img2 image.Image) bool {
	vectorSize := s.VectorSize
	if vectorSize == 0 {
		vectorSize = DefaultSquashCompVectorSize
	}
	minOverlap := s.MinOverlap
	if minOverlap == 0 {
		minOverlap = DefaultSquashCompMinOverlap
	}

	mainVec := s.squash(img1, vectorSize)
	minSize := int(math.Ceil(float64(vectorSize) * minOverlap))

	for size := minSize; size <= vectorSize; size++ {
		secondaryVec := s.squash(img2, size)
		allowedMiss := size - minSize
		for x := -allowedMiss; x <= vectorSize-size+allowedMiss; x++ {
			if s.vectorMatch(mainVec, secondaryVec, x) {
				return true
			}
		}
	}

	return false
}

// squash generates a squashed image vector of size n*3,
// with packed components for R, G, and B.
func (s *SquashComp) squash(img image.Image, n int) linalg.Vector {
	res := make(linalg.Vector, 0, n*3)
	if s.Axis == HorizontalSquash {
		scaledImg := resize.Resize(1, uint(n), img, resize.Bilinear)
		x := scaledImg.Bounds().Min.X
		for y := scaledImg.Bounds().Min.Y; y < scaledImg.Bounds().Max.Y; y++ {
			r, g, b, _ := scaledImg.At(x, y).RGBA()
			res = append(res, float64(r)/0xffff)
			res = append(res, float64(g)/0xffff)
			res = append(res, float64(b)/0xffff)
		}
	} else {
		scaledImg := resize.Resize(uint(n), 1, img, resize.Bilinear)
		y := scaledImg.Bounds().Min.Y
		for x := scaledImg.Bounds().Min.X; x < scaledImg.Bounds().Max.X; x++ {
			r, g, b, _ := scaledImg.At(x, y).RGBA()
			res = append(res, float64(r)/0xffff)
			res = append(res, float64(g)/0xffff)
			res = append(res, float64(b)/0xffff)
		}
	}
	return res
}

// vectorMatch calculates the correlation of two vectors
// (one with an offset) and determines if the correlation
// exceeds the match threshold.
func (s *SquashComp) vectorMatch(v1, v2 linalg.Vector, v2Offset int) bool {
	if v2Offset < 0 {
		v2 = v2[-v2Offset:]
	} else {
		v1 = v1[v2Offset:]
		if len(v1) < len(v2) {
			v2 = v2[:len(v1)]
		}
	}
	v1 = v1[:len(v2)]
	cor := v1.Dot(v2) / (v1.Mag() * v2.Mag())

	if s.Threshold != 0 {
		return cor >= s.Threshold
	} else {
		return cor >= DefaultSquashCompThreshold
	}
}
