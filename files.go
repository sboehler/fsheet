package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// offlinePage downloads individual sheet images.
type offlinePage struct {
	Path string
}

// findImages finds all sheet images.
func (dl *offlinePage) findImages() ([]image.Image, error) {
	path := filepath.Dir(dl.Path)
	files := make(map[int]image.Image)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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

type metadata struct {
	Title    string
	Composer string
}

// parseMetaData tries to extract composer and title.
func (dl offlinePage) parseMetaData() (*metadata, error) {
	f, err := os.Open(dl.Path)
	if err != nil {
		return nil, err
	}
	z := html.NewTokenizer(f)

	// strategy: look for a <span> followed by a <h4>, these
	// will contain composer and title.
	for {
		// expect <span>
		tt := z.Next()
		if tt == html.ErrorToken {
			return nil, z.Err()
		}
		if tt != html.StartTagToken {
			continue
		}
		tn, _ := z.TagName()
		if string(tn) != "span" {
			continue
		}
		// expect text with composer
		tt = z.Next()
		if tt != html.TextToken {
			continue
		}
		composer := string(z.Text())
		// expect </span>
		tt = z.Next()
		if tt != html.EndTagToken {
			continue
		}
		// skip any white space
		tt = z.Next()
		for tt == html.TextToken {
			tt = z.Next()
		}
		// expect <h4>
		if tt != html.StartTagToken {
			continue
		}
		if tn, _ = z.TagName(); string(tn) != "h4" {
			continue
		}
		// expect text with title
		tt = z.Next()
		if tt != html.TextToken {
			continue
		}
		title := string(z.Text())
		return &metadata{
			Title:    title,
			Composer: composer,
		}, nil
	}
}
