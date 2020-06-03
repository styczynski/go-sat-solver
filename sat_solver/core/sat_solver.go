package core

import (
	"os"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/parser"
	"github.com/go-sat-solver/sat_solver/preprocessor"
	"github.com/go-sat-solver/sat_solver/preprocessor/nwf_converter"
	"github.com/go-sat-solver/sat_solver/solver"
	"github.com/go-sat-solver/sat_solver/solver/cdcl_solver"
)

func RunSATSolverOnFilePath(filePath string, context *sat_solver.SATContext) (error, int) {
	r, err := os.Open(filePath)
	if err != nil {
		return err, 0
	}
	defer r.Close()

	err, ast := parser.ParseInputFormula(r)
	if err != nil {
		return err, 0
	}
	err, result := RunSATSolver(ast, context)
	if err != nil {
		return err, 0
	}
	if result {
		return nil, 1
	}
	return nil, 0
}

func RunSATSolver(formula *sat_solver.Entry, context *sat_solver.SATContext) (error, bool) {
	err, nwfFormula := nwf_converter.ConvertToNWF(formula, context)
	if err != nil {
		if _, ok := err.(*sat_solver.UnsatError); ok {
			return nil, false
		}
		return err, false
	}
	if nwfFormula.IsQuickUNSAT() {
		return nil, false
	}

	optimizedAST := nwfFormula.AST()
	err, satFormula := preprocessor.PreprocessAST(optimizedAST, context)
	if err != nil {
		if _, ok := err.(*sat_solver.UnsatError); ok {
			return nil, false
		}
		return err, false
	}
	if satFormula.IsQuickUNSAT() {
		return nil, false
	}

	err, result := solver.Solve(satFormula, cdcl_solver.NewCDCLSolver(), context)
	if err != nil {
		if _, ok := err.(*sat_solver.UnsatError); ok {
			return nil, false
		}
		return err, false
	}
	return nil, result
}