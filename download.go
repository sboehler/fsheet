package main

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/sync/errgroup"
)

// Downloader downloads individual sheet images.
type Downloader struct {
	client   http.Client
	basepath string
}

// Download downloads the i-th image.
func (dl *Downloader) Download(ctx context.Context, index int) (image.Image, error) {
	u := fmt.Sprintf("%s%d.png", dl.basepath, index)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := dl.client.Do(req)
	if resp.StatusCode == 403 {
		// handle gracefully
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return png.Decode(resp.Body)
}

func (dl *Downloader) DownloadAll(ctx context.Context) ([]image.Image, error) {
	var cache = make(map[int]image.Image)

	max, err := findMax(func(n int) (bool, error) {
		img, err := dl.Download(ctx, n)
		if err != nil {
			return false, err
		}
		if img != nil {
			cache[n] = img
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	grp, ctx := errgroup.WithContext(ctx)
	res := make([]image.Image, max)
	for i := 0; i < max; i++ {
		if img, ok := cache[i]; ok {
			res[i] = img
			continue
		}
		i := i
		grp.Go(func() error {
			img, err := dl.Download(ctx, i)
			res[i] = img
			return err
		})
	}
	return res, grp.Wait()
}

// NewDownloader creates a new Downloader from the given
// image URL.
func NewDownloader(u *url.URL) (*Downloader, error) {
	s := u.String()
	img := path.Base(s)
	return &Downloader{
		basepath: strings.TrimSuffix(s, img),
	}, nil
}

// findMax returns the smallest index for which the given function
// returns false. If f returns an error, the computation is stopped.
func findMax(f func(int) (bool, error)) (int, error) {
	n := 4
	for {
		ok, err := f(n)
		if err != nil {
			return n, err
		}
		if ok {
			n *= 2
		} else {
			break
		}
	}
	var (
		l = 0
		r = n
		i = n / 2
	)
	for {
		ok, err := f(i)
		if err != nil {
			return i, err
		}
		if ok {
			l = i + 1
		} else {
			r = i
		}
		fmt.Println(l, r)
		i = (r + l) / 2
		if r <= l {
			break
		}
	}
	return i, nil
}
