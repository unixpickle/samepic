// Command samer_dir finds similar images in a directory
// of images.
//
// The tool prints pairs of filenames.
// Every filename is separated by a newline.
package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/samepic"

	_ "image/jpeg"
	_ "image/png"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	samerFlags := &samepic.Flags{}
	samerFlags.AddToSet(fs)

	var dir string
	fs.StringVar(&dir, "dir", "", "directory of images")

	fs.Parse(os.Args[1:])

	if dir == "" {
		essentials.Die("Required flag: -dir. See -help.")
	}

	samer, err := samerFlags.BatchSamer()
	if err != nil {
		essentials.Die(err)
	}

	imgChan := streamImages(dir)
	for dup := range samer.SameBatch(imgChan) {
		fmt.Println(dup[0])
		fmt.Println(dup[1])
	}
}

func streamImages(dir string) <-chan *samepic.IDImage {
	listing, err := ioutil.ReadDir(dir)
	if err != nil {
		essentials.Die(err)
	}

	imageChan := make(chan *samepic.IDImage, 1)
	go func() {
		defer close(imageChan)
		for _, item := range listing {
			ext := strings.ToLower(filepath.Ext(item.Name()))
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
				continue
			}
			path := filepath.Join(dir, item.Name())
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			img, _, err := image.Decode(f)
			f.Close()
			if err != nil {
				fmt.Fprintln(os.Stderr, "decode "+path+":", err)
				continue
			}
			imageChan <- &samepic.IDImage{Image: img, ID: path}
		}
	}()
	return imageChan
}
