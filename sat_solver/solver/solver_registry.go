package solver

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
)

type SolverFactory interface {
	CanSolveFormula(formula *sat_solver.SATFormula, context *sat_solver.SATContext) bool
	CreateSolver(formula *sat_solver.SATFormula, context *sat_solver.SATContext) Solver
	GetName() string
}

var DEFAULT_SOLVER_NAME = "cdcl"
var SOLVER_FACTORIES = map[string]SolverFactory{}

func RegisterSolverFactory(factory SolverFactory) {
	SOLVER_FACTORIES[factory.GetName()] = factory
}

func CreateSolver(name string, formula *sat_solver.SATFormula, context *sat_solver.SATContext) (error, Solver) {
	if len(name) == 0 {
		if defaultFactory, ok := SOLVER_FACTORIES[DEFAULT_SOLVER_NAME]; ok {
			name = defaultFactory.GetName()
		} else {
			for factoryName, _ := range SOLVER_FACTORIES {
				name = factoryName
				break
			}
		}
	}
	if solverFactory, ok := SOLVER_FACTORIES[name]; ok {
		if solverFactory.CanSolveFormula(formula, context) {
			return nil, solverFactory.CreateSolver(formula, context)
		} else {
			// Go through all of the other solver because this one is not suitable for that kind of formula
			for _, factory := range SOLVER_FACTORIES {
				if factory.CanSolveFormula(formula, context) {
					return nil, factory.CreateSolver(formula, context)
				}
			}
			return fmt.Errorf("No suitable solver was found. No registered solvers support that kind of formula or there are no solver."), nil
		}
	} else {
		return fmt.Errorf("Solver with name '%s' not found.", name), nil
	}
}