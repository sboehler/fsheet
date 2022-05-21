package main

import (
	"fmt"
	"os"

	_ "embed"

	"github.com/signintech/gopdf"
	"github.com/spf13/cobra"
)

var (
	//go:embed Roboto-Medium.ttf
	font []byte
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
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.Flags().IntP("max-line-length-px", "l", 2500, "maximum line length in pixels")
}

func runE(cmd *cobra.Command, args []string) error {
	f := offlinePage{
		Path: args[0],
	}
	md, err := f.parseMetaData()
	if err != nil {
		return err
	}
	res, err := f.findImages()
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
	maxLinelengthPx, err := cmd.Flags().GetInt("max-line-length-px")
	if err != nil {
		return err
	}
	lines, err := m.computeLines(maxLinelengthPx)
	if err != nil {
		return err
	}
	r := renderer{
		pageFormat: gopdf.PageSizeA4,
		font:       font,
		marginTop:  30, marginRight: 30, marginLeft: 30, marginBottom: 30,
		ySpacing: 20,
		title:    md.Title,
		composer: md.Composer,
	}
	return r.render(img, lines, fmt.Sprintf("%s - %s.pdf", md.Title, md.Composer))
}
