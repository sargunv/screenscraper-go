package cache

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sargunv/rom-tools/internal/cache"
)

var Cmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the screenscraper cache",
}

var dirCmd = &cobra.Command{
	Use:   "dir",
	Short: "Print the cache directory path",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := cache.DefaultCacheDir()
		if err != nil {
			return err
		}
		fmt.Println(dir)
		return nil
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clear all cached data",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := cache.DefaultCacheDir()
		if err != nil {
			return err
		}

		// Create a cache instance just to clear it
		c, err := cache.New(dir, 0, cache.ModeNormal)
		if err != nil {
			return err
		}

		if err := c.Clear(); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}

		fmt.Println("Cache cleared.")
		return nil
	},
}

func init() {
	Cmd.AddCommand(dirCmd)
	Cmd.AddCommand(cleanCmd)
}
