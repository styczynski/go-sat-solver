package sat_solver

import (
	"fmt"
	"math"
	"strings"
)

type CNFFormula struct {
	// -1 is false
	// 1 is true
	// x > 1 is a variable with unique id = x
	// x < 1 is a variable with unique id = -x
	Variables [][]int64
}

func (f *CNFFormula) AST(vars *SATVariableMapping) *Formula {
	var currentRet *Formula = nil
	for _, clause := range f.Variables {
		var currentClause *Formula = nil
		for _, v := range clause {
			if currentClause == nil {
				if v < 0 {
					currentClause = MakeNot(MakeVar(vars.Reverse(v)))
				} else {
					currentClause = MakeVar(vars.Reverse(v))
				}
			} else {
				if v < 0 {
					currentClause = MakeOr(currentClause, MakeNot(MakeVar(vars.Reverse(v))))
				} else {
					currentClause = MakeOr(currentClause, MakeVar(vars.Reverse(v)))
				}
			}
		}
		if currentRet == nil {
			currentRet = currentClause
		} else {
			currentRet = MakeAnd(currentRet, currentClause)
		}
	}
	return currentRet
}

type UnsatReasonCNFNormalization struct {}

func NewUnsatReasonCNFNormalization() *UnsatReasonCNFNormalization {
	return &UnsatReasonCNFNormalization{}
}

func (reason *UnsatReasonCNFNormalization) Describe() string {
	return fmt.Sprintf("Empty clause detected when normalizing CNF formula.")
}

func (f *CNFFormula) NormalizeVars(vars *SATVariableMapping) (error, *SATVariableMapping, int) {
	newVars := &SATVariableMapping{
		names:    map[string]int64{},
		reverse:  map[int64]string{},
		uniqueID: 1,
		freshVarNameID: vars.freshVarNameID,
	}
	newMapping := map[int64]int64{}

	for i, clause := range f.Variables {
		if len(clause) == 0 {
			return NewUnsatError(NewUnsatReasonCNFNormalization()), nil, 0
		}
		for j, v := range clause {
			varID := v
			neg := false
			if varID < 0 {
				varID = -varID
				neg = true
			}
			if entry, ok := newMapping[varID]; ok {
				if neg {
					f.Variables[i][j] = -entry
				} else {
					f.Variables[i][j] = entry
				}
			} else {
				newID := newVars.uniqueID
				newMapping[varID] = newID
				if neg {
					f.Variables[i][j] = -newID
				} else {
					f.Variables[i][j] = newID
				}

				newVars.reverse[newID] = vars.reverse[varID]
				newVars.names[vars.reverse[varID]] = newID

				newVars.uniqueID++
			}
		}
	}

	return nil, newVars, len(newVars.names)
}

func (f *CNFFormula) Evaluate(vars []bool) bool {
	for _, clause := range f.Variables {
		isClauseOk := false
		for _, v := range clause {
			if v > 0 && vars[v-1] {
				isClauseOk = true
				break
			} else if v < 0 && !vars[-v-1] {
				isClauseOk = true
				break
			}
		}
		if !isClauseOk {
			return false
		}
	}
	return true
}

func (f *CNFFormula) AndWith(e *CNFFormula) {
	f.Variables = append(f.Variables, e.Variables...)
}

func (f *CNFFormula) MulWith(e *CNFFormula) {
	lenF := len(f.Variables)
	lenE := len(e.Variables)
	newVariables := make([][]int64, 0, lenF * lenE)
	shouldAddE := true
	shouldAddF := true

	for _, clauseF := range f.Variables {
		shouldAddF = true
		for _, varF := range clauseF {
			if varF == 1 {
				shouldAddF = false
				break
			}
		}
		if !shouldAddF {
			continue
		}
		for _, clauseE := range e.Variables {
			shouldAddE = true
			for _, varE := range clauseE {
				if varE == 1 {
					shouldAddE = false
					break
				}
			}
			if shouldAddE {
				newVariables = append(newVariables, append(clauseF, clauseE...))
			}
		}
	}
	f.Variables = newVariables
}

func (f *CNFFormula) Measure() *SATFormulaStatistics {
	varIDs := map[int64]struct{}{}
	clauseCount := int64(len(f.Variables))
	longestClause := int64(0)
	shortestClause := int64(math.MaxInt64)
	clauseLenSum := int64(0)
	for _, clause := range f.Variables {
		l := int64(len(clause))
		if l > longestClause {
			longestClause = l
		}
		if l < shortestClause {
			shortestClause = l
		}
		clauseLenSum += l
		for _, varID := range clause {
			if varID > 0 {
				varIDs[varID] = struct{}{}
			} else {
				varIDs[-varID] = struct{}{}
			}
		}
	}
	varCount := int64(0)
	for range varIDs {
		varCount++
	}
	return &SATFormulaStatistics{
		variableCount:    varCount,
		clauseCount:      clauseCount,
		clauseLenSum:     clauseLenSum,
		clauseDepth:      2,
		clauseComplexity: clauseLenSum,
	}
}

func (f *CNFFormula) String(vars *SATVariableMapping) string {
	result := make([]string, len(f.Variables))
	for j, clause := range f.Variables {
		partialResult := make([]string, len(clause))
		for i, id := range clause {
			v := int64(id)
			if (v == 1) {
				partialResult[i] = "True"
			} else if (v == -1) {
				partialResult[i] = "False"
			} else if (v > 0) {
				partialResult[i] = trimVarQuotes(vars.reverse[v])
			} else {
				partialResult[i] = "-" + trimVarQuotes(vars.reverse[-v])
			}
		}
		result[j] = "(" + strings.Join(partialResult, " v ") + ")"
	}
	return strings.Join(result, "^")
}