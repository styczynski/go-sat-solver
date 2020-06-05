package core

import (
	"os"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/parser"
	"github.com/go-sat-solver/sat_solver/preprocessor"
	"github.com/go-sat-solver/sat_solver/preprocessor/nwf_converter"
	"github.com/go-sat-solver/sat_solver/solver"

	_ "github.com/go-sat-solver/sat_solver/solver/cdcl_solver"
	_ "github.com/go-sat-solver/sat_solver/solver/naive_solver"
)

func RunSATSolverOnFilePath(filePath string, context *sat_solver.SATContext) (error, solver.SolverResult) {
	r, err := os.Open(filePath)
	if err != nil {
		return err, solver.EmptySolverResult{}
	}
	defer r.Close()

	err, ast := parser.ParseInputFormula(r)
	if err != nil {
		return err, solver.EmptySolverResult{}
	}
	err, result := RunSATSolver(ast, context)
	if err != nil {
		return err, solver.EmptySolverResult{}
	}
	return nil, result
}

type ConvertableToAST interface {
	AST() *sat_solver.Formula
}

func RunSATSolver(formula *sat_solver.Entry, context *sat_solver.SATContext) (error, solver.SolverResult) {
	context.Trace("init", "SAT solver inited with the following configuration:\n%s", context.DescribeConfiguration())

	var globalResult solver.SolverResult
	err, executionContext := context.StartProcessing("SolveInstance", "")
	if err != nil {
		return err, solver.EmptySolverResult{}
	}
	defer func() {
		err = executionContext.EndProcessing(globalResult)
		if err != nil {
			panic(err)
		}
	}()

	var optimizedAST ConvertableToAST = formula
	var satFormula *sat_solver.SATFormula = nil

	if context.GetConfiguration().EnableASTOptimization || !context.GetConfiguration().EnableCNFConversion {
		err, nwfFormula := nwf_converter.ConvertToNWF(formula, executionContext)
		if err != nil {
			if _, ok := err.(*sat_solver.UnsatError); ok {
				return nil, solver.EmptySolverResult{}
			}
			return err, solver.EmptySolverResult{}
		}
		if nwfFormula.IsQuickUNSAT() {
			return nil, solver.EmptySolverResult{}
		}
		optimizedAST = nwfFormula
		satFormula = nwfFormula
	}

	if context.GetConfiguration().EnableCNFConversion {
		err, satFormula = preprocessor.PreprocessAST(optimizedAST.AST(), executionContext)
		if err != nil {
			if _, ok := err.(*sat_solver.UnsatError); ok {
				return nil, solver.EmptySolverResult{}
			}
			return err, solver.EmptySolverResult{}
		}
		if satFormula.IsQuickUNSAT() {
			globalResult = solver.SolverQuickUnsatResult{}
			return nil, solver.EmptySolverResult{}
		}
	}

	err, result := solver.Solve(satFormula, context.GetConfiguration().SolverName, executionContext)
	if err != nil {
		if _, ok := err.(*sat_solver.UnsatError); ok {
			return nil, solver.EmptySolverResult{}
		}
		return err, solver.EmptySolverResult{}
	}

	globalResult = result
	return nil, result
}