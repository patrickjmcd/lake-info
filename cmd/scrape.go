package cmd

import (
	"context"
	"github.com/patrickjmcd/lake-info/dal"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"github.com/patrickjmcd/lake-info/lib/tablerock"
	"github.com/patrickjmcd/lake-info/pkg/sheets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
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
				record, err := tablerock.GetLatestRecord(tablerock.LakeURL)
				if err != nil {
					slog.Error("error getting latest record", "error", err)
				}
				recordsToStore = append(recordsToStore, record)
			}

			if dryRun {
				for _, record := range recordsToStore {
					slog.Info("got record", "record", record)
				}
				return
			}
			spreadsheetId := viper.GetString(spreadsheet_id)
			googleSAB64 := viper.GetString(google_sa_b64)

			sheetsCli, err := sheets.New(ctx, spreadsheetId, nil, &googleSAB64)
			if err != nil {
				slog.Error("error creating sheets client", "error", err)
				os.Exit(1)
			}
			err = sheetsCli.WriteMeasurements(ctx, sheets.TableRockLake, recordsToStore)

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
			slog.Info("successfully stored lake info", "lake", sheets.TableRockLake, "records", len(recordsToStore))
		}

	},
}
