package core

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/preprocessor"
	"github.com/go-sat-solver/sat_solver/preprocessor/nwf_converter"
	"github.com/go-sat-solver/sat_solver/solver"
)

func RunSATSolver(formula *sat_solver.Entry) (error, bool) {
	err, nwfFormula := nwf_converter.ConvertToNWF(formula)
	if err != nil {
		if _, ok := err.(*sat_solver.UnsatError); ok {
			return nil, false
		}
		return err, false
	}
	if nwfFormula.IsQuickUNSAT() {
		return nil, false
	}
	fmt.Printf("Converted to NWF:\n [%s]\n", nwfFormula.String())

	optimizedAST := nwfFormula.AST()
	err, satFormula := preprocessor.PreprocessAST(optimizedAST)
	if err != nil {
		if _, ok := err.(*sat_solver.UnsatError); ok {
			return nil, false
		}
		return err, false
	}
	if satFormula.IsQuickUNSAT() {
		return nil, false
	}

	err, result := solver.Solve(satFormula)
	if err != nil {
		if _, ok := err.(*sat_solver.UnsatError); ok {
			return nil, false
		}
		return err, false
	}
	return nil, result
}