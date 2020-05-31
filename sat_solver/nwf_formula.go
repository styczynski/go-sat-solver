package sat_solver

import "fmt"

type NWFAnd struct {
	Arg1  *NWFFormula
	Arg2  *NWFFormula
	IsNeg bool
}

func (f *NWFAnd) Evaluate(vars []bool) bool {
	return (f.Arg1.Evaluate(vars) && f.Arg2.Evaluate(vars)) != f.IsNeg
}

func (f *NWFAnd) String(vars *SATVariableMapping) string {
	if f.IsNeg {
		return fmt.Sprintf("-(%s ^ %s)", f.Arg1.String(vars), f.Arg2.String(vars))
	}
	return fmt.Sprintf("(%s ^ %s)", f.Arg1.String(vars), f.Arg2.String(vars))
}

func (f *NWFAnd) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping) {
	f.Arg1.recNormalizeVars(vars, newVars)
	f.Arg2.recNormalizeVars(vars, newVars)
}

type NWFOr struct {
	Arg1  *NWFFormula
	Arg2  *NWFFormula
	IsNeg bool
}

func (f *NWFOr) Evaluate(vars []bool) bool {
	return (f.Arg1.Evaluate(vars) || f.Arg2.Evaluate(vars)) != f.IsNeg
}

func (f *NWFOr) String(vars *SATVariableMapping) string {
	if f.IsNeg {
		return fmt.Sprintf("-(%s v %s)", f.Arg1.String(vars), f.Arg2.String(vars))
	}
	return fmt.Sprintf("(%s v %s)", f.Arg1.String(vars), f.Arg2.String(vars))
}

func (f *NWFOr) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping) {
	f.Arg1.recNormalizeVars(vars, newVars)
	f.Arg2.recNormalizeVars(vars, newVars)
}

type NWFVar struct {
	ID    int64
	IsNeg bool
}

func (f *NWFVar) Evaluate(vars []bool) bool {
	if f.ID < 0 {
		return vars[-f.ID-1] != f.IsNeg
	}
	return vars[f.ID-1] != f.IsNeg
}

func (f *NWFVar) String(vars *SATVariableMapping) string {
	if f.IsNeg {
		return fmt.Sprintf("-%s", vars.Reverse(f.ID))
	}
	return fmt.Sprintf("%s", vars.Reverse(f.ID))
}

func (f *NWFVar) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping) {
	f.ID = newVars.Get(vars.reverse[f.ID])
}

type NWFConst struct {
	Value bool
}

func (f *NWFConst) Evaluate(vars []bool) bool {
	return f.Value
}

func (f *NWFConst) String(vars *SATVariableMapping) string {
	if f.Value {
		return "T"
	}
	return "F"
}

func (f *NWFConst) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping) {}

type NWFFormula struct {
	And      *NWFAnd
	Or       *NWFOr
	Variable *NWFVar
	Const    *NWFConst
}

func (f *NWFFormula) Evaluate(vars []bool) bool {
	if f.And != nil {
		return f.And.Evaluate(vars)
	} else if f.Or != nil {
		return f.Or.Evaluate(vars)
	} else if f.Const != nil {
		return f.Const.Evaluate(vars)
	} else if f.Variable != nil {
		return f.Variable.Evaluate(vars)
	}
	return false
}

func (f *NWFFormula) String(vars *SATVariableMapping) string {
	if f.And != nil {
		return f.And.String(vars)
	} else if f.Or != nil {
		return f.Or.String(vars)
	} else if f.Const != nil {
		return f.Const.String(vars)
	} else if f.Variable != nil {
		return f.Variable.String(vars)
	}
	return ""
}

func (f *NWFFormula) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping) {
	if f.And != nil {
		f.And.recNormalizeVars(vars, newVars)
	} else if f.Or != nil {
		f.Or.recNormalizeVars(vars, newVars)
	} else if f.Const != nil {
		f.Const.recNormalizeVars(vars, newVars)
	} else if f.Variable != nil {
		f.Variable.recNormalizeVars(vars, newVars)
	}
}

func (f *NWFFormula) Negate() {
	if f.Variable != nil {
		f.Variable.IsNeg = !f.Variable.IsNeg
	} else if f.And != nil {
		f.And.IsNeg = !f.And.IsNeg
	} else if f.Or != nil {
		f.Or.IsNeg = !f.Or.IsNeg
	} else if f.Const != nil {
		f.Const.Value = !f.Const.Value
	}
}

func (f *NWFFormula) NormalizeVars(vars *SATVariableMapping) (error, *SATVariableMapping, int) {
	newVars := &SATVariableMapping{
		names:    map[string]int64{},
		reverse:  map[int64]string{},
		uniqueID: 1,
		freshVarNameID: vars.freshVarNameID,
	}
	f.recNormalizeVars(vars, newVars)
	return nil, newVars, len(newVars.names)
}

func (f *NWFFormula) Measure() *SATFormulaStatistics {
	return &SATFormulaStatistics{
		variableCount: 0,
		clauseCount:   0,
		clauseLenSum:  0,
	}
}
