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

var familiesCmd = &cobra.Command{
	Use:   "families",
	Short: "Get list of families",
	Long:  "Retrieves the list of all game families (e.g., Sonic, Mario, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListFamiliesWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no families data in response")
		}

		families := resp.JSON200.Response.Families

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(families, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderFamiliesList(families, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(familiesCmd)
}
