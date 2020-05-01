package main

import (
	"image"
	"log"
	"math"
	"os"
	"sync"

	"github.com/faiface/pixel"
)

func getFrames(pic pixel.Picture, w, h, qw, qh float64) (frames []pixel.Rect) {
	for y := pic.Bounds().Min.Y; y < pic.Bounds().Max.Y; y += pic.Bounds().Max.Y / qh {
		for x := pic.Bounds().Min.X; x < pic.Bounds().Max.X; x += pic.Bounds().Max.X / qw {
			frames = append(frames, pixel.R(x, y, x+w, y+h))
		}
	}
	return
}

func VectorMag(vec pixel.Vec) float64 {
	return math.Sqrt((vec.X * vec.X) + (vec.Y * vec.Y))
}

func VectorDiv(vec pixel.Vec, n float64) pixel.Vec {
	return pixel.V(
		vec.X/n,
		vec.Y/n,
	)
}

func VectorNormalize(vec pixel.Vec) pixel.Vec {
	m := VectorMag(vec)
	if m != 0 {
		return VectorDiv(vec, m)
	}
	return pixel.ZV
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func loadPictures(files ...string) map[string]pixel.Picture {
	var wg sync.WaitGroup
	var m sync.Mutex

	filesLength := len(files)
	contents := make(map[string]pixel.Picture, filesLength)
	wg.Add(filesLength)

	for _, file := range files {
		go func(file string) {
			content, err := loadPicture(file)

			if err != nil {
				log.Fatal(err)
			}

			m.Lock()
			contents[file] = content
			m.Unlock()
			wg.Done()
		}(file)
	}

	wg.Wait()

	return contents
}
