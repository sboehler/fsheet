package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/signintech/gopdf"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:  "fsheet",
	RunE: runE,
}

func init() {
	rootCmd.Flags().String("title", "", "song title")
	rootCmd.Flags().String("url", "", "image url")
	rootCmd.MarkFlagRequired("url")
}

func runE(cmd *cobra.Command, args []string) error {
	imgURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	u, err := url.Parse(imgURL)
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
