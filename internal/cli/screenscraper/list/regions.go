package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"
	"github.com/sargunv/rom-tools/lib/screenscraper"

	"github.com/spf13/cobra"
)

var regionsCmd = &cobra.Command{
	Use:   "regions",
	Short: "Get list of regions",
	Long:  "Retrieves the list of all regions (USA, Europe, Japan, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListRegionsWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no regions data in response")
		}

		regions := resp.JSON200.Response.Regions

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(regions, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderRegionsList(regions, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(regionsCmd)
}
