package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/signintech/gopdf"
	"github.com/spf13/cobra"
)

func CreateSheetCommand() *cobra.Command {

	var sheet Sheet

	cmd := &cobra.Command{
		Use: "sheet",
		RunE: func(cmd *cobra.Command, args []string) error {
			sheet.InputDir = args[0]
			return sheet.Create()
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringVar(&sheet.Title, "title", "unnamed", "Song title")
	cmd.Flags().IntVar(&sheet.LineLength, "max-line-length-px", 2500, "maximum line length in pixels")
	cmd.MarkFlagRequired("title")
	return cmd
}

type Sheet struct {
	Title      string
	InputDir   string
	LineLength int
}

func (sh *Sheet) Create() error {
	images, err := sh.findImages()
	if err != nil {
		return err
	}
	img, err := sh.merge(images)
	if err != nil {
		return err
	}
	m, err := newMeasures(img)
	if err != nil {
		return err
	}
	lines, err := m.computeLines(sh.LineLength)
	if err != nil {
		return err
	}
	r := renderer{
		pageFormat: gopdf.PageSizeA4,
		font:       font,
		marginTop:  30, marginRight: 30, marginLeft: 30, marginBottom: 30,
		ySpacing: 10,
		title:    sh.Title,
	}
	return r.render(img, lines, fmt.Sprintf("%s.pdf", sh.Title))
}

// findImages finds all sheet images.
func (sheet *Sheet) findImages() ([]image.Image, error) {
	files := make(map[int]image.Image)
	err := filepath.Walk(sheet.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		segs := strings.SplitN(info.Name(), ".", 2)
		if len(segs) != 2 {
			return nil
		}
		if strings.ToLower(segs[1]) != "png" {
			return nil
		}
		n, err := strconv.ParseInt(segs[0], 10, 64)
		if err != nil {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		img, err := png.Decode(f)
		if err != nil {
			return err
		}
		files[int(n)] = img
		return nil
	})
	if err != nil {
		return nil, err
	}
	res := make([]image.Image, len(files))
	var missing []int
	for i := 0; i < len(files); i++ {
		img, ok := files[i]
		if !ok {
			missing = append(missing, i)
			continue
		}
		res[i] = img
	}
	if len(missing) > 0 {
		return res, fmt.Errorf("missing images: %v", missing)
	}
	return res, nil
}

func (sh *Sheet) merge(imgs []image.Image) (*image.Gray, error) {
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

func (mss measures) computeLines(maxPixels int) ([]line, error) {
	var (
		res     []line
		current line
	)
	for _, m := range mss.measures {
		if current.extraLength(m) < maxPixels {
			current = append(current, m)
			continue
		}
		if current.length() == 0 {
			return nil, fmt.Errorf("maximum line length too small: %d", maxPixels)
		}
		res = append(res, current)
		current = []measure{m}
	}
	return append(res, current), nil
}

func detectBars(img image.Image) ([]bar, error) {
	var (
		res []bar

		first = -1
		last  = -1

		tol = 10
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
		ratio := float64(blacks) / float64(img.Bounds().Dy())
		if ratio > 0.5 {
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
	font                                             []byte
	marginTop, marginRight, marginBottom, marginLeft float64
	ySpacing                                         float64
	pageFormat                                       *gopdf.Rect
	title, composer                                  string

	pixelsPerPoint float64
}

func (rnd *renderer) render(img *image.Gray, lls []line, path string) error {
	var maxLineLengthPx int
	for _, l := range lls {
		if l.length() > maxLineLengthPx {
			maxLineLengthPx = l.length()
		}
	}
	rnd.pixelsPerPoint = float64(maxLineLengthPx) / (rnd.pageFormat.W - rnd.marginLeft - rnd.marginRight)
	var (
		pdf        = new(gopdf.GoPdf)
		lineHeight = rnd.pxToPt(img.Bounds().Dy())
	)

	if rnd.marginTop+rnd.marginBottom+lineHeight > rnd.pageFormat.H {
		return fmt.Errorf("can't print one line per page")
	}

	pdf.Start(gopdf.Config{
		PageSize: *rnd.pageFormat,
	})
	if err := pdf.AddTTFFontByReader("roboto", bytes.NewBuffer(rnd.font)); err != nil {
		return err
	}
	pdf.AddPage()

	if err := pdf.SetFont("roboto", "", 20); err != nil {
		return err
	}
	y := rnd.marginTop
	pdf.SetY(y)

	if err := rnd.centeredText(pdf, rnd.title, 20); err != nil {
		return err
	}
	y += 30.0
	pdf.SetY(y)
	if err := rnd.centeredText(pdf, rnd.composer, 13.33); err != nil {
		return err
	}
	y += 30.0
	pdf.SetY(y)

	for _, l := range lls {
		if y+lineHeight > rnd.pageFormat.H-rnd.marginBottom {
			pdf.AddPage()
			y = rnd.marginTop
		}
		sub := image.Rect(l[0].Start.StartPos, 0, l[len(l)-1].End.EndPos, img.Bounds().Dy())
		tgt := &gopdf.Rect{
			W: rnd.pxToPt(sub.Dx()),
			H: rnd.pxToPt(sub.Dy()),
		}
		if err := pdf.ImageFrom(img.SubImage(sub), rnd.marginLeft, y, tgt); err != nil {
			return err
		}
		y += lineHeight + rnd.ySpacing
	}
	return pdf.WritePdf(path)
}

func (rnd renderer) pxToPt(px int) float64 {
	return float64(px) / rnd.pixelsPerPoint

}

func (rnd renderer) centeredText(pdf *gopdf.GoPdf, text string, size float64) error {
	pdf.SetX(rnd.marginLeft)
	if err := pdf.SetFont("roboto", "", size); err != nil {
		return err
	}
	box := &gopdf.Rect{
		W: rnd.pageFormat.W - rnd.marginLeft - rnd.marginRight,
		H: size,
	}
	return pdf.CellWithOption(box, text, gopdf.CellOption{
		Align: gopdf.Center,
	})
}
