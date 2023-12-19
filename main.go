package main

import (
	"context"
	"fmt"
	"github.com/patrickjmcd/lake-info/dal"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"github.com/patrickjmcd/lake-info/gen/lakeinfo/v1/lakeinfoconnect"
	"github.com/patrickjmcd/lake-info/lib/tablerock"
	"github.com/patrickjmcd/lake-info/server"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"
)

var (
	port       string
	allRecords bool
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
			dalCli, err := dal.New()
			var recordsToStore []*lakeinfov1.LakeInfoMeasurement

			if allRecords {
				if r, err := tablerock.GetAllRecords(tablerock.LakeURL); err != nil {
					log.Fatal(err)
				} else {
					recordsToStore = r
				}
			} else {
				record, err := tablerock.GetLatestRecord(tablerock.LakeURL)
				if err != nil {
					log.Fatal(err)
				}
				recordsToStore = append(recordsToStore, record)
			}

			err = dalCli.StoreLakeInfo(ctx, recordsToStore)
			if err != nil {
				log.Fatal(err)
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

		log.Println("Starting server on localhost:8080")
		http.ListenAndServe(
			fmt.Sprintf("localhost:%s", port),
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
			panic(err)
		}
		err = dalClient.Setup(ctx)
		if err != nil {
			panic(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(scrapeCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(setupCmd)
	scrapeCmd.Flags().BoolVarP(&allRecords, "all", "A", false, "Get all records")
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to serve on")
}
