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

var supportTypesCmd = &cobra.Command{
	Use:   "support-types",
	Short: "Get list of support types",
	Long:  "Retrieves the list of all support types (cartridge, CD, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListSupportTypesWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no support types data in response")
		}

		supportTypes := resp.JSON200.Response.SupportTypes

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(supportTypes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderSupportTypesList(supportTypes, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(supportTypesCmd)
}
