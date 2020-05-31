package solver

import (
	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/solver/naive_solver"
)

type Solver interface {
	Solve(formula *sat_solver.SATFormula) bool
}

func Solve(formula *sat_solver.SATFormula) (error, bool) {
	solver := naive_solver.NewNaiveSolver()
	err, result := solver.Solve(formula)
	if err != nil {
		return err, false
	}

	return nil, result
}