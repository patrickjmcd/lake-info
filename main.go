package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/patrickjmcd/lake-info/dal"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"github.com/patrickjmcd/lake-info/gen/lakeinfo/v1/lakeinfoconnect"
	"github.com/patrickjmcd/lake-info/lib/tablerock"
	"github.com/patrickjmcd/lake-info/logger"
	"github.com/patrickjmcd/lake-info/server"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	port       string
	allRecords bool
	dryRun     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lake-info",
	Short: "Lake Info CLI",
	Long:  `Lake Info CLI for interacting with USACE Lake Info.`,
}

func main() {
	// Execute the root command
	err := rootCmd.Execute()
	if err != nil {
		log.Println("error executing root command")
	}
}

var scrapeCmd = &cobra.Command{
	Use:   "scrape [lake]",
	Short: "Scrape lake info",
	Long:  `Scrape lake info from USACE Lake Info.`,
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {

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
		}

	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the Lake Info service",
	Long:  `Runs the Lake Info service function provided during CLI initialization.`,
	Run: func(_ *cobra.Command, _ []string) {
		db, err := dal.New()
		if err != nil {
			log.Fatal(err)
		}
		lakeInfoServer := server.New(db)
		mux := http.NewServeMux()
		path, handler := lakeinfoconnect.NewLakeInfoServiceHandler(lakeInfoServer)
		mux.Handle(path, handler)
		address := fmt.Sprintf("localhost:%s", port)

		slog.Info(fmt.Sprintf("Starting server on localhost:%s", port))
		http.ListenAndServe(
			address,
			// Use h2c so we can serve HTTP/2 without TLS.
			h2c.NewHandler(mux, &http2.Server{}),
		)
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the database",
	Long:  `Setup the database. This command will create the database and run migrations.`,
	Run: func(_ *cobra.Command, _ []string) {
		ctx := context.Background()
		dalClient, err := dal.New()
		if err != nil {
			slog.Error("error creating dal client", "error", err)
			os.Exit(1)
		}
		err = dalClient.Setup(ctx)
		if err != nil {
			slog.Error("error setting up database", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	logger.Setup()
	rootCmd.AddCommand(scrapeCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(setupCmd)
	scrapeCmd.Flags().BoolVarP(&allRecords, "all", "A", false, "Get all records")
	scrapeCmd.Flags().BoolVarP(&dryRun, "dry-run", "D", false, "Dry run")
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to serve on")
}
