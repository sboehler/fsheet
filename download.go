package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/spf13/cobra"
)

func CreateDownloadCommand() *cobra.Command {

	var downloader Downloader

	cmd := &cobra.Command{
		Use: "download",
		RunE: func(cmd *cobra.Command, args []string) error {
			downloader.RootPath = args[0]
			return downloader.DownloadAll()
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringVar(&downloader.OutputDir, "output-dir", "", "directory")
	return cmd
}

type Downloader struct {
	RootPath  string
	OutputDir string
}

func (dl *Downloader) DownloadAll() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	var counter int
	for {
		u, err := dl.createURL(counter)
		if err != nil {
			return fmt.Errorf("could not create URL for number %d: %v", counter, err)
		}
		resp, err := http.Get(u.String())
		if err != nil {
			return fmt.Errorf("error fetching data from URL %s: %w", u.String(), err)
		}
		if resp.StatusCode != 200 {
			return nil
		}
		defer resp.Body.Close()
		outPath := path.Join(wd, dl.OutputDir, fmt.Sprintf("%04d.png", counter))
		if err := dl.write(outPath, resp.Body); err != nil {
			return err
		}
		fmt.Printf("Downloaded %s -> %s\n", u, outPath)
		counter++
	}
}

func (dl *Downloader) write(filepath string, reader io.ReadCloser) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

func (dl *Downloader) createURL(n int) (*url.URL, error) {
	p, err := url.JoinPath(dl.RootPath, fmt.Sprintf("%d.png", n))
	if err != nil {
		return nil, err
	}
	return url.Parse(p)
}
