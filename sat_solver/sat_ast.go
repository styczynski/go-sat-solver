package sat_solver

import "strings"

type SATFormula struct {
	formula *CNFFormula
	vars *SATVariableMapping
}

func NewSATFormula(formula *CNFFormula, vars *SATVariableMapping) *SATFormula {
	return &SATFormula{
		formula: formula,
		vars:    vars,
	}
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func (f *SATFormula) String() string {
	result := make([]string, len(f.formula.Variables))
	for j, clause := range f.formula.Variables {
		partialResult := make([]string, len(clause))
		for i, v := range clause {
			if (v == 1) {
				partialResult[i] = "True"
			} else if (v == -1) {
				partialResult[i] = "False"
			} else if (v > 0) {
				partialResult[i] = trimQuotes(f.vars.reverse[v])
			} else {
				partialResult[i] = "-" + trimQuotes(f.vars.reverse[-v])
			}
		}
		result[j] = "(" + strings.Join(partialResult, " v ") + ")"
	}
	return strings.Join(result, "^")
}

type CNFFormula struct {
	// -1 is false
	// 1 is true
	// x > 1 is a variable with unique id = x
	// x < 1 is a variable with unique id = -x
	Variables [][]int
}

func (f *CNFFormula) AndWith(e *CNFFormula) {
	f.Variables = append(f.Variables, e.Variables...)
}

func (f *CNFFormula) MulWith(e *CNFFormula) {
	lenF := len(f.Variables)
	lenE := len(e.Variables)
	newVariables := make([][]int, lenF * lenE)
	for indexF, clauseF := range f.Variables {
		for indexE, clauseE := range e.Variables {
			newVariables[indexF*lenE + indexE] = append(clauseF, clauseE...)
		}
	}
	f.Variables = newVariables
}

type SATVariableMapping struct {
	names map[string]int
	reverse map[int]string
	uniqueID int
}

func NewSATVariableMapping() *SATVariableMapping {
	return &SATVariableMapping{
		names:    map[string]int{},
		reverse:  map[int]string{},
		uniqueID: 2,
	}
}

func (vars *SATVariableMapping) Get(name string) int {
	if id, ok := vars.names[name]; ok {
		return id
	}
	newID := vars.uniqueID
	vars.uniqueID++
	vars.names[name] = newID
	vars.reverse[newID] = name
	return newID
}