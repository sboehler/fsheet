package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/signintech/gopdf"
)

var (
	title  = flag.String("title", "", "song title")
	imgURL = flag.String("url", "", "image url")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	u, err := url.Parse(*imgURL)
	if err != nil {
		return err
	}
	dl, err := NewDownloader(u)
	if err != nil {
		return err
	}
	ctx := context.Background()
	res, err := dl.DownloadAll(ctx)
	if err != nil {
		return err
	}
	img, err := merge(res)
	if err != nil {
		return err
	}
	m, err := newMeasures(img)
	if err != nil {
		return err
	}
	lines := m.computeLines(4000)
	r := renderer{
		pageFormat: gopdf.PageSizeA4,
		marginTop:  10, marginRight: 10, marginLeft: 10, marginBottom: 10,
		pixelsPerPoint: 4000 / 550.0,
	}
	return r.render(img, lines, "output.pdf")
}
