package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:   "quant",
	Short: "Quantitative research CLI",
}

var databasePath string

func Execute() {
	rootCommand.PersistentFlags().StringVar(
		&databasePath,
		"db",
		"btcqd.sqlite",
		"Path to SQLite database file",
	)

	rootCommand.AddCommand(ingestCommand)
	rootCommand.AddCommand(validateCommand)
	rootCommand.AddCommand(featuresCommand)
	rootCommand.AddCommand(labelsCommand)
	rootCommand.AddCommand(trainCommand)
	rootCommand.AddCommand(confidenceCommand)
	rootCommand.AddCommand(paperCommand)

	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
