# [Go SAT solver](http://styczynski.in/go-sat-solver/)

This isn't the most effective implementation. This code was written as an assignment for a Logic Course at MIM UW (2020).
The implementation should be fairly bug-less.

[See online (WASM) version](http://styczynski.in/go-sat-solver/)

## Quickstart

You have to [install go](https://golang.org/doc/install) on your machine 

```
   $ GO111MODULE=on go get -u github.com/styczynski/go-sat-solver
   $ go-sat-solver input.txt # Run the executable
```

If the shell cannot find the executable you may want to do `export PATH=$PATH:/usr/local/go/bin`.

## Build or Install using make

* `make install` - adds the binary to `$GOPATH/bin`
* `make build` - builds the binary

The `Makefile` builds the binary and adds the short git commit id and the version to the final build.

## Usage

The solver can the executed using `go-sat-solver [input files]` command.
The default input format is `haskell`-like ADT syntax:
```
    X := And (X) (X) | Or (X) (X) | Iff (X) (X) | Implies (X) (X) | Not (X) | Var "string" | T | F
```

You may want to load other types of files for example DIMACS CNF:
```bash
    $ go-sat-solver -f cnf input.cnf
```

Or use other solver than the default one (currently `cnf` and `naive` options are supported):
```bash
    $ go-sat-solver -s naive input.txt
```

## About the solver itself

The solver was firstly a DPLL-style solver but further improvements led to CDCL-like solver. 
I was using Minisat source code as a reference.
The solver supports the following features:
* Selection partially based on [Adaptive VSIDS](https://arxiv.org/pdf/1506.08905.pdf)
* [TWL](http://people.mpi-inf.mpg.de/~mfleury/sat_twl.pdf)
* Clause learning

The learned clauses are not optimized based on adaptive VSIDS, but this feature is planned in the future.

This solver is suitable for any serious application, but you shall consider using other Go solvers, or native C/C++ solvers for a better performance.

## Web interface

Solver has its web [interface available](http://styczynski.in/go-sat-solver/) (this was done using compilation of Go to WASM).

You can compile the frontend in a following way:
```bash
    $ make wasm
    $ make build-web
```

## Running tests

You can run testing scripts that examines the solver on inputs from `/tests/` directory using the following command:
```bash
    $ go run ./cmd/tester/tester.go ./tests
```