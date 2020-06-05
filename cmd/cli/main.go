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
		Files                  []string `arg:"" type:"existingfile" required:"" help:"GraphQL schema files to parse."`
		Debug                  bool
		Trace                  bool
		PrintFoundAssignment   bool
		SolverName             string `default:"cdcl"`
		ExpectedResult         int    `enum:"-1,0,1" default:"-1"`
		DisableCNFConversion    bool   `default:"false"`
		DisableASTOptimization  bool   `default:"false"`
		DisableCNFOptimizations bool   `default:"false"`
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
			EnableASTOptimization:  !cli.DisableASTOptimization,
			EnableCNFOptimizations: !cli.DisableCNFOptimizations,
			SolverName:             cli.SolverName,
		}))
		ctx.FatalIfErrorf(err)
		if cli.PrintFoundAssignment {
			fmt.Printf("%s\n", solver.GetSolverResultSatisfyingAssignmentString(result))
		}
		fmt.Printf("%d\n", result.ToInt())
	}
}