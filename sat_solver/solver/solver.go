package solver

import "github.com/go-sat-solver/sat_solver"

type Solver interface {
	Solve(formula *sat_solver.SATFormula) bool
}

func Solve(formula *sat_solver.SATFormula) error {


	return nil
}