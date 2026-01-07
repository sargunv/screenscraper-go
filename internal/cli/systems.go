package cli

import (
	"encoding/json"
	"fmt"

	"sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var systemsCmd = &cobra.Command{
	Use:   "systems",
	Short: "Get list of systems/consoles",
	Long:  "Retrieves the complete list of systems (consoles) available in Screenscraper",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetSystemsList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Systems, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderSystemsList(resp.Response.Systems, lang))
		return nil
	},
}

func init() {
	listCmd.AddCommand(systemsCmd)
}
