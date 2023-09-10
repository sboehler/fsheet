package main

import (
	"fmt"
	"os"

	_ "embed"

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
	Use: "fsheet",
}

func init() {
	rootCmd.AddCommand(CreateDownloadCommand())
	rootCmd.AddCommand(CreateSheetCommand())
}
