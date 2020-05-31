package test_utils

import (
	"fmt"

	"github.com/mitchellh/go-sat"
	"github.com/mitchellh/go-sat/cnf"

	"github.com/go-sat-solver/sat_solver"
)

func AssertSatResult(formula *sat_solver.SATFormula, expectedResult bool) {
	fmt.Printf("Test SAT assertion.\n")
	r := TestSolveFormula(formula)
	if r != expectedResult {
		panic(fmt.Sprintf("assertSatResult: Expected %t got %t.", expectedResult, r))
	}
}

func TestSolveFormula(formula *sat_solver.SATFormula) bool {
	variableRemap := map[int64]int{}
	freeID := 1
	if f, ok := formula.Formula().(*sat_solver.CNFFormula); ok {
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
					}
				} else {
					if k, ok := variableRemap[v]; ok {
						vars[i][j] = k
					} else {
						variableRemap[v] = freeID
						vars[i][j] = freeID
						freeID++
					}
				}
			}
		}
		formula := cnf.NewFormulaFromInts(vars)
		// Create a solver and add the formulas to solve
		s := sat.New()
		s.AddFormula(formula)

		// For low level details on how a solution came to be:
		// s.Trace = true
		// s.Tracer = log.New(os.Stderr, "", log.LstdFlags)

		// Solve it!
		satResult := s.Solve()
		return satResult
	} else {
		panic("TestSolveFormula: Non-CNF formulas are not supported yet")
	}
}