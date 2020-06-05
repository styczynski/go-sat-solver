package core

import (
	"os"

	"github.com/go-sat-solver/sat_solver"
	solver2 "github.com/go-sat-solver/sat_solver/loaders"
	"github.com/go-sat-solver/sat_solver/preprocessor"
	"github.com/go-sat-solver/sat_solver/preprocessor/nwf_converter"
	"github.com/go-sat-solver/sat_solver/solver"

	_ "github.com/go-sat-solver/sat_solver/solver/cdcl_solver"
	_ "github.com/go-sat-solver/sat_solver/solver/naive_solver"

	_ "github.com/go-sat-solver/sat_solver/loaders/haskell"
	_ "github.com/go-sat-solver/sat_solver/loaders/dimacs_cnf"
)

func RunSATSolverOnFilePath(filePath string, context *sat_solver.SATContext) (error, solver.SolverResult) {
	r, err := os.Open(filePath)
	if err != nil {
		return err, solver.EmptySolverResult{}
	}
	defer r.Close()

	err, loadedFormula := solver2.LoadFormula(context.GetConfiguration().LoaderName, r, context)
	if err != nil {
		return err, solver.EmptySolverResult{}
	}
	err, result := RunSATSolver(loadedFormula, context)
	if err != nil {
		return err, solver.EmptySolverResult{}
	}
	return nil, result
}

type ConvertableToAST interface {
	AST() *sat_solver.Formula
}

func RunSATSolver(formula solver2.LoadedFormula, context *sat_solver.SATContext) (error, solver.SolverResult) {
	context.Trace("init", "SAT solver inited with the following configuration:\n%s", context.DescribeConfiguration())

	var globalResult solver.SolverResult
	err, executionContext := context.StartProcessing("SolveInstance", "")
	if err != nil {
		return err, solver.EmptySolverResult{}
	}

	var optimizedAST solver2.LoadedFormula = formula
	var satFormula *sat_solver.SATFormula = nil

	if !optimizedAST.CanBeConvertedToFormula() {
		if context.GetConfiguration().EnableASTOptimization || !context.GetConfiguration().EnableCNFConversion {
			err, nwfFormula := nwf_converter.ConvertToNWF(optimizedAST.ConvertToAST(), executionContext)
			if err != nil {
				if _, ok := err.(*sat_solver.UnsatError); ok {
					err = executionContext.EndProcessing(globalResult)
					if err != nil {
						panic(err)
					}
					return nil, solver.SolverQuickUnsatResult{}
				}
				return err, solver.EmptySolverResult{}
			}
			if nwfFormula.IsQuickUNSAT() {
				err = executionContext.EndProcessing(globalResult)
				if err != nil {
					panic(err)
				}
				return nil, solver.SolverQuickUnsatResult{}
			}
			optimizedAST = nwfFormula
			satFormula = nwfFormula
		}
	}

	if context.GetConfiguration().EnableCNFConversion {
		err, satFormula = preprocessor.PreprocessAST(optimizedAST, executionContext)
		if err != nil {
			if _, ok := err.(*sat_solver.UnsatError); ok {
				err = executionContext.EndProcessing(globalResult)
				if err != nil {
					panic(err)
				}
				return nil, solver.SolverQuickUnsatResult{}
			}
			return err, solver.EmptySolverResult{}
		}
		if satFormula.IsQuickUNSAT() {
			globalResult = solver.SolverQuickUnsatResult{}
			err = executionContext.EndProcessing(globalResult)
			if err != nil {
				panic(err)
			}
			return nil, solver.SolverQuickUnsatResult{}
		}
	}

	err, result := solver.Solve(satFormula, context.GetConfiguration().SolverName, executionContext)
	if err != nil {
		if _, ok := err.(*sat_solver.UnsatError); ok {
			err = executionContext.EndProcessing(globalResult)
			if err != nil {
				panic(err)
			}
			return nil, solver.SolverQuickUnsatResult{}
		}
		return err, solver.EmptySolverResult{}
	}

	globalResult = result
	err = executionContext.EndProcessing(globalResult)
	if err != nil {
		return err, result
	}
	return nil, result
}