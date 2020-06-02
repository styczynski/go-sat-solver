package sat_solver

import (
	"fmt"
	"math"
)

type SATFormulaStatistics struct {
	variableCount int64
	clauseCount int64
	clauseLenSum int64
	clauseDepth int64
	clauseComplexity int64
	isCNF bool
}

func (stats SATFormulaStatistics) Score() float64 {
	if !stats.isCNF {
		return (float64(stats.clauseDepth) * float64(stats.clauseDepth) + float64(stats.clauseComplexity)) / 10
	}
	return float64(stats.clauseCount) + math.Exp(float64(stats.variableCount)/100)/4 + 5 * float64(stats.clauseLenSum) / float64(stats.clauseCount)
}

func (stats SATFormulaStatistics) String() string {
	if !stats.isCNF {
		return fmt.Sprintf("scoreNWF=%.0f, depth=%d, complexity=%d", stats.Score(), stats.clauseDepth, stats.clauseComplexity)
	}
	return fmt.Sprintf("scoreCNF=%.0f, #clauses=%d, #vars=%d, avg(|clause|)=%.2f",
		stats.Score(), stats.clauseCount, stats.variableCount, float64(stats.clauseLenSum) / float64(stats.clauseCount) )
}
