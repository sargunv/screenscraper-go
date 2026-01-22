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

var mediaTypesCmd = &cobra.Command{
	Use:   "media-types",
	Short: "Get list of media types",
	Long:  "Retrieves the list of available media types for games and systems",
}

var gameMediaTypesCmd = &cobra.Command{
	Use:   "games",
	Short: "Get list of game media types",
	Long:  "Retrieves the list of available media types for games",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListGameMediaTypesWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no media types data in response")
		}

		medias := resp.JSON200.Response.Media

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(medias, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderMediaTypesList(medias, lang))
		return nil
	},
}

var systemMediaTypesCmd = &cobra.Command{
	Use:   "systems",
	Short: "Get list of system media types",
	Long:  "Retrieves the list of available media types for systems",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.ListSystemMediaTypesWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no media types data in response")
		}

		medias := resp.JSON200.Response.Media

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(medias, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderMediaTypesList(medias, lang))
		return nil
	},
}

func init() {
	mediaTypesCmd.AddCommand(gameMediaTypesCmd)
	mediaTypesCmd.AddCommand(systemMediaTypesCmd)
	Cmd.AddCommand(mediaTypesCmd)
}
