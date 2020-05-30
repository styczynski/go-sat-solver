package sat_solver

import (
	"fmt"
	"math"
	"strings"
)

type SATFormula struct {
	formula *CNFFormula
	vars *SATVariableMapping
	err *UnsatError
}

type SATFormulaStatistics struct {
	variableCount int64
	clauseCount int64
	shortestClause int64
	longestClause int64
}

func (stats SATFormulaStatistics) String() string {
	return fmt.Sprintf("(formula stats: #clauses=%d, #vars=%d, max(|clause|)=%d min(|clause|)=%d)",
		stats.clauseCount, stats.variableCount, stats.shortestClause, stats.longestClause)
}

func NewSATFormulaShortcut(formula *CNFFormula, vars *SATVariableMapping, unsatError *UnsatError) *SATFormula {
	return &SATFormula{
		formula: formula,
		vars:    vars,
		err: unsatError,
	}
}

func NewSATFormula(formula *CNFFormula, vars *SATVariableMapping) *SATFormula {
	return &SATFormula{
		formula: formula,
		vars:    vars,
	}
}

func trimVarQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '[' && s[len(s)-1] == ']' {
			return "var" + s[1: len(s)-1]
		}
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func (f *SATFormula) Variables() *SATVariableMapping {
	return f.vars
}

func (f *SATFormula) Formula() *CNFFormula {
	return f.formula
}

func (f *SATFormula) Measure() SATFormulaStatistics {
	varIDs := map[int64]struct{}{}
	clauseCount := int64(len(f.formula.Variables))
	longestClause := int64(0)
	shortestClause := int64(math.MaxInt64)
	for _, clause := range f.formula.Variables {
		l := int64(len(clause))
		if l > longestClause {
			longestClause = l
		}
		if l < shortestClause {
			shortestClause = l
		}
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
	return SATFormulaStatistics{
		variableCount: varCount,
		clauseCount: clauseCount,
		shortestClause: shortestClause,
		longestClause: longestClause,
	}
}

func (f *SATFormula) FormulaString() string {
	return ""
	result := make([]string, len(f.formula.Variables))
	for j, clause := range f.formula.Variables {
		partialResult := make([]string, len(clause))
		for i, id := range clause {
			v := int64(id)
			if (v == 1) {
				partialResult[i] = "True"
			} else if (v == -1) {
				partialResult[i] = "False"
			} else if (v > 0) {
				partialResult[i] = trimVarQuotes(f.vars.reverse[v])
			} else {
				partialResult[i] = "-" + trimVarQuotes(f.vars.reverse[-v])
			}
		}
		result[j] = "(" + strings.Join(partialResult, " v ") + ")"
	}
	return strings.Join(result, "^")
}

func (f *SATFormula) String() string {
	detailsStr := ""
	if f.err != nil {
		return fmt.Sprintf("UNSAT Formula:\n %s\n %s", f.err.Error(), f.FormulaString())+detailsStr
	}
	detailsStr = f.Measure().String()
	return f.FormulaString()+detailsStr
}

type CNFFormula struct {
	// -1 is false
	// 1 is true
	// x > 1 is a variable with unique id = x
	// x < 1 is a variable with unique id = -x
	Variables [][]int64
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

type SATVariableMapping struct {
	names map[string]int64
	reverse map[int64]string
	uniqueID int64
	freshVarNameID uint64
}

func NewSATVariableMapping() *SATVariableMapping {
	return &SATVariableMapping{
		names:    map[string]int64{},
		reverse:  map[int64]string{},
		uniqueID: 2,
		freshVarNameID: 1,
	}
}

func (vars *SATVariableMapping) Reverse(id int64) string {
	if id < 0 {
		return fmt.Sprintf("-%s", trimVarQuotes(vars.reverse[-id]))
	}
	return trimVarQuotes(vars.reverse[id])
}

func (vars *SATVariableMapping) Fresh() (string, int64) {
	newVarNameID := vars.freshVarNameID
	newID := vars.uniqueID
	name := fmt.Sprintf("[%d]", newVarNameID)
	vars.uniqueID++
	vars.freshVarNameID++
	vars.names[name] = newID
	vars.reverse[newID] = name
	return name, newID
}

func (vars *SATVariableMapping) Get(name string) int64 {
	if id, ok := vars.names[name]; ok {
		return id
	}
	newID := vars.uniqueID
	vars.uniqueID++
	vars.names[name] = newID
	vars.reverse[newID] = name
	return newID
}
