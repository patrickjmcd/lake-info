package cmd

import (
	"github.com/patrickjmcd/gsheets"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"github.com/patrickjmcd/lake-info/lib/measurement"
	"github.com/patrickjmcd/lake-info/lib/tablerock"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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

		spreadsheetId := viper.GetString(spreadsheet_id)
		googleSAB64 := viper.GetString(google_sa_b64)
		sheetsCli, err := gsheets.New[*lakeinfov1.LakeInfoMeasurement](
			ctx,
			spreadsheetId,
			gsheets.WithB64ServiceAccount[*lakeinfov1.LakeInfoMeasurement](googleSAB64),
			gsheets.WithParseRowFn[*lakeinfov1.LakeInfoMeasurement](measurement.MapRowToMeasurement),
			gsheets.WithFormatRowFn[*lakeinfov1.LakeInfoMeasurement](measurement.MakeMeasurementRow),
		)
		if err != nil {
			log.Error().Err(err).Msg("error creating sheets client")
			os.Exit(1)
		}
		log.Info().Msgf("reading from %s/%s", spreadsheetId, tablerock.SheetName)
		measurements, err := sheetsCli.ReadFromSheet(ctx, tablerock.SheetName, "A2:H")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read measurements")
		}

		for _, m := range measurements {
			log.Info().Msgf("Measurement: %v", m)
		}
	},
}
