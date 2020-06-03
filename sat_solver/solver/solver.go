package solver

import (
	"github.com/go-sat-solver/sat_solver"
)

type Solver interface {
	Solve(formula *sat_solver.SATFormula, context *sat_solver.SATContext) (error, SolverResult)
}

type SolverResult interface {
	ToBool() bool
	String() string
	Brief() string
}

func Solve(formula *sat_solver.SATFormula, solver Solver, context *sat_solver.SATContext) (error, bool) {
	err, solvingContext := context.StartProcessing("Solve formula (CDCL solver)", "")
	if err != nil {
		return err, false
	}
	err, result := solver.Solve(formula, context)
	if err != nil {
		return err, false
	}
	err = solvingContext.EndProcessing(result)
	if err != nil {
		return err, false
	}

	return nil, result.ToBool()
}