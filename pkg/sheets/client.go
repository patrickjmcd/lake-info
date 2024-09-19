package sheets

import (
	"fmt"
	"github.com/patrickjmcd/gsheets"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

type Client struct {
	gSheets       *gsheets.Client
	spreadsheetId string
	sheetName     string
}

func New(ctx context.Context, spreadsheetId string, credentialsFilePath, b64ServiceAccount *string) (*Client, error) {
	svc, err := gsheets.New(ctx, credentialsFilePath, b64ServiceAccount)
	if err != nil {
		log.Error().Err(err).Msg("unable to retrieve Sheets client")
		return nil, fmt.Errorf("unable to retrieve Sheets client: %w", err)
	}

	client := &Client{
		gSheets:       svc,
		spreadsheetId: spreadsheetId,
	}

	return client, nil
}
