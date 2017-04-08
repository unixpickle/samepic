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
	return hashMatchRatio(hash1, hash2) >= a.threshold()
}

// SameBatch finds pairs of near duplicates.
func (a *AverageHash) SameBatch(images <-chan *IDImage) <-chan *Pair {
	res := make(chan *Pair, 1)
	go func() {
		defer close(res)
		ids := []interface{}{}
		hashes := [][]bool{}
		for image := range images {
			hash := a.Hash(image.Image)
			for i, hash1 := range hashes {
				if hashMatchRatio(hash, hash1) >= a.threshold() {
					res <- &Pair{ids[i], image.ID}
				}
			}
			ids = append(ids, image.ID)
			hashes = append(hashes, hash)
		}
	}()
	return res
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

func (a *AverageHash) threshold() float64 {
	if a.Threshold == 0 {
		return DefaultAverageHashThreshold
	} else {
		return a.Threshold
	}
}

func hashMatchRatio(h1, h2 []bool) float64 {
	var matchCount int
	for i, x := range h1 {
		if x == h2[i] {
			matchCount++
		}
	}
	return float64(matchCount) / float64(len(h1))
}
