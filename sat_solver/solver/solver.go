package solver

import (
	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/solver/naive_solver"
)

type Solver interface {
	Solve(formula *sat_solver.SATFormula, context *sat_solver.SATContext) bool
}

func Solve(formula *sat_solver.SATFormula, context *sat_solver.SATContext) (error, bool) {
	solver := naive_solver.NewNaiveSolver()
	err, result := solver.Solve(formula, context)
	if err != nil {
		return err, false
	}

	return nil, result
}