// Package samepic provides a suite of tools for detecting
// if two pictures contain the same subject.
package samepic

import "image"

// A Samer estimates whether or not two images are of the
// same subject.
type Samer interface {
	Same(img1, img2 image.Image) bool
}
