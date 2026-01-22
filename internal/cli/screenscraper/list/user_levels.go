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

var userLevelsCmd = &cobra.Command{
	Use:   "user-levels",
	Short: "Get list of user levels",
	Long:  "Retrieves the list of all ScreenScraper user levels",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListUserLevelsWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no user levels data in response")
		}

		userLevels := resp.JSON200.Response.UserLevels

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(userLevels, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderUserLevelsList(userLevels, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(userLevelsCmd)
}
