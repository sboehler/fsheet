package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/signintech/gopdf"
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

type bar struct {
	StartPos, EndPos int
}

type measure struct {
	Start, End bar
}

func (m measure) length() int {
	return m.End.EndPos - m.Start.StartPos + 1
}

type line []measure

func (l line) length() int {
	if len(l) == 0 {
		return 0
	}
	return l[len(l)-1].End.EndPos - l[0].Start.StartPos + 1
}

func (l line) extraLength(m measure) int {
	if len(l) == 0 {
		return m.length()
	}
	return m.End.EndPos - l[0].Start.StartPos + 1
}

type measures struct {
	measures      []measure
	width, height int
}

func newMeasures(img image.Image) (*measures, error) {
	bars, err := detectBars(img)
	if err != nil {
		return nil, fmt.Errorf("detectMeasures(): %w", err)
	}
	if len(bars) < 2 {
		return nil, fmt.Errorf("only %d bars detected", len(bars))
	}
	var m []measure
	prev, tail := bars[0], bars[1:]
	for _, b := range tail {
		m = append(m, measure{Start: prev, End: b})
		prev = b
	}
	return &measures{
		measures: m,
		width:    img.Bounds().Dx(),
		height:   img.Bounds().Dy(),
	}, nil
}

func (mss measures) computeLines(maxPixels int) []line {
	var (
		res     []line
		current line
	)
	for _, m := range mss.measures {
		if current.extraLength(m) < maxPixels {
			current = append(current, m)
			continue
		}
		res = append(res, current)
		current = []measure{m}
	}
	return append(res, current)
}

func detectBars(img image.Image) ([]bar, error) {
	var (
		res []bar

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
				res = append(res, bar{StartPos: first, EndPos: last})
				first = -1
				last = -1
			}
		}
	}
	if first > 0 {
		res = append(res, bar{StartPos: first, EndPos: last})
	}
	return res, nil
}

type renderer struct {
	marginTop, marginRight, marginBottom, marginLeft float64
	pixelsPerPoint                                   float64
	ySpacing                                         int
	pageFormat                                       *gopdf.Rect
	title, composer                                  string
}

func (rnd renderer) render(img *image.Gray, lls []line, path string) error {
	var (
		pdf        gopdf.GoPdf
		lineHeight = rnd.pxToPt(img.Bounds().Dy())
	)

	if rnd.marginTop+rnd.marginBottom+lineHeight > rnd.pageFormat.H {
		return fmt.Errorf("can't print one line per page")
	}

	pdf.Start(gopdf.Config{
		PageSize: *rnd.pageFormat,
	})
	pdf.AddPage()

	y := rnd.marginTop

	for _, l := range lls {
		if y+lineHeight > rnd.pageFormat.H-rnd.marginBottom {
			pdf.AddPage()
			y = rnd.marginTop
		}
		sub := image.Rect(l[0].Start.StartPos, 0, l[len(l)-1].End.EndPos, img.Bounds().Dy())
		tgt := &gopdf.Rect{W: rnd.pxToPt(sub.Dx()), H: rnd.pxToPt(sub.Dy())}
		if err := pdf.ImageFrom(img.SubImage(sub), rnd.marginLeft, y, tgt); err != nil {
			return err
		}
		y += lineHeight
	}
	return pdf.WritePdf(path)
}

func (rnd renderer) pxToPt(px int) float64 {
	return float64(px) / rnd.pixelsPerPoint

}
