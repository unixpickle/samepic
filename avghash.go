package samepic

import (
	"image"
	"image/color"

	"github.com/nfnt/resize"
)

const (
	DefaultAverageHashScaleSize = 8
	DefaultAverageHashThreshold = 0.9
)

// AverageHash compares two images using the average hash
// algorithm described on
// http://www.hackerfactor.com/blog/index.php?/archives/432-Looks-Like-It.html.
type AverageHash struct {
	// ScaleSize is the size to which images are scaled
	// before running the algorithm.
	// For example, if this is 8, images are scaled to 8x8.
	//
	// If this is 0, DefaultAverageHashScaleSize is used.
	ScaleSize int

	// Threshold is the minimum fraction of hash bits that
	// must match for two images to be considered the same.
	//
	// If this is 0, DefaultAverageHashThreshold is used.
	Threshold float64
}

// Same computes the hashes of two images and uses the
// results to determine if the images are the same.
func (a *AverageHash) Same(img1, img2 image.Image) bool {
	hash1 := a.Hash(img1)
	hash2 := a.Hash(img2)

	var matchCount int
	for i, x := range hash1 {
		if x == hash2[i] {
			matchCount++
		}
	}

	matchRatio := float64(matchCount) / float64(len(hash1))
	thresh := a.Threshold
	if thresh == 0 {
		thresh = DefaultAverageHashThreshold
	}
	return matchRatio >= thresh
}

// Hash creates the perceptual hash of an image.
func (a *AverageHash) Hash(img image.Image) []bool {
	scaleSize := uint(a.ScaleSize)
	if scaleSize == 0 {
		scaleSize = DefaultAverageHashScaleSize
	}
	scaled := resize.Resize(scaleSize, scaleSize, img, resize.Bilinear)

	brightnesses := make([]float64, 0, scaleSize*scaleSize)
	var sum float64
	for y := scaled.Bounds().Min.Y; y < scaled.Bounds().Max.Y; y++ {
		for x := scaled.Bounds().Min.X; x < scaled.Bounds().Max.X; x++ {
			gray := color.GrayModel.Convert(scaled.At(x, y))
			r, _, _, _ := gray.RGBA()
			b := float64(r) / 0xffff
			brightnesses = append(brightnesses, b)
			sum += b
		}
	}
	mean := sum / float64(len(brightnesses))
	res := make([]bool, len(brightnesses))
	for i, x := range brightnesses {
		res[i] = x > mean
	}
	return res
}
