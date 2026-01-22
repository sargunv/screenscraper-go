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

var gameInfoTypesCmd = &cobra.Command{
	Use:   "game-info-types",
	Short: "Get list of game info types",
	Long:  "Retrieves the list of all available info types for games",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListGameInfoTypesWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no game info types data in response")
		}

		infoTypes := resp.JSON200.Response.Info

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(infoTypes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderGameInfoTypesList(infoTypes, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(gameInfoTypesCmd)
}
