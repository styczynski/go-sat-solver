package preprocessor

import (
	"fmt"

	"github.com/go-sat-solver/sat_solver"
	"github.com/go-sat-solver/sat_solver/preprocessor/cnf_tseytins"
	"github.com/go-sat-solver/sat_solver/preprocessor/nwf_converter"
)

func PreprocessAST(formula sat_solver.Entry) (error, *sat_solver.SATFormula) {

	err, nwfFormula := nwf_converter.ConvertToNWF(formula)
	if err != nil {
		return err, nil
	}
	fmt.Printf("Converted to NWF:\n %s\n", nwfFormula.String())
	return nil, nwfFormula

	err, satFormula := cnf_tseytins.ConvertToCNFTseytins(formula) //cnf_naive.ConvertToCNFNaive(formula)
	if err != nil {
		return err, nil
	}

	fmt.Printf("Preprocessed formula:\n %s\n", satFormula.String())

	err, simplFormula := Optimize(satFormula)
	if err != nil {
		return err, nil
	}

	fmt.Printf("Simplified formula:\n %s\n", simplFormula.Brief())

	return nil, satFormula
}