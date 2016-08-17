package samepic

import (
	"errors"
	"image"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
)

// Samples represents any source of image samples.
type Samples interface {
	// Random randomly selects a sample image.
	Random() (image.Image, error)

	// RandomPair randomly selects a pair of sample
	// images without replacement (i.e. the images
	// will be different)
	RandomPair() (image.Image, image.Image, error)
}

// DirSamples loads image samples from a directory
// of image files.
type DirSamples struct {
	imagePaths []string
}

// NewDirSamples creates a DirSamples instance by
// reading a directory listing.
func NewDirSamples(dir string) (*DirSamples, error) {
	listing, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(listing))
	for _, l := range listing {
		if !l.IsDir() {
			paths = append(paths, filepath.Join(dir, l.Name()))
		}
	}
	return &DirSamples{imagePaths: paths}, nil
}

// Random selects a random image from the directory
// and loads it.
// If an image fails to load, another one is tried,
// and that process is repeated either until an image
// is successfully read, or until no images are left.
// If no images can be read, this returns an error.
func (d *DirSamples) Random() (image.Image, error) {
	for len(d.imagePaths) > 0 {
		idx := rand.Intn(len(d.imagePaths))
		path := d.imagePaths[idx]
		f, err := os.Open(path)
		if err == nil {
			img, _, err := image.Decode(f)
			f.Close()
			if err == nil {
				return img, nil
			}
		}
		d.imagePaths[idx] = d.imagePaths[len(d.imagePaths)-1]
		d.imagePaths = d.imagePaths[:len(d.imagePaths)-1]
	}
	return nil, errors.New("no usable images")
}

// RandomPair is like Random, but it selects a pair
// of different images.
// If two different images cannot be loaded, this fails.
func (d *DirSamples) RandomPair() (image.Image, image.Image, error) {
	var pair []image.Image
	var lastPath string
	for len(d.imagePaths) > 1 && len(pair) < 2 {
		idx := rand.Intn(len(d.imagePaths))
		path := d.imagePaths[idx]
		if path == lastPath {
			continue
		}
		f, err := os.Open(path)
		if err == nil {
			img, _, err := image.Decode(f)
			f.Close()
			if err == nil {
				lastPath = path
				pair = append(pair, img)
				continue
			}
		}
		d.imagePaths[idx] = d.imagePaths[len(d.imagePaths)-1]
		d.imagePaths = d.imagePaths[:len(d.imagePaths)-1]
	}
	if len(pair) == 2 {
		return pair[0], pair[1], nil
	}
	return nil, nil, errors.New("no usable pair")
}
