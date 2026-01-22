package detail

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"
	"github.com/sargunv/rom-tools/lib/screenscraper"

	"github.com/spf13/cobra"
)

var (
	systemDetailID int
)

var systemCmd = &cobra.Command{
	Use:     "system",
	Short:   "Get detailed system information",
	Long:    "Retrieves detailed information about a specific system/console",
	Example: `  rom-tools screenscraper detail system --id=1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListSystemsWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no systems data in response")
		}

		systems := resp.JSON200.Response.Systems

		// Find system by ID
		var found *screenscraper.System
		for i := range systems {
			if systems[i].Id == systemDetailID {
				found = &systems[i]
				break
			}
		}

		if found == nil {
			return fmt.Errorf("system with ID %d not found", systemDetailID)
		}

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(found, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderSystemDetail(*found, lang))
		return nil
	},
}

func init() {
	systemCmd.Flags().IntVar(&systemDetailID, "id", 0, "System ID")
	systemCmd.MarkFlagRequired("id")
	Cmd.AddCommand(systemCmd)
}
