package main

import (
	"context"
	"flag"
	"fmt"
	"image/png"
	"net/url"
	"os"
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
	fmt.Println(*imgURL)
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
	ins, outs, err := detectMeasures(img)
	if err != nil {
		return err
	}
	fmt.Println(len(ins), ins)
	fmt.Println(len(outs), outs)
	out, err := os.Create("./output.png")
	if err != nil {
		return err
	}
	if err := png.Encode(out, img); err != nil {
		return err
	}
	fmt.Println("fsheet!")
	return nil
}
