package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"btc-4h-prediction-model/internal/model"
	"btc-4h-prediction-model/internal/store"
)

var trainSymbol string
var trainTimeframe string

var trainFolds int
var trainEpochs int
var trainLearningRate float64
var trainL2Lambda float64
var trainSeed int64

var trainCommand = &cobra.Command{
	Use:   "train",
	Short: "Walk-forward baselines + multinomial logistic regression on dataset view",
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()

		db, err := store.OpenSQLite(databasePath)
		if err != nil {
			return err
		}
		defer db.Close()

		datasetRows, err := model.LoadDatasetOrdered(ctx, db, "binance", trainSymbol, trainTimeframe)
		if err != nil {
			return err
		}

		fmt.Println("db:", databasePath)
		fmt.Println("exchange: binance")
		fmt.Println("symbol:", trainSymbol)
		fmt.Println("timeframe:", trainTimeframe)
		fmt.Println("dataset rows:", len(datasetRows))

		result, err := model.EvaluateWalkForward(datasetRows, model.TrainConfig{
			Folds:        trainFolds,
			Epochs:       trainEpochs,
			LearningRate: trainLearningRate,
			L2Lambda:     trainL2Lambda,
			Seed:         trainSeed,
		})
		if err != nil {
			return err
		}

		fmt.Println("always NO_TRADE:", result.BaselineNoTrade.SummaryString())
		fmt.Println("random baseline:", result.BaselineRandom.SummaryString())
		fmt.Println("logreg softmax:", result.LogReg.SummaryString())

		return nil
	},
}

func init() {
	trainCommand.Flags().StringVar(&trainSymbol, "symbol", "BTCUSDT", "Symbol (e.g. BTCUSDT)")
	trainCommand.Flags().StringVar(&trainTimeframe, "timeframe", "4h", "Timeframe (e.g. 4h)")

	trainCommand.Flags().IntVar(&trainFolds, "folds", 5, "Number of walk-forward folds")
	trainCommand.Flags().IntVar(&trainEpochs, "epochs", 500, "Training epochs for logistic regression")
	trainCommand.Flags().Float64Var(&trainLearningRate, "lr", 0.5, "Learning rate")
	trainCommand.Flags().Float64Var(&trainL2Lambda, "l2", 0.001, "L2 regularization strength")
	trainCommand.Flags().Int64Var(&trainSeed, "seed", 42, "Random seed for baselines")
}
