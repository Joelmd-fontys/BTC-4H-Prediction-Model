package model

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

type Standardizer struct {
	Mean []float64
	Std  []float64
}

func FitStandardizer(X *mat.Dense) Standardizer {
	_, d := X.Dims()
	mean := make([]float64, d)
	std := make([]float64, d)

	r, _ := X.Dims()
	for j := 0; j < d; j++ {
		var sum float64
		for i := 0; i < r; i++ {
			sum += X.At(i, j)
		}
		mean[j] = sum / float64(r)
	}

	for j := 0; j < d; j++ {
		var s float64
		for i := 0; i < r; i++ {
			diff := X.At(i, j) - mean[j]
			s += diff * diff
		}
		// population std
		v := s / float64(r)
		std[j] = math.Sqrt(v)
		if std[j] == 0 {
			std[j] = 1
		}
	}

	return Standardizer{Mean: mean, Std: std}
}

func (s Standardizer) TransformInPlace(X *mat.Dense) {
	r, d := X.Dims()
	for i := 0; i < r; i++ {
		for j := 0; j < d; j++ {
			X.Set(i, j, (X.At(i, j)-s.Mean[j])/s.Std[j])
		}
	}
}

// Softmax logistic regression weights: K x (d+1) (bias included as last column)
type SoftmaxLogReg struct {
	W *mat.Dense
}

func NewSoftmaxLogReg(numClasses int, numFeatures int) SoftmaxLogReg {
	// W dims: K x (d+1)
	return SoftmaxLogReg{W: mat.NewDense(numClasses, numFeatures+1, nil)}
}

func addBiasColumn(X *mat.Dense) *mat.Dense {
	r, d := X.Dims()
	Xb := mat.NewDense(r, d+1, nil)
	for i := 0; i < r; i++ {
		for j := 0; j < d; j++ {
			Xb.Set(i, j, X.At(i, j))
		}
		Xb.Set(i, d, 1.0) // bias
	}
	return Xb
}

func softmaxRow(scores []float64) []float64 {
	// stable softmax
	max := scores[0]
	for _, v := range scores[1:] {
		if v > max {
			max = v
		}
	}
	var sum float64
	out := make([]float64, len(scores))
	for i, v := range scores {
		ev := math.Exp(v - max)
		out[i] = ev
		sum += ev
	}
	for i := range out {
		out[i] /= sum
	}
	return out
}

func (m SoftmaxLogReg) FitGradientDescent(
	X *mat.Dense, // r x d
	y []Class,
	learningRate float64,
	l2Lambda float64,
	epochs int,
) error {
	r, d := X.Dims()
	if len(y) != r {
		return fmt.Errorf("y length mismatch")
	}
	K, wD := m.W.Dims()
	if wD != d+1 {
		return fmt.Errorf("model expects %d features, got %d", wD-1, d)
	}

	Xb := addBiasColumn(X) // r x (d+1)
	_, dB := Xb.Dims()

	// Gradient: K x (d+1)
	grad := mat.NewDense(K, dB, nil)

	for epoch := 0; epoch < epochs; epoch++ {
		grad.Zero()

		// accumulate gradient over samples
		for i := 0; i < r; i++ {
			// scores[k] = W_k dot x_i
			scores := make([]float64, K)
			for k := 0; k < K; k++ {
				var s float64
				for j := 0; j < dB; j++ {
					s += m.W.At(k, j) * Xb.At(i, j)
				}
				scores[k] = s
			}
			prob := softmaxRow(scores)

			yi := int(y[i])
			for k := 0; k < K; k++ {
				indicator := 0.0
				if k == yi {
					indicator = 1.0
				}
				diff := prob[k] - indicator // (p - y)
				for j := 0; j < dB; j++ {
					grad.Set(k, j, grad.At(k, j)+diff*Xb.At(i, j))
				}
			}
		}

		// average + L2 (don’t regularize bias column if you want; we’ll regularize all for simplicity)
		scale := 1.0 / float64(r)
		for k := 0; k < K; k++ {
			for j := 0; j < dB; j++ {
				g := grad.At(k, j)*scale + l2Lambda*m.W.At(k, j)
				m.W.Set(k, j, m.W.At(k, j)-learningRate*g)
			}
		}
	}

	return nil
}

func (m SoftmaxLogReg) PredictProba(X *mat.Dense) *mat.Dense {
	r, d := X.Dims()
	K, wD := m.W.Dims()
	if wD != d+1 {
		panic("feature dimension mismatch")
	}

	Xb := addBiasColumn(X)
	_, dB := Xb.Dims()

	P := mat.NewDense(r, K, nil)

	for i := 0; i < r; i++ {
		scores := make([]float64, K)
		for k := 0; k < K; k++ {
			var s float64
			for j := 0; j < dB; j++ {
				s += m.W.At(k, j) * Xb.At(i, j)
			}
			scores[k] = s
		}
		prob := softmaxRow(scores)
		for k := 0; k < K; k++ {
			P.Set(i, k, prob[k])
		}
	}

	return P
}

func (m SoftmaxLogReg) Predict(X *mat.Dense) []Class {
	P := m.PredictProba(X)
	r, _ := P.Dims()
	out := make([]Class, r)

	for i := 0; i < r; i++ {
		bestK := 0
		bestV := P.At(i, 0)
		for k := 1; k < 3; k++ {
			v := P.At(i, k)
			if v > bestV {
				bestV = v
				bestK = k
			}
		}
		out[i] = Class(bestK)
	}
	return out
}
