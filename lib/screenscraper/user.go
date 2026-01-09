package screenscraper

import (
	"encoding/json"
	"fmt"
)

// UserInfoResponse is the complete response for the user info endpoint
type UserInfoResponse struct {
	Header   Header `json:"header"`
	Response struct {
		Servers ServerInfo `json:"serveurs"`
		SSUser  UserInfo   `json:"ssuser"`
	} `json:"response"`
}

// GetUserInfo retrieves user information and quotas (ssuserInfos.php).
// Requires user credentials (ssid and sspassword) to be set on the client.
// Returns user level, contribution statistics, thread limits, download speed limits,
// API request quotas (per minute and per day), current usage statistics, and proposal acceptance/rejection rates.
// This information is essential for quota management in scraping software.
func (c *Client) GetUserInfo() (*UserInfoResponse, error) {
	body, err := c.get("ssuserInfos.php", nil)
	if err != nil {
		return nil, err
	}

	var resp UserInfoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	if err := validateResponse(resp.Header); err != nil {
		return nil, err
	}

	return &resp, nil
}
