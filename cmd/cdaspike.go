package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/patrickjmcd/lake-info/lib/cda"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cdaSpikeCmd)
}

// cdaSpikeCmd is a read-only spike: it fetches Table Rock measurements from the
// CWMS Data API and prints them for comparison with the tab7d HTML scraper. It
// does not write to Sheets or Mongo.
var cdaSpikeCmd = &cobra.Command{
	Use:   "cda-spike",
	Short: "Spike: fetch Table Rock data from the CWMS Data API (read-only)",
	Long:  `Fetches Table Rock measurements from the USACE CWMS Data API and prints them. Read-only; nothing is stored.`,
	Run: func(_ *cobra.Command, _ []string) {
		ctx := context.Background()
		client := cda.New()

		end := time.Now().UTC()
		begin := end.Add(-48 * time.Hour)

		records, err := client.GetMeasurements(ctx, begin, end)
		if err != nil {
			slog.Error("error fetching from CDA", "error", err)
			os.Exit(1)
		}
		if len(records) == 0 {
			slog.Error("no records returned")
			os.Exit(1)
		}

		fmt.Printf("%-20s %10s %8s %10s %10s %10s\n",
			"measuredAt (UTC)", "level", "gen", "turbine", "spillway", "total")
		for _, r := range records {
			fmt.Printf("%-20s %10.2f %8.1f %10.1f %10.1f %10.1f\n",
				r.MeasuredAt.AsTime().UTC().Format("2006-01-02 15:04"),
				r.Level, r.Generation, r.TurbineReleaseRate,
				r.SpillwayReleaseRate, r.TotalReleaseRate)
		}

		latest := records[len(records)-1]
		fmt.Printf("\nlatest: %s  level=%.2f ft  turbine=%.0f  spillway=%.0f  total=%.0f cfs  gen=%.0f MWh\n",
			latest.MeasuredAt.AsTime().UTC().Format(time.RFC3339),
			latest.Level, latest.TurbineReleaseRate, latest.SpillwayReleaseRate,
			latest.TotalReleaseRate, latest.Generation)
	},
}
