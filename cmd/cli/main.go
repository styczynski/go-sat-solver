package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/core"
)

var (
	cli struct {
		Files []string `arg:"" type:"existingfile" required:"" help:"GraphQL schema files to parse."`
	}
)

func main() {
	ctx := kong.Parse(&cli)
	for _, file := range cli.Files {
		err, result := core.RunSATSolverOnFilePath(file, sat_solver.NewSATContextAssert(file, false))
		ctx.FatalIfErrorf(err)
		fmt.Printf("Result:\n  %d\n", result)
	}
}