package screenscraper

import (
	"context"
	"net/http"
)

// ScreenscraperClient wraps the generated client with stored credentials
type ScreenscraperClient struct {
	*ClientWithResponses
	devID       string
	devPassword string
	softName    string
	ssID        string
	ssPassword  string
}

// NewScreenscraperClient creates a client with stored credentials.
// devID, devPassword, and softName are required developer credentials.
// ssID and ssPassword are optional user credentials.
func NewScreenscraperClient(devID, devPassword, softName, ssID, ssPassword string) (*ScreenscraperClient, error) {
	inner, err := NewClientWithResponses(
		"https://api.screenscraper.fr/api2",
		WithRequestEditorFn(credentialEditor(devID, devPassword, softName, ssID, ssPassword)),
	)
	if err != nil {
		return nil, err
	}
	return &ScreenscraperClient{
		ClientWithResponses: inner,
		devID:               devID,
		devPassword:         devPassword,
		softName:            softName,
		ssID:                ssID,
		ssPassword:          ssPassword,
	}, nil
}

func credentialEditor(devID, devPassword, softName, ssID, ssPassword string) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Set("devid", devID)
		q.Set("devpassword", devPassword)
		q.Set("softname", softName)
		q.Set("output", "json")
		// For botProposition, user creds go in the form body (handled by proposal.go)
		if req.URL.Path != "/api2/botProposition.php" {
			if ssID != "" {
				q.Set("ssid", ssID)
			}
			if ssPassword != "" {
				q.Set("sspassword", ssPassword)
			}
		}
		req.URL.RawQuery = q.Encode()
		return nil
	}
}
