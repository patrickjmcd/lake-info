package cmd

import (
	"fmt"
	"github.com/patrickjmcd/lake-info/dal"
	"github.com/patrickjmcd/lake-info/gen/lakeinfo/v1/lakeinfoconnect"
	"github.com/patrickjmcd/lake-info/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log/slog"
	"net/http"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("port", "p", "8080", "The port to listen on")
	viper.BindPFlag(port, serveCmd.Flags().Lookup("port"))
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the Lake Info service",
	Long:  `Runs the Lake Info service function provided during CLI initialization.`,
	Run: func(_ *cobra.Command, _ []string) {
		db, err := dal.New()
		if err != nil {
			log.Fatal().Err(err).Msg("error creating dal client")
		}
		lakeInfoServer := server.New(db)
		mux := http.NewServeMux()
		path, handler := lakeinfoconnect.NewLakeInfoServiceHandler(lakeInfoServer)
		mux.Handle(path, handler)

		httpPort := viper.GetString(port)

		address := fmt.Sprintf("localhost:%s", httpPort)

		slog.Info(fmt.Sprintf("Starting server on localhost:%s", httpPort))
		http.ListenAndServe(
			address,
			// Use h2c so we can serve HTTP/2 without TLS.
			h2c.NewHandler(mux, &http2.Server{}),
		)
	},
}
