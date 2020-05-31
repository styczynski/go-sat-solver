package sat_solver

import (
	"fmt"
)

type FormulaRepresentation interface {
	NormalizeVars(vars *SATVariableMapping) (error, *SATVariableMapping, int)
	Evaluate(vars []bool) bool
	Measure() *SATFormulaStatistics
	String(vars *SATVariableMapping) string
	AST(vars *SATVariableMapping) *Formula
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

func (f* SATFormula) AST() *Formula {
	return f.formula.AST(f.vars)
}

func (f *SATFormula) Normalize() (error, []bool) {
	err, newVars, varCount := f.formula.NormalizeVars(f.vars)
	if err != nil {
		return err, nil
	}
	f.vars = newVars
	return nil, make([]bool, varCount)
}

func (f * SATFormula) Evaluate(vars []bool) bool {
	return f.formula.Evaluate(vars)
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
