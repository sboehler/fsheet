package main

import (
	"fmt"
	"image"
	"image/color"
)

func merge(imgs []image.Image) (*image.Gray, error) {
	var w, h int
	for i, img := range imgs {
		if h == 0 {
			h = img.Bounds().Dy()
		} else {
			if h != img.Bounds().Dy() {
				return nil, fmt.Errorf("invalid images: image %d has height %d != %d", i, img.Bounds().Dy(), h)
			}
		}
		w += img.Bounds().Dx()
	}
	res := image.NewGray(image.Rect(0, 0, w, h))
	var x int
	for _, img := range imgs {
		for xx := 0; xx < img.Bounds().Dx(); xx++ {
			for yy := 0; yy < h; yy++ {
				_, _, _, a := img.At(xx, yy).RGBA()
				res.Set(x+xx, yy, color.Gray{Y: uint8(255 - a)})
			}
		}
		x += img.Bounds().Dx()
	}
	return res, nil
}

type measure struct {
	Start, End int
}
type measures struct {
	measures []measure
}

func newMeasures(img image.Image) (*measures, error) {
	m, err := detectMeasures(img)
	if err != nil {
		return nil, fmt.Errorf("detectMeasures(): %w", err)
	}
	return &measures{
		measures: m,
	}, nil
}

func detectMeasures(img image.Image) ([]measure, error) {
	var (
		res []measure

		first = -1
		last  = -1

		tol = 100
	)
	for x := 0; x < img.Bounds().Dx(); x++ {
		var blacks int
		for y := 0; y < img.Bounds().Dy(); y++ {
			if g, ok := img.At(x, y).(color.Gray); ok {
				if g.Y < 25 {
					blacks++
				}
			} else {
				panic("expected gray scale image")
			}
		}
		if blacks > img.Bounds().Dy()/2 {
			if first < 0 {
				first = x
				last = x
			} else {
				last = x
			}
		} else {
			if first > 0 && first < x-tol {
				res = append(res, measure{Start: first, End: last})
				first = -1
				last = -1
			}
		}
	}
	if first > 0 {
		res = append(res, measure{Start: first, End: last})
	}
	return res, nil
}
