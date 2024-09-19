package cmd

import (
	"context"
	"github.com/patrickjmcd/lake-info/dal"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

func init() {
	rootCmd.AddCommand(setupCmd)
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
