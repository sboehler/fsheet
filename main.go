package main

import (
	"fmt"
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
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.Flags().String("title", "", "song title")
}

func runE(cmd *cobra.Command, args []string) error {
	f := offlinePage{
		Path: args[0],
	}
	md, err := f.parseMetaData()
	if err != nil {
		return err
	}
	fmt.Println(*md)
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
	lines := m.computeLines(4000)
	r := renderer{
		pageFormat: gopdf.PageSizeA4,
		marginTop:  10, marginRight: 10, marginLeft: 10, marginBottom: 10,
		pixelsPerPoint: 4000 / 550.0,
	}
	return r.render(img, lines, "output.pdf")
}
