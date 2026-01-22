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

var classificationsCmd = &cobra.Command{
	Use:   "classifications",
	Short: "Get list of classifications",
	Long:  "Retrieves the list of all game classifications/ratings (ESRB, PEGI, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListClassificationsWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no classifications data in response")
		}

		classifications := resp.JSON200.Response.Classifications

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(classifications, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderClassificationsList(classifications, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(classificationsCmd)
}
