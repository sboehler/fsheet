package main

import (
	"fmt"
	"image"
	"image/draw"
)

func merge(imgs []image.Image) (image.Image, error) {
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
	res := image.NewNRGBA(image.Rect(0, 0, w, h))
	var x int
	for _, img := range imgs {
		draw.Draw(res, image.Rect(x, 0, x+img.Bounds().Dx(), h), img, image.Point{0, 0}, draw.Src)
		x += img.Bounds().Dx()
	}
	return res, nil
}

func detectMeasures(img image.Image) ([]int, []int, error) {
	var (
		ins, outs []int

		first = -1
		last  = -1

		tol = 100
	)
	for x := 0; x < img.Bounds().Dx(); x++ {
		var blacks int
		for y := 0; y < img.Bounds().Dy(); y++ {
			_, _, _, a := img.At(x, y).RGBA()
			if float64(a) > float64(0xffff*0.9) {
				blacks++
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
				ins = append(ins, first)
				outs = append(outs, last)
				first = -1
				last = -1
			}
		}
	}
	if first > 0 {
		ins = append(ins, first)
		outs = append(outs, last)
	}
	return ins, outs, nil
}
