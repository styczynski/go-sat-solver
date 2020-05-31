package sat_solver

import (
	"fmt"
	"math"
)

type SATFormulaStatistics struct {
	variableCount int64
	clauseCount int64
	clauseLenSum int64
}

func (stats SATFormulaStatistics) Score() float64 {
	return float64(stats.clauseCount) + math.Exp(float64(stats.variableCount)/100)/4 + 5 * float64(stats.clauseLenSum) / float64(stats.clauseCount)
}

func (stats SATFormulaStatistics) String() string {
	return fmt.Sprintf("score=%.0f, #clauses=%d, #vars=%d, avg(|clause|)=%.2f",
		stats.Score(), stats.clauseCount, stats.variableCount, float64(stats.clauseLenSum) / float64(stats.clauseCount) )
}
