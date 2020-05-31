package sat_solver

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type NWFAnd struct {
	Arg1  *NWFFormula
	Arg2  *NWFFormula
	IsNeg bool
	Depth int32
	Complexity int32
}

func (f *NWFAnd) AST(vars *SATVariableMapping) *Formula {
	return MakeAnd(f.Arg1.AST(vars), f.Arg2.AST(vars))
}

func (f *NWFAnd) UpdateTopNodeMetrics() {
	d1, c1 := f.Arg1.NodeMetrics()
	d2, c2 := f.Arg2.NodeMetrics()

	d := d1
	if d2 > d {
		d = d2
	}

	f.Depth = d
	f.Complexity = c1 + c2
}

func (f *NWFAnd) NodeMetrics() (int32, int32) {
	return f.Depth, f.Complexity
}

func (f *NWFAnd) Evaluate(vars []bool) bool {
	return (f.Arg1.Evaluate(vars) && f.Arg2.Evaluate(vars)) != f.IsNeg
}

func (f *NWFAnd) Serialize() string {
	children := []string{ f.Arg1.Serialize(), f.Arg2.Serialize() }
	sort.Strings(children)
	if f.IsNeg {
		return "{" + strings.Join(children, "*") + "}"
	}
	return "(" + strings.Join(children, "*") + ")"
}

func (f *NWFAnd) String(vars *SATVariableMapping) string {
	if f.IsNeg {
		return fmt.Sprintf("-(%s ^ %s)", f.Arg1.String(vars), f.Arg2.String(vars))
	}
	return fmt.Sprintf("(%s ^ %s)", f.Arg1.String(vars), f.Arg2.String(vars))
}

func (f *NWFAnd) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping, notified *map[*NWFVar]struct{}) {
	f.Arg1.recNormalizeVars(vars, newVars, notified)
	f.Arg2.recNormalizeVars(vars, newVars, notified)
}

type NWFOr struct {
	Arg1  *NWFFormula
	Arg2  *NWFFormula
	IsNeg bool
	Depth int32
	Complexity int32
}

func (f *NWFOr) AST(vars *SATVariableMapping) *Formula {
	return MakeOr(f.Arg1.AST(vars), f.Arg2.AST(vars))
}

func (f *NWFOr) UpdateTopNodeMetrics() {
	d1, c1 := f.Arg1.NodeMetrics()
	d2, c2 := f.Arg2.NodeMetrics()

	d := d1
	if d2 > d {
		d = d2
	}

	f.Depth = d
	f.Complexity = c1 + c2
}

func (f *NWFOr) NodeMetrics() (int32, int32) {
	return f.Depth, f.Complexity
}

func (f *NWFOr) Evaluate(vars []bool) bool {
	return (f.Arg1.Evaluate(vars) || f.Arg2.Evaluate(vars)) != f.IsNeg
}

func (f *NWFOr) Serialize() string {
	children := []string{ f.Arg1.Serialize(), f.Arg2.Serialize() }
	sort.Strings(children)
	if f.IsNeg {
		return "{" + strings.Join(children, "+") + "}"
	}
	return "(" + strings.Join(children, "+") + ")"
}

func (f *NWFOr) String(vars *SATVariableMapping) string {
	if f.IsNeg {
		return fmt.Sprintf("-(%s v %s)", f.Arg1.String(vars), f.Arg2.String(vars))
	}
	return fmt.Sprintf("(%s v %s)", f.Arg1.String(vars), f.Arg2.String(vars))
}

func (f *NWFOr) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping, notified *map[*NWFVar]struct{}) {
	f.Arg1.recNormalizeVars(vars, newVars, notified)
	f.Arg2.recNormalizeVars(vars, newVars, notified)
}

type NWFVar struct {
	ID    int64
}

func (f *NWFVar) AST(vars *SATVariableMapping) *Formula {
	if f.ID < 0 {
		return MakeNot(MakeVar(vars.Reverse(-f.ID)))
	}
	return MakeVar(vars.Reverse(f.ID))
}

func (f *NWFVar) UpdateTopNodeMetrics() {}

func (f *NWFVar) NodeMetrics() (int32, int32) {
	return 1, 1
}

func (f *NWFVar) Evaluate(vars []bool) bool {
	if f.ID < 0 {
		return !vars[-f.ID-1]
	}
	return vars[f.ID-1]
}

func (f *NWFVar) Serialize() string {
	return strconv.Itoa(int(f.ID))
}

func (f *NWFVar) String(vars *SATVariableMapping) string {
	if f.ID < 0 {
		return fmt.Sprintf("-%s", vars.Reverse(-f.ID))
	}
	return fmt.Sprintf("%s", vars.Reverse(f.ID))
}

func (f *NWFVar) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping, notified *map[*NWFVar]struct{}) {
	if _, ok := (*notified)[f]; ok {
		return
	}
	if f.ID < 0 {
		f.ID = -newVars.Get(vars.reverse[-f.ID])
		(*notified)[f] = struct{}{}
		return
	}
	f.ID = newVars.Get(vars.reverse[f.ID])
	(*notified)[f] = struct{}{}
}

type NWFConst struct {
	Value bool
}

func (f* NWFConst) AST(vars *SATVariableMapping) *Formula {
	return MakeBoolConstant(f.Value)
}

func (f *NWFConst) UpdateTopNodeMetrics() {}

func (f *NWFConst) NodeMetrics() (int32, int32) {
	return 1, 1
}

func (f *NWFConst) Evaluate(vars []bool) bool {
	return f.Value
}

func (f *NWFConst) Serialize() string {
	if f.Value {
		return "t"
	}
	return "f"
}

func (f *NWFConst) String(vars *SATVariableMapping) string {
	if f.Value {
		return "T"
	}
	return "F"
}

func (f *NWFConst) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping, notified *map[*NWFVar]struct{}) {}

type NWFFormula struct {
	And      *NWFAnd
	Or       *NWFOr
	Variable *NWFVar
	Const    *NWFConst
}

func (f *NWFFormula) UpdateTopNodeMetrics() *NWFFormula {
	if f.And != nil {
		f.And.UpdateTopNodeMetrics()
	} else if f.Or != nil {
		f.Or.UpdateTopNodeMetrics()
	} else if f.Const != nil {
		f.Const.UpdateTopNodeMetrics()
	} else if f.Variable != nil {
		f.Variable.UpdateTopNodeMetrics()
	}
	return f
}

func (f *NWFFormula) NodeMetrics() (int32, int32) {
	if f.And != nil {
		return f.And.NodeMetrics()
	} else if f.Or != nil {
		return f.Or.NodeMetrics()
	} else if f.Const != nil {
		return f.Const.NodeMetrics()
	} else if f.Variable != nil {
		return f.Variable.NodeMetrics()
	}
	return 0, 0
}

func (f *NWFFormula) Copy() *NWFFormula {
	if f.And != nil {
		return &NWFFormula{
			And: &NWFAnd{
				Arg1:       f.And.Arg1,
				Arg2:       f.And.Arg2,
				IsNeg:      f.And.IsNeg,
				Depth:      f.And.Depth,
				Complexity: f.And.Complexity,
			},
		}
	} else if f.Or != nil {
		return &NWFFormula{
			Or: &NWFOr{
				Arg1:       f.Or.Arg1,
				Arg2:       f.Or.Arg2,
				IsNeg:      f.Or.IsNeg,
				Depth:      f.Or.Depth,
				Complexity: f.Or.Complexity,
			},
		}
	} else if f.Variable != nil {
		return &NWFFormula{
			Variable: &NWFVar{
				ID: f.Variable.ID,
			},
		}
	} else if f.Const != nil {
		return &NWFFormula{
			Const: &NWFConst{
				Value: f.Const.Value,
			},
		}
	}
	return nil
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
	return "???"
}

func (f *NWFFormula) Serialize() string {
	if f.And != nil {
		return f.And.Serialize()
	} else if f.Or != nil {
		return f.Or.Serialize()
	} else if f.Const != nil {
		return f.Const.Serialize()
	} else if f.Variable != nil {
		return f.Variable.Serialize()
	}
	return "???"
}

func (f *NWFFormula) recNormalizeVars(vars *SATVariableMapping, newVars *SATVariableMapping, notified *map[*NWFVar]struct{}) {
	if f.And != nil {
		f.And.recNormalizeVars(vars, newVars, notified)
	} else if f.Or != nil {
		f.Or.recNormalizeVars(vars, newVars, notified)
	} else if f.Const != nil {
		f.Const.recNormalizeVars(vars, newVars, notified)
	} else if f.Variable != nil {
		f.Variable.recNormalizeVars(vars, newVars, notified)
	}
}


func (f *NWFFormula) AST(vars *SATVariableMapping) *Formula {
	if f.Variable != nil {
		return f.Variable.AST(vars)
	} else if f.And != nil {
		return f.And.AST(vars)
	} else if f.Or != nil {
		return f.Or.AST(vars)
	} else if f.Const != nil {
		return f.Const.AST(vars)
	}
	return nil
}

func (f *NWFFormula) Negate() {
	if f.Variable != nil {
		f.Variable.ID = -f.Variable.ID
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
	notified := map[*NWFVar]struct{}{}
	f.recNormalizeVars(vars, newVars, &notified)
	return nil, newVars, len(newVars.names)
}

func (f *NWFFormula) Measure() *SATFormulaStatistics {
	d, c := f.NodeMetrics()
	return &SATFormulaStatistics{
		variableCount:    0,
		clauseCount:      0,
		clauseLenSum:     0,
		clauseDepth:      int64(d),
		clauseComplexity: int64(c),
	}
}
