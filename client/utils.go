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

func Map(v, s1, st1, s2, st2 float64) float64 {
	newval := (v-s1)/(st1-s1)*(st2-s2) + s2
	if s2 < st2 {
		if newval < s2 {
			return s2
		}
		if newval > st2 {
			return st2
		}
	} else {
		if newval > s2 {
			return s2
		}
		if newval < st2 {
			return st2
		}
	}
	return newval
}

func Dist(v1, v2 pixel.Vec) float64 {
	return math.Sqrt(math.Pow(math.Abs(v1.X-v2.X), 2) + math.Pow(math.Abs(v1.Y-v2.Y), 2))
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
