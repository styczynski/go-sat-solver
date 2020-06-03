package naive_solver

import (
	"fmt"
	"math"

	"github.com/go-sat-solver/sat_solver"
)

type NaiveSolver struct {}

func NewNaiveSolver() *NaiveSolver {
	return &NaiveSolver{}
}

func (solver *NaiveSolver) Solve(formula *sat_solver.SATFormula, context *sat_solver.SATContext) (error, bool) {
	fmt.Printf("Naive solver input:\n %s\n", formula.String())

	err, vars := formula.Normalize()
	if err != nil {
		return err, false
	}

	varCount := int64(len(vars))
	if varCount > 63 {
		return fmt.Errorf("Too many variables for naive solver (%d)", varCount), false
	}

	values := int64(0)
	fmt.Printf("Vars count: %d\n", varCount)
	iterCount := int64(math.Exp2(float64(varCount)))
	for i := int64(0); i < iterCount; i++ {
		for j := int64(0); j < varCount; j++ {
			vars[j] = (int64(1) >> j) & values != 0
		}
		if formula.Evaluate(vars) {
			fmt.Printf("Naive solution:\n")
			for k, v := range vars {
				fmt.Printf("  %s = %t\n", formula.Variables().Reverse(sat_solver.CNFLiteral(k+1)), v)
			}
			return nil, true
		}
		values++
	}

	return nil, false
}