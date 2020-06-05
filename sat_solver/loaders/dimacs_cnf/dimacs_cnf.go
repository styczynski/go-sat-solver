package dimacs_cnf

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-sat-solver/sat_solver"
	solver "github.com/go-sat-solver/sat_solver/loaders"
)

type CNFLoaderFactory struct {}

type CNFLoader struct {}

func (hlf *CNFLoaderFactory) CreateLoader(context *sat_solver.SATContext) solver.Loader {
	return CNFLoader{}
}

func (hlf *CNFLoaderFactory) GetName() string {
	return "cnf"
}

func (loader CNFLoader) Load(inputFormula io.Reader, context *sat_solver.SATContext) (error, solver.LoadedFormula) {
	vars := sat_solver.NewSATVariableMapping()
	cnf := &sat_solver.CNFFormula{
		Variables: []sat_solver.CNFClause{},
	}

	scanner := bufio.NewScanner(inputFormula)
	varCount := 0
	clauseCount := 0
	firstLineOk := false
	clauseLineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			if line[0] == 'c' {
				// Ignore
			} else {
				if !firstLineOk {
					_, err := fmt.Sscanf(line, "p cnf %d %d\n", &varCount, &clauseCount)
					if err != nil {
						return err, nil
					}
					firstLineOk = true
					cnf.Variables = make([]sat_solver.CNFClause, clauseCount)
				} else {
					varTokens := strings.Split(line, " ")
					if len(varTokens) == 0 {
						return fmt.Errorf("Clause line does not contain any numbers."), nil
					}
					if varTokens[len(varTokens)-1] != "0" {
						return fmt.Errorf("Clause line does not end with 0."), nil
					}
					clause := make(sat_solver.CNFClause, len(varTokens)-1)
					for i, token := range varTokens[:1] {
						literal, err := strconv.Atoi(token)
						if err != nil {
							return err, nil
						}
						varID := literal
						if varID < 0 {
							varID = -varID
						}
						varName := fmt.Sprintf("%d", varID)
						newID := vars.Get(varName)
						if literal < 0 {
							newID = -newID
						}
						clause[i] = newID
					}
					cnf.Variables[clauseLineNo] = clause

					clauseLineNo++
					if clauseLineNo >= clauseCount {
						break
					}
				}
			}
		}
	}

	return nil, sat_solver.NewSATFormula(cnf, vars, nil)
}

func init() {
	solver.RegisterLoaderFactory(&CNFLoaderFactory{})
}