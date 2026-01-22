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

var playerCountsCmd = &cobra.Command{
	Use:   "player-counts",
	Short: "Get list of player counts",
	Long:  "Retrieves the list of all player counts (1 player, 2 players, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListPlayerCountsWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no player counts data in response")
		}

		playerCounts := resp.JSON200.Response.PlayerCounts

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(playerCounts, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderPlayerCountsList(playerCounts, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(playerCountsCmd)
}
