package cmd

import (
	"github.com/patrickjmcd/lake-info/pkg/sheets"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(printCmd)
}

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print all measurements for a given lake",
	Long:  `Print all measurements for a given lake.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		googleSAB64 := viper.GetString(google_sa_b64)
		spreadsheetId := viper.GetString(spreadsheet_id)
		gSheetsClient, err := sheets.New(ctx, spreadsheetId, nil, &googleSAB64)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create Google Sheets client")
		}
		measurements, err := gSheetsClient.ReadMeasurements(ctx, sheets.TableRockLake)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read measurements")
		}

		for _, m := range measurements {
			log.Info().Msgf("Measurement: %v", m)
		}
	},
}
