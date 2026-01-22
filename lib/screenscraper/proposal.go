package screenscraper

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
)

// SubmitProposalWithResponse is a convenience wrapper that handles multipart form encoding
// for the SubmitProposal endpoint.
func (c *ScreenscraperClient) SubmitProposalWithResponse(ctx context.Context, body SubmitProposalMultipartBody, reqEditors ...RequestEditorFn) (*SubmitProposalResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Helper to add text fields
	addField := func(name string, value string) error {
		if value != "" {
			return writer.WriteField(name, value)
		}
		return nil
	}

	// Add user credentials to form body (credential editor skips query params for this endpoint)
	if c.ssID != "" {
		if err := writer.WriteField("ssid", c.ssID); err != nil {
			return nil, err
		}
	}
	if c.ssPassword != "" {
		if err := writer.WriteField("sspassword", c.ssPassword); err != nil {
			return nil, err
		}
	}

	if err := addField("gameid", body.GameID); err != nil {
		return nil, err
	}
	if err := addField("romid", body.ROMID); err != nil {
		return nil, err
	}
	if err := addField("modiftypeinfo", body.ModifyInfoType); err != nil {
		return nil, err
	}
	if err := addField("modiftexte", body.ModifyText); err != nil {
		return nil, err
	}
	if err := addField("modiftypemedia", body.ModifyMediaType); err != nil {
		return nil, err
	}
	if err := addField("modifmediafileurl", body.ModifyMediaFileURL); err != nil {
		return nil, err
	}
	if err := addField("modifregion", body.ModifyRegion); err != nil {
		return nil, err
	}
	if err := addField("modiflangue", body.ModifyLanguage); err != nil {
		return nil, err
	}
	if err := addField("modifversion", body.ModifyVersion); err != nil {
		return nil, err
	}
	if err := addField("modiftyperegion", body.ModifyTypeRegion); err != nil {
		return nil, err
	}
	if err := addField("modiftypenumsupport", body.ModifySupportNumber); err != nil {
		return nil, err
	}
	if err := addField("modifsource", body.ModifySource); err != nil {
		return nil, err
	}

	// Add file if provided
	if body.ModifyMediaFile.Filename() != "" {
		part, err := writer.CreateFormFile("modifmediafile", body.ModifyMediaFile.Filename())
		if err != nil {
			return nil, err
		}
		reader, err := body.ModifyMediaFile.Reader()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		if _, err := io.Copy(part, reader); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return c.ClientWithResponses.SubmitProposalWithBodyWithResponse(
		ctx,
		writer.FormDataContentType(),
		&buf,
		reqEditors...,
	)
}
