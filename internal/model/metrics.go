package model

import "fmt"

type ConfusionMatrix struct {
	// rows = actual, cols = predicted
	M [3][3]int
}

func (cm *ConfusionMatrix) Add(actual Class, predicted Class) {
	cm.M[int(actual)][int(predicted)]++
}

func (cm ConfusionMatrix) Total() int {
	total := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			total += cm.M[i][j]
		}
	}
	return total
}

func (cm ConfusionMatrix) Accuracy() float64 {
	correct := cm.M[0][0] + cm.M[1][1] + cm.M[2][2]
	total := cm.Total()
	if total == 0 {
		return 0
	}
	return float64(correct) / float64(total)
}

func precisionRecallForClass(cm ConfusionMatrix, class Class) (precision float64, recall float64) {
	k := int(class)

	// precision = TP / (TP + FP) where FP = predicted k but actual != k
	tp := cm.M[k][k]
	fp := 0
	for actual := 0; actual < 3; actual++ {
		if actual == k {
			continue
		}
		fp += cm.M[actual][k]
	}

	// recall = TP / (TP + FN) where FN = actual k but predicted != k
	fn := 0
	for predicted := 0; predicted < 3; predicted++ {
		if predicted == k {
			continue
		}
		fn += cm.M[k][predicted]
	}

	if tp+fp > 0 {
		precision = float64(tp) / float64(tp+fp)
	}
	if tp+fn > 0 {
		recall = float64(tp) / float64(tp+fn)
	}
	return
}

func (cm ConfusionMatrix) SummaryString() string {
	upP, upR := precisionRecallForClass(cm, ClassUp)
	downP, downR := precisionRecallForClass(cm, ClassDown)

	return fmt.Sprintf(
		"accuracy=%.4f | UP(p=%.4f r=%.4f) | DOWN(p=%.4f r=%.4f) | cm=[[ %d %d %d ],[ %d %d %d ],[ %d %d %d ]]",
		cm.Accuracy(),
		upP, upR,
		downP, downR,
		cm.M[0][0], cm.M[0][1], cm.M[0][2],
		cm.M[1][0], cm.M[1][1], cm.M[1][2],
		cm.M[2][0], cm.M[2][1], cm.M[2][2],
	)
}
