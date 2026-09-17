package cmd

import (
	"context"
	"github.com/patrickjmcd/gsheets"
	"github.com/patrickjmcd/lake-info/dal"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"github.com/patrickjmcd/lake-info/lib/cda"
	"github.com/patrickjmcd/lake-info/lib/measurement"
	"github.com/patrickjmcd/lake-info/lib/tablerock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"time"
)

func init() {
	rootCmd.AddCommand(scrapeCmd)

	scrapeCmd.Flags().BoolP("all", "A", false, "Get all records")
	viper.BindPFlag(all_records, scrapeCmd.Flags().Lookup("all"))

	scrapeCmd.Flags().BoolP("dry-run", "D", false, "Dry run")
	viper.BindPFlag(dry_run, scrapeCmd.Flags().Lookup("dry-run"))
}

var scrapeCmd = &cobra.Command{
	Use:   "scrape [lake]",
	Short: "Scrape lake info",
	Long:  `Scrape lake info from USACE Lake Info.`,
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {

		allRecords := viper.GetBool(all_records)
		dryRun := viper.GetBool(dry_run)

		slog.Info("scraping lake info", "lake", args[0], "all", allRecords, "dry-run", dryRun)

		if args[0] == "tablerock" {
			ctx := context.Background()
			var recordsToStore []*lakeinfov1.LakeInfoMeasurement

			if allRecords {
				if r, err := tablerock.GetAllRecords(tablerock.LakeURL); err != nil {
					slog.Error("error getting all records", "error", err)
				} else {
					recordsToStore = r
				}
			} else {
				record, err := latestTableRockRecord(ctx)
				if err != nil {
					slog.Error("error getting latest record", "error", err)
					os.Exit(1)
				}

				// Temperature is not part of the USACE tabular data, so fetch
				// it from the configured source and attach it to the current
				// reading. A failure here is non-fatal: level data is still
				// stored, temperature is simply left unset.
				tempToken := viper.GetString(whiteriversky_api_token)
				if tempToken != "" {
					tempURL := viper.GetString(whiteriversky_api_url)
					if temp, tErr := tablerock.GetTemperature(tempURL, tempToken); tErr != nil {
						slog.Error("error getting temperature", "error", tErr)
					} else {
						record.Temperature = temp
					}
				} else {
					slog.Warn("temperature source not configured, skipping temperature",
						"envVar", "WHITERIVERSKY_API_TOKEN")
				}

				recordsToStore = append(recordsToStore, record)
			}

			if len(recordsToStore) == 0 {
				slog.Error("no records to store")
				os.Exit(1)
			}

			if dryRun {
				for _, record := range recordsToStore {
					slog.Info("got record", "record", record)
				}
				return
			}
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
				slog.Error("error creating sheets client", "error", err)
				os.Exit(1)
			}
			err = sheetsCli.AppendToSheet(ctx, tablerock.SheetName, recordsToStore)
			if err != nil {
				slog.Error("error appending to sheet", "error", err)
				os.Exit(1)
			}

			dalCli, err := dal.New()
			if err != nil {
				slog.Error("error creating dal client", "error", err)
				os.Exit(1)
			}
			err = dalCli.StoreLakeInfo(ctx, recordsToStore)
			if err != nil {
				slog.Error("error storing lake info", "error", err)
				os.Exit(1)
			}
			slog.Info("successfully stored lake info", "lake", tablerock.SheetName, "records", len(recordsToStore))
		}

	},
}

// latestTableRockRecord returns the most recent complete Table Rock reading.
// It prefers the CWMS Data API (structured JSON, valid TLS, UTC timestamps) and
// falls back to the tab7d HTML scraper when CDA is disabled, errors, or has no
// complete row yet. Set CDA_DISABLED=true to force the HTML scraper.
func latestTableRockRecord(ctx context.Context) (*lakeinfov1.LakeInfoMeasurement, error) {
	if viper.GetBool(cda_disabled) {
		slog.Info("CDA disabled, using HTML scraper", "source", "tab7d")
		return tablerock.GetLatestRecord(tablerock.LakeURL)
	}

	end := time.Now().UTC()
	begin := end.Add(-48 * time.Hour)
	record, err := cda.New().GetLatestCompleteRecord(ctx, begin, end)
	if err != nil {
		slog.Warn("CDA fetch failed, falling back to HTML scraper", "error", err, "source", "tab7d")
		return tablerock.GetLatestRecord(tablerock.LakeURL)
	}
	slog.Info("fetched latest record from CDA",
		"source", "cwms-data-api", "measuredAt", record.MeasuredAt.AsTime())
	return record, nil
}
