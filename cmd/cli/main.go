package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/go-sat-solver/sat_solver/parser"
	"github.com/go-sat-solver/sat_solver/preprocessor"
	"github.com/go-sat-solver/sat_solver/solver"
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
		err, sat := preprocessor.PreprocessAST(*ast)
		ctx.FatalIfErrorf(err)
		err = solver.Solve(sat)
		ctx.FatalIfErrorf(err)
		err = r.Close()
		ctx.FatalIfErrorf(err)
	}
}