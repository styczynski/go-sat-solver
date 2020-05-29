package preprocessor

import "github.com/go-sat-solver/sat_solver"

func PreprocessAST(formula sat_solver.Entry) (error, sat_solver.SATFormula) {
	return nil, sat_solver.SATFormula{}
}