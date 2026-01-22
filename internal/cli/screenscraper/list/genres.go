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

var genresCmd = &cobra.Command{
	Use:   "genres",
	Short: "Get list of genres",
	Long:  "Retrieves the list of all game genres",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListGenresWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no genres data in response")
		}

		genres := resp.JSON200.Response.Genres

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(genres, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderGenresList(genres, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(genresCmd)
}
