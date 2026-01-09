package screenscraper

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

// SubmitInfoProposalParams contains parameters for submitting a text info proposal
type SubmitInfoProposalParams struct {
	// Target: exactly one must be specified
	GameID string // Numeric identifier of the game on ScreenScraper
	ROMID  string // Numeric identifier of the ROM on ScreenScraper

	// Required fields
	InfoType string // Type of info sent (name, editeur, developpeur, players, score, rating, genres, etc.)
	Text     string // The information itself

	// Optional fields
	Region   string // Short name of the info region (see regionsListe.php)
	Language string // Short name of the info language (see languesListe.php)
	Version  string // Version of the info
	Source   string // Source (URL of web page, scan of original support, author, etc.)
}

// SubmitMediaProposalParams contains parameters for submitting a media proposal
type SubmitMediaProposalParams struct {
	// Target: exactly one must be specified
	GameID string // Numeric identifier of the game on ScreenScraper
	ROMID  string // Numeric identifier of the ROM on ScreenScraper

	// Required fields
	MediaType string // Type of media sent (sstitle, ss, fanart, video, wheel, box-2D, etc.)

	// Media source: exactly one must be specified
	MediaFile     io.Reader // File content to upload
	MediaFileName string    // Name of the file being uploaded (required if MediaFile is set)
	MediaFileURL  string    // URL of media to download

	// Optional fields
	Region        string // Short name of the region (see regionsListe.php)
	SupportNumber string // Support number (0 to 10)
	Version       string // Version of the info
	Source        string // Source (URL of web page, scan of original support, author, etc.)
}

// SubmitProposalResponse contains the response from submitting a proposal
type SubmitProposalResponse struct {
	// Message is the textual response from the API
	Message string
}

// SubmitInfoProposal submits a text info proposal (botProposition.php).
// The request must be sent as an HTML form of type "multipart/form-data" with the "POST" method.
// Requires user credentials (SSID and SSPassword) to be set on the client.
// Info types for games: name, editeur, developpeur, players, score, rating, genres, datessortie, rotation, resolution, modes, familles, numero, styles, themes, description.
// Info types for ROMs: developpeur, editeur, datessortie, players, regions, langues, clonetype, hacktype, friendly, serial, description.
// See the API documentation for complete lists and format requirements.
func (c *Client) SubmitInfoProposal(params SubmitInfoProposalParams) (*SubmitProposalResponse, error) {
	// Validate that user credentials are provided
	if c.SSID == "" || c.SSPassword == "" {
		return nil, fmt.Errorf("user credentials (SSID and SSPassword) are required to submit proposals")
	}

	// Validate that exactly one of GameID or ROMID is provided
	if (params.GameID == "" && params.ROMID == "") || (params.GameID != "" && params.ROMID != "") {
		return nil, fmt.Errorf("exactly one of GameID or ROMID must be specified")
	}

	// Validate required fields
	if params.InfoType == "" {
		return nil, fmt.Errorf("InfoType is required")
	}
	if params.Text == "" {
		return nil, fmt.Errorf("Text is required")
	}

	// Build form fields
	fields := map[string]string{
		"modiftypeinfo": params.InfoType,
		"modiftexte":    params.Text,
	}

	if params.GameID != "" {
		fields["gameid"] = params.GameID
	}
	if params.ROMID != "" {
		fields["romid"] = params.ROMID
	}
	if params.Region != "" {
		fields["modifregion"] = params.Region
	}
	if params.Language != "" {
		fields["modiflangue"] = params.Language
	}
	if params.Version != "" {
		fields["modifversion"] = params.Version
	}
	if params.Source != "" {
		fields["modifsource"] = params.Source
	}

	return c.postProposal(fields, nil, "")
}

// SubmitMediaProposal submits a media proposal (botProposition.php).
// The request must be sent as an HTML form of type "multipart/form-data" with the "POST" method.
// Requires user credentials (SSID and SSPassword) to be set on the client.
// Media types include: sstitle, ss, fanart, video, overlay, steamgrid, wheel, wheel-hd, marquee, screenmarquee, box-2D, box-2D-side, box-2D-back, box-texture, manuel, flyer, maps, figurine, support-texture, box-scan, support-scan, bezel-4-3, bezel-16-9, etc.
// See the API documentation for complete list and format requirements.
func (c *Client) SubmitMediaProposal(params SubmitMediaProposalParams) (*SubmitProposalResponse, error) {
	// Validate that user credentials are provided
	if c.SSID == "" || c.SSPassword == "" {
		return nil, fmt.Errorf("user credentials (SSID and SSPassword) are required to submit proposals")
	}

	// Validate that exactly one of GameID or ROMID is provided
	if (params.GameID == "" && params.ROMID == "") || (params.GameID != "" && params.ROMID != "") {
		return nil, fmt.Errorf("exactly one of GameID or ROMID must be specified")
	}

	// Validate required fields
	if params.MediaType == "" {
		return nil, fmt.Errorf("MediaType is required")
	}

	// Validate that exactly one media source is provided
	hasFile := params.MediaFile != nil
	hasURL := params.MediaFileURL != ""
	if (!hasFile && !hasURL) || (hasFile && hasURL) {
		return nil, fmt.Errorf("exactly one of MediaFile or MediaFileURL must be specified")
	}

	// If MediaFile is provided, MediaFileName must also be provided
	if hasFile && params.MediaFileName == "" {
		return nil, fmt.Errorf("MediaFileName is required when MediaFile is provided")
	}

	// Build form fields
	fields := map[string]string{
		"modiftypemedia": params.MediaType,
	}

	if params.GameID != "" {
		fields["gameid"] = params.GameID
	}
	if params.ROMID != "" {
		fields["romid"] = params.ROMID
	}
	if params.MediaFileURL != "" {
		fields["modifmediafileurl"] = params.MediaFileURL
	}
	if params.Region != "" {
		fields["modiftyperegion"] = params.Region
	}
	if params.SupportNumber != "" {
		fields["modiftypenumsupport"] = params.SupportNumber
	}
	if params.Version != "" {
		fields["modiftypeversion"] = params.Version
	}
	if params.Source != "" {
		fields["modifmediasource"] = params.Source
	}

	return c.postProposal(fields, params.MediaFile, params.MediaFileName)
}

func (c *Client) postProposal(fields map[string]string, fileReader io.Reader, fileName string) (*SubmitProposalResponse, error) {
	// Create a buffer to write the multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add credential fields
	if err := writer.WriteField("ssid", c.SSID); err != nil {
		return nil, fmt.Errorf("failed to write ssid field: %w", err)
	}
	if err := writer.WriteField("sspassword", c.SSPassword); err != nil {
		return nil, fmt.Errorf("failed to write sspassword field: %w", err)
	}

	// Add all other fields
	for key, value := range fields {
		if value != "" {
			if err := writer.WriteField(key, value); err != nil {
				return nil, fmt.Errorf("failed to write field %s: %w", key, err)
			}
		}
	}

	// Add file if provided
	if fileReader != nil && fileName != "" {
		part, err := writer.CreateFormFile("modifmediafile", fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}
		if _, err := io.Copy(part, fileReader); err != nil {
			return nil, fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	// Close the writer to finalize the multipart message
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Build the URL with dev credentials
	apiURL := c.buildURL("botProposition.php", nil)

	// Create the POST request
	req, err := http.NewRequest("POST", apiURL, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	message := strings.TrimSpace(string(responseBody))

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, newAPIError(resp.StatusCode, message)
	}

	return &SubmitProposalResponse{
		Message: message,
	}, nil
}
