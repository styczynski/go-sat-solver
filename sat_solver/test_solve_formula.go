package sat_solver

import (
	"fmt"

	"github.com/mitchellh/go-sat"
	"github.com/mitchellh/go-sat/cnf"
)

func AssertSatResult(formula *SATFormula, expectedResult bool) {
	r, assgn := TestSolveFormula(formula)
	if r != expectedResult {
		for name, val := range assgn {
			fmt.Printf("  | %s => %t\n", name, val)
		}
		panic(fmt.Sprintf("assertSatResult: Expected %t got %t.", expectedResult, r))
	}
}

func TestSolveFormula(formula *SATFormula) (bool, map[string]bool) {
	variableRemap := map[int64]int{}
	variableNames := map[int]string{}
	freeID := 1
	if f, ok := formula.Formula().(*CNFFormula); ok {
		vars := make([][]int, len(f.Variables))
		for i, clause := range f.Variables {
			vars[i] = make([]int, len(clause))
			for j, v := range clause {
				if v == 1 || v == -1 || v == 0 {
					panic("TestSolveFormula: Formula cannot contain T/F")
				}
				if v < 0 {
					if k, ok := variableRemap[-v]; ok {
						vars[i][j] = -k
					} else {
						variableRemap[-v] = freeID
						vars[i][j] = -freeID
						freeID++
						if formula.Variables().IsFounderVariable(-v) {
							variableNames[freeID] = formula.Variables().Reverse(-v)
						}
					}
				} else {
					if k, ok := variableRemap[v]; ok {
						vars[i][j] = k
					} else {
						variableRemap[v] = freeID
						vars[i][j] = freeID
						freeID++
						if formula.Variables().IsFounderVariable(v) {
							variableNames[freeID] = formula.Variables().Reverse(v)
						}
					}
				}
			}
		}
		formula := cnf.NewFormulaFromInts(vars)
		// Create a solver and add the formulas to solve
		s := sat.New()
		s.AddFormula(formula)

		// Solve it!
		satResult := s.Solve()

		satAssgn := map[string]bool{}
		for vID, val := range s.Assignments() {
			if v, ok := variableNames[vID]; ok {
				satAssgn[v] = val
			}
		}

		return satResult, satAssgn
	} else {
		panic("TestSolveFormula: Non-CNF formulas are not supported yet")
	}
}