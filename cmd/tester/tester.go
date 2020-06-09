package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/styczynski/go-sat-solver/sat_solver"
	"github.com/styczynski/go-sat-solver/sat_solver/core"
)

var (
	cli struct {
		Directory string `arg:"" type:"path" required:"" help:"Directory with tests."`
	}
)

var TESTS_REGEX = `.*test([0-9]+)\.txt`

func main() {
	ctx := kong.Parse(&cli)
	r, err := regexp.Compile(TESTS_REGEX)
	ctx.FatalIfErrorf(err)
	err = filepath.Walk(cli.Directory, func(path string, info os.FileInfo, err error) error {
		matched := r.MatchString(path)
		if matched {
			testNoPostfix := r.FindStringSubmatch(path)[1]
			dir := filepath.Dir(path)
			expectedTestResultBytes, err := ioutil.ReadFile(filepath.Join(dir, fmt.Sprintf("result%s.txt", testNoPostfix)))
			if err != nil {
				return err
			}
			expectedTestResultStr := strings.ReplaceAll(strings.ReplaceAll(string(expectedTestResultBytes), "\n", ""), " ", "")
			expectedTestResult, err := strconv.Atoi(expectedTestResultStr)
			if err != nil {
				return err
			}

			fmt.Printf("Execute test %s: ", testNoPostfix)

			err, result := core.RunSATSolverOnFilePath(path, sat_solver.DefaultSATContext())
			if err != nil {
				fmt.Printf("______________RESULT____________:\n  Test: %s, Err: %s\n___________________", testNoPostfix, err.Error())
				return nil
			}

			if result.IsUndefined() || result.ToInt() != expectedTestResult {
				fmt.Printf(" ERR\n______________RESULT____________:\n  Test: %s, Got: %d, Expected: %d\n___________________", testNoPostfix, result.ToInt(), expectedTestResult)
				panic(fmt.Sprintf("WRONG ANSWER ON TEST %s", testNoPostfix))
			} else {
				fmt.Printf(" OK\n")
			}
		}
		return nil
	})
	ctx.FatalIfErrorf(err)
}