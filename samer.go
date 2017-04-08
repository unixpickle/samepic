// Package samepic provides a suite of tools for detecting
// if two pictures contain the same subject.
package samepic

import "image"

// A Samer estimates whether or not two images are of the
// same subject.
type Samer interface {
	Same(img1, img2 image.Image) bool
}

// IDImage is an image paired with an indentifier.
// It is used by BatchSamer to identify images.
type IDImage struct {
	Image image.Image
	ID    interface{}
}

// A Pair represents the fact that two images (given by
// their IDs) are similar.
// The order of the IDs does not matter.
type Pair [2]interface{}

// A BatchSamer can find duplicate pairs of images in a
// stream of images.
//
// Exactly one *Pair should be produced per pair.
// Avoid producing identical or mirrored *Pair objects.
//
// A BatchSamer should be implemented asynchronously,
// allowing pairs to be streamed as they are found.
type BatchSamer interface {
	SameBatch(images <-chan *IDImage) <-chan *Pair
}
