package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"
	"github.com/sargunv/rom-tools/internal/format"
	"github.com/sargunv/rom-tools/lib/screenscraper"

	"github.com/spf13/cobra"
)

var infraCmd = &cobra.Command{
	Use:   "infra",
	Short: "Get infrastructure/server information",
	Long:  "Retrieves Screenscraper infrastructure info including CPU usage, API load, and status",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetInfraInfoWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil {
			return fmt.Errorf("no infrastructure data in response")
		}

		servers := resp.JSON200.Response.Servers

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(servers, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderInfra(servers, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(infraCmd)
}
