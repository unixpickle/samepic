package samepic

import (
	"bytes"
	"image"
	"image/jpeg"
	"math"
	"math/rand"

	"github.com/nfnt/resize"
)

// DefaultManipulator is a manipulator which can be used
// to produce reasonable manipulations.
var DefaultManipulator Manipulator = &AggregateManipulator{
	Manipulators: []Manipulator{
		&Scale{
			MinScale: 0.5,
			MaxScale: 1.5,
		},
		&Crop{
			MinMajorKeep: 0.5,
			MinMinorKeep: 0.8,
		},
		&CompressJPEG{},
	},
	Probabilities: []float64{0.5, 0.5, 0.5},
}

// A Manipulator applies a realistic manipulation to
// an image, such as cropping, scaling, or compressing.
// A manipulation may be probabilistic, meaning it may
// do different things each time.
type Manipulator interface {
	Manipulate(img image.Image) image.Image
}

// A CompressJPEG manipulates images by compressing
// and then decompressing them with JPEG.
type CompressJPEG struct {
	// These parameters control the minimum and maximum
	// amount of JPEG compression, where 100 is top
	// quality and 1 is lowest quality.
	// If a parameter is 0, the appropriate bound (1 or
	// 100) is used.
	MinQuality int
	MaxQuality int
}

// Manipulate compresses the image with a random amount
// of compression and returns the lower quality image.
func (c *CompressJPEG) Manipulate(img image.Image) image.Image {
	min := c.MinQuality
	max := c.MaxQuality
	if min == 0 {
		min = 1
	}
	if max == 0 {
		max = 100
	}
	quality := rand.Intn(max-min+1) + min

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		panic(err)
	}
	res, err := jpeg.Decode(&buf)
	if err != nil {
		panic(err)
	}
	return res
}

// A Scale manipulates images by resizing them.
type Scale struct {
	// Interpolations stores the allowed interpolation
	// techniques for resizing.
	// If this is nil, a default list is used.
	Interpolations []resize.InterpolationFunction

	// These parameters determine the allowed range of
	// scale ratios, where the scale ratio measures the
	// new size divided by the old size.
	MinScale float64
	MaxScale float64
}

// Manipulate resizes the image to a random size in the
// allowed range of scale ratios.
func (s *Scale) Manipulate(img image.Image) image.Image {
	interps := s.Interpolations
	if interps == nil {
		interps = []resize.InterpolationFunction{
			resize.NearestNeighbor,
			resize.Bilinear,
			resize.Bicubic,
			resize.MitchellNetravali,
			resize.Lanczos2,
			resize.Lanczos3,
		}
	}
	interp := interps[rand.Intn(len(interps))]
	scale := rand.Float64()*(s.MaxScale-s.MinScale) + s.MinScale
	newWidth := uint(float64(img.Bounds().Dx())*scale + 0.5)
	return resize.Resize(newWidth, 0, img, interp)
}

// A Crop manipulates images by cropping out a random
// region from them.
type Crop struct {
	// These parameters determine the smallest region
	// that may be cropped.
	// The "major" axis is defined as the direction in
	// which the original image is larger (e.g. the x
	// axis for landscape images).
	// The "minor" axis is the opposite of the "major".
	// These two parameters determine how much of the
	// two axes must be retained in the cropped image.
	//
	// For example, if MinMinorKeep is 0.8, then the
	// cropping may never crop the image such that less
	// than 80% of the original image's minor axis is
	// kept.
	// With this setting, if an image was 1500x1000, 1000
	// is the size of the minor axis, and the height of
	// the new image couldn't be less than 800 pixels.
	MinMajorKeep float64
	MinMinorKeep float64
}

// Manipulate randomly crops the image according to the
// allowed boundaries.
func (c *Crop) Manipulate(img image.Image) image.Image {
	minorSize := math.Min(float64(img.Bounds().Dx()), float64(img.Bounds().Dy()))
	majorSize := math.Max(float64(img.Bounds().Dx()), float64(img.Bounds().Dy()))

	minorKeep := rand.Float64()*(1-c.MinMinorKeep) + c.MinMinorKeep
	majorKeep := rand.Float64()*(1-c.MinMajorKeep) + c.MinMajorKeep

	newMinor := int(minorSize*minorKeep + 0.5)
	newMajor := int(majorSize*majorKeep + 0.5)
	minorOffset := rand.Intn(int(minorSize) - newMinor + 1)
	majorOffset := rand.Intn(int(majorSize) - newMajor + 1)

	xMajor := img.Bounds().Dx() > img.Bounds().Dy()

	// Prevent bias towards either axis.
	if img.Bounds().Dx() == img.Bounds().Dy() {
		xMajor = rand.Intn(2) == 0
	}

	if xMajor {
		return cropImage(img, majorOffset, minorOffset, newMajor, newMinor)
	} else {
		return cropImage(img, minorOffset, majorOffset, newMinor, newMajor)
	}
}

func cropImage(img image.Image, x, y, width, height int) image.Image {
	newImage := image.NewRGBA(image.Rect(0, 0, width, height))
	for destY := 0; destY < height; destY++ {
		for destX := 0; destX < width; destX++ {
			newImage.Set(destX, destY, img.At(destX+x, destY+y))
		}
	}
	return newImage
}

// An AggregateManipulator probabilistically applies
// an assortment of Manipulators (in order) to images.
type AggregateManipulator struct {
	// Manipulators is the list of manipulators.
	// The order of this list determines the order
	// in which the manipulators may be applied.
	Manipulators []Manipulator

	// Probabilities stores one probability for each
	// Manipulator, indicating the likelihood of
	// applying that manipulator to a given image.
	Probabilities []float64
}

// Manipulate randomly applies the manipulators.
func (a *AggregateManipulator) Manipulate(img image.Image) image.Image {
	for i, manip := range a.Manipulators {
		prob := a.Probabilities[i]
		if rand.Float64() <= prob {
			img = manip.Manipulate(img)
		}
	}
	return img
}
