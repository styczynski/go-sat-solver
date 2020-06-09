package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/core"
	"github.com/go-sat-solver/sat_solver/solver"
)

var (
	cli struct {
		Files                  []string `arg:"" type:"existingfile" required:"" help:"Input files with formulas."`
		Debug                  bool     `help:"Display debugging information" short:"d"`
		Trace                  bool     `help:"Trace solver execution" short:"t"`
		PrintFoundAssignment   bool     `help:"Print variables assignment on SAT result" short:"a"`
		SolverName             string   `help:"Specify solver to use" short:"s" default:"cdcl"`
		LoaderName             string   `help:"Specify format of the loaded input" short:"f" default:"haskell"`
		ExpectedResult         int      `help:"Specify expected result. This is useful when debugging the solver. Terribly slows down computation." enum:"-1,0,1" default:"-1"`
		DisableCNFConversion   bool     `help:"Disable conversion to CNF." default:"false"`
		EnableASTOptimization  bool     `help:"Enable input AST mangling." default:"false"`
		EnableCNFOptimizations bool     `help:"Enable CNF preprocessing" default:"false"`
	}
)

func main() {
	ctx := kong.Parse(&cli)
	for _, file := range cli.Files {
		var expectedResult *bool = nil
		if cli.ExpectedResult == 0 || cli.ExpectedResult == 1 {
			expectedResultVal := cli.ExpectedResult == 1
			expectedResult = &expectedResultVal
		}
		err, result := core.RunSATSolverOnFilePath(file, sat_solver.NewSATContext(sat_solver.SATConfiguration{
			InputFile:              file,
			ExpectedResult:         expectedResult,
			EnableSelfVerification: expectedResult != nil,
			EnableEventCollector:   cli.Trace || cli.Debug,
			EnableSolverTracing:    cli.Trace,
			EnableCNFConversion:    !cli.DisableCNFConversion,
			EnableASTOptimization:  cli.EnableASTOptimization,
			EnableCNFOptimizations: cli.EnableCNFOptimizations,
			SolverName:             cli.SolverName,
			LoaderName:             cli.LoaderName,
		}))
		ctx.FatalIfErrorf(err)
		if cli.PrintFoundAssignment {
			fmt.Printf("%s\n", solver.GetSolverResultSatisfyingAssignmentString(result))
		}
		fmt.Printf("%d\n", result.ToInt())
	}
}