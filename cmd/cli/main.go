package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/go-sat-solver/sat_solver/core"
	"github.com/go-sat-solver/sat_solver/parser"
)

var (
	cli struct {
		Files []string `arg:"" type:"existingfile" required:"" help:"GraphQL schema files to parse."`
	}
)

func main() {
	ctx := kong.Parse(&cli)
	for _, file := range cli.Files {
		fmt.Printf("open: %s\n", file)
		r, err := os.Open(file)
		ctx.FatalIfErrorf(err)
		err, ast := parser.ParseInputFormula(r)
		ctx.FatalIfErrorf(err)
		err, result := core.RunSATSolver(ast)
		fmt.Printf("Result is:\n\n %t\n\n", result)
		ctx.FatalIfErrorf(err)
		err = r.Close()
		ctx.FatalIfErrorf(err)
	}
}