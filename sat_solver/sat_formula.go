package sat_solver

import (
	"fmt"
)

type FormulaRepresentation interface {
	NormalizeVars(vars *SATVariableMapping) *SATVariableMapping
	Evaluate(vars []bool) bool
	Measure() *SATFormulaStatistics
	String(vars *SATVariableMapping) string
}

type SATFormula struct {
	formula FormulaRepresentation
	vars *SATVariableMapping
	err *UnsatError
	stats *SATFormulaStatistics
}

func NewSATFormulaShortcut(formula FormulaRepresentation, vars *SATVariableMapping, stats *SATFormulaStatistics, unsatError *UnsatError) *SATFormula {
	return &SATFormula{
		formula: formula,
		vars:    vars,
		err: unsatError,
		stats: stats,
	}
}

func NewSATFormula(formula FormulaRepresentation, vars *SATVariableMapping, stats *SATFormulaStatistics) *SATFormula {
	return &SATFormula{
		formula: formula,
		vars:    vars,
		stats: stats,
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

func (f *SATFormula) Formula() FormulaRepresentation {
	return f.formula
}

func (f *SATFormula) Stats() SATFormulaStatistics {
	if f.stats == nil {
		f.stats = f.formula.Measure()
	}
	return *f.stats
}

func (f *SATFormula) Brief() string {
	if f.err != nil {
		return fmt.Sprintf("UNSAT Formula: %s", f.err.Error())
	}
	return f.Stats().String()
}

func (f *SATFormula) String() string {
	if f.err != nil {
		return fmt.Sprintf("UNSAT Formula:\n %s\n %s", f.err.Error(), f.formula.String(f.vars))
	}
	return f.formula.String(f.vars)
}
