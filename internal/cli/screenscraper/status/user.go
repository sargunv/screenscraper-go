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

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Get user information and quotas",
	Long:  "Retrieves user information including quotas, rate limits, and contribution stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := shared.Client.GetUserInfoWithResponse(context.Background())
		if err != nil {
			return err
		}

		if !screenscraper.IsSuccess(resp) {
			return fmt.Errorf("API error: HTTP %d", resp.StatusCode())
		}

		if resp.JSON200 == nil || resp.JSON200.Response.User.Id == "" {
			return fmt.Errorf("no user data in response")
		}

		user := resp.JSON200.Response.User

		if shared.JsonOutput {
			formatted, err := json.MarshalIndent(user, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(shared.Locale)
		fmt.Print(format.RenderUser(user, lang))
		return nil
	},
}

func init() {
	Cmd.AddCommand(userCmd)
}
