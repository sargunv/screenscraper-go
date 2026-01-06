package cli

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List metadata and reference data",
	Long:  "Commands to retrieve lists of metadata and reference data from Screenscraper",
}

func init() {
	rootCmd.AddCommand(listCmd)
}
