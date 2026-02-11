package model

import (
	"fmt"
	"math/rand"

	"gonum.org/v1/gonum/mat"
)

type TrainConfig struct {
	Folds        int
	Epochs       int
	LearningRate float64
	L2Lambda     float64
	Seed         int64
}

type WalkForwardResult struct {
	BaselineNoTrade ConfusionMatrix
	BaselineRandom  ConfusionMatrix
	LogReg          ConfusionMatrix

	LogRegPredictions []PredictionRow
}

func rowsToMatrix(rows []DatasetRow) (*mat.Dense, []Class) {
	X := mat.NewDense(len(rows), 7, nil) // 7 features
	y := make([]Class, len(rows))

	for i, r := range rows {
		X.Set(i, 0, r.Ret1)
		X.Set(i, 1, r.Vol20)
		X.Set(i, 2, r.Mom6)
		X.Set(i, 3, r.EmaSpread)
		X.Set(i, 4, r.RangeHL)
		X.Set(i, 5, r.RangeCO)
		X.Set(i, 6, r.VolChg)
		y[i] = r.Label
	}
	return X, y
}

func EvaluateWalkForward(dataset []DatasetRow, config TrainConfig) (WalkForwardResult, error) {
	if config.Folds < 2 {
		return WalkForwardResult{}, fmt.Errorf("folds must be >= 2")
	}
	if len(dataset) < 50 {
		return WalkForwardResult{}, fmt.Errorf("dataset too small (%d)", len(dataset))
	}

	n := len(dataset)
	testBlock := n / config.Folds
	if testBlock < 10 {
		return WalkForwardResult{}, fmt.Errorf("test block too small (%d); reduce folds", testBlock)
	}

	rng := rand.New(rand.NewSource(config.Seed))

	var result WalkForwardResult

	// Expanding window:
	// fold i: train [0:trainEnd), test [trainEnd:trainEnd+testBlock)
	for fold := 0; fold < config.Folds; fold++ {
		trainEnd := (fold + 1) * testBlock
		testStart := trainEnd
		testEnd := testStart + testBlock

		if testEnd > n {
			break
		}
		if trainEnd < 30 {
			continue
		}

		trainRows := dataset[:trainEnd]
		testRows := dataset[testStart:testEnd]

		// Baseline: always no trade
		predA := PredictAlwaysNoTrade(testRows)
		for i := range testRows {
			result.BaselineNoTrade.Add(testRows[i].Label, predA[i])
		}

		// Baseline: random by train distribution
		predR := PredictRandomByTrainDistribution(trainRows, testRows, rng)
		for i := range testRows {
			result.BaselineRandom.Add(testRows[i].Label, predR[i])
		}

		// Logistic regression
		Xtrain, ytrain := rowsToMatrix(trainRows)
		Xtest, _ := rowsToMatrix(testRows)

		standardizer := FitStandardizer(Xtrain)
		standardizer.TransformInPlace(Xtrain)
		standardizer.TransformInPlace(Xtest)

		model := NewSoftmaxLogReg(3, 7)
		if err := model.FitGradientDescent(Xtrain, ytrain, config.LearningRate, config.L2Lambda, config.Epochs); err != nil {
			return WalkForwardResult{}, err
		}

		P := model.PredictProba(Xtest) // r x 3
		pred := model.Predict(Xtest)

		for i := range testRows {
			actual := testRows[i].Label
			predicted := pred[i]

			result.LogReg.Add(actual, predicted)

			// Save OOS prediction row
			pUp := P.At(i, int(ClassUp))
			pDown := P.At(i, int(ClassDown))
			pNoTrade := P.At(i, int(ClassNoTrade))

			result.LogRegPredictions = append(result.LogRegPredictions, PredictionRow{
				Exchange:  testRows[i].Exchange,
				Symbol:    testRows[i].Symbol,
				Timeframe: testRows[i].Timeframe,
				Timestamp: testRows[i].Timestamp,
				ModelName: "logreg_softmax",

				PUp:      pUp,
				PDown:    pDown,
				PNoTrade: pNoTrade,

				Predicted: predicted,
				Actual:    actual,
			})
		}
	}

	return result, nil
}
