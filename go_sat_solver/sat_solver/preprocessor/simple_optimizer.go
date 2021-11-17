package preprocessor

import (
	"fmt"
	"math"
	"strings"

	"github.com/styczynski/go-sat-solver/sat_solver"
)

func hashVarID(varID sat_solver.CNFLiteral) int64 {
	if varID < 0 {
		return (1 >> uint64(-varID) % 63)
	}
	return (1 >> uint64(varID) % 63)
}

type UnsatReasonUP struct {}

func NewUnsatReasonUP() *UnsatReasonUP {
	return &UnsatReasonUP{}
}

func (reason *UnsatReasonUP) Describe() string {
	return fmt.Sprintf("Unit propagation cannot propagate variable with T or F.")
}

type UnsatReasonStrengthening struct {
	clause string
	varName string
}

func NewUnsatReasonStrengthening(clause *Clause, varID sat_solver.CNFLiteral, opt *SimpleOptimizer) *UnsatReasonStrengthening {
	return &UnsatReasonStrengthening{
		clause: clause.String(opt),
		varName: opt.vars.Reverse(varID),
	}
}

func (reason *UnsatReasonStrengthening) Describe() string {
	return fmt.Sprintf("Strengthening clause %s by variable %s produced an empty clause.", reason.clause, reason.varName)
}

type Clause struct {
	hash int64
	vars map[sat_solver.CNFLiteral]struct{}
	isDeleted bool
}

func (c *Clause) negateClauseVar(varIDToNegate sat_solver.CNFLiteral) *Clause {
	newVars := map[sat_solver.CNFLiteral]struct{}{}
	for varID := range c.vars {
		if varID == varIDToNegate || varID == -varIDToNegate {
			newVars[-varID] = struct{}{}
		} else {
			newVars[varID] = struct{}{}
		}
	}
	return &Clause{
		hash:  c.hash,
		vars:  newVars,
	}
}

func (c *Clause) String(bve *SimpleOptimizer) string {
	strs := []string{}
	for v := range c.vars {
		strs = append(strs, bve.vars.Reverse(v))
	}
	return fmt.Sprintf("[%s]<%#v>", strings.Join(strs, " v "), c.isDeleted)
}

func (c *Clause) Rehash() {
	hashVal := int64(0)
	for v := range c.vars {
		hashVal = hashVal | hashVarID(v)
	}
	c.hash = hashVal
}

type SimpleOptimizer struct {
	clauses map[*Clause]struct{}
	occur map[sat_solver.CNFLiteral]map[*Clause]struct{}
	singular map[*Clause]struct{}

	added map[*Clause]struct{}
	touched map[sat_solver.CNFLiteral]struct{}
	strenghtened map[*Clause]struct{}

	visitedUnits map[sat_solver.CNFLiteral]struct{}

	vars *sat_solver.SATVariableMapping
	context *sat_solver.SATContext
}

func (opt *SimpleOptimizer) Formula() *sat_solver.SATFormula {
	newFormula := sat_solver.CNFFormula{
		Variables: make([]sat_solver.CNFClause, len(opt.clauses)),
	}
	i := 0
	for c := range opt.clauses {
		newClause := make(sat_solver.CNFClause, len(c.vars))
		j := 0
		for v := range c.vars {
			newClause[j] = v
			j++
		}
		newFormula.Variables[i] = newClause
		i++
	}
	return sat_solver.NewSATFormula(&newFormula, opt.vars, nil)
}

func (opt *SimpleOptimizer) ToSATFormula() *sat_solver.SATFormula {
	return opt.Formula()
}

func (opt *SimpleOptimizer) notEqual(clause *Clause, clause2 *Clause) bool {
	if clause == clause2 {
		return false
	}
	if clause.hash != clause2.hash {
		return true
	}
	if len(clause.vars) != len(clause2.vars) {
		return true
	}
	for v := range clause.vars {
		if _, ok := clause2.vars[v]; !ok {
			return true
		}
	}
	return false
}

func (opt *SimpleOptimizer) subset(clause *Clause, clause2 *Clause) bool {
	//fmt.Printf("  Is %s subset of %s?", clause.String(opt), clause2.String(opt))
	if (clause.hash & ^clause2.hash != 0) {
		return false
	}
	found := false
	for varA := range clause.vars {
		found = false
		for varB := range clause2.vars {
			if varA == varB {
				found = true
				break
			}
		}
		if !found {
			//fmt.Printf("  NO\n")
			return false
		}
	}
	//fmt.Printf("  YES\n")
	return true
}

func (opt *SimpleOptimizer) findSubsumed(clause *Clause) []*Clause {
	res := []*Clause{}
	pLit := sat_solver.CNFLiteral(0)
	pMin := int64(math.MaxInt64)
	for v := range clause.vars {
		occurLen := int64(len(opt.occur[v]))
		if occurLen < pMin {
			pMin = occurLen
			pLit = v
		}
	}
	if pLit == 0 {
		return res
	}
	for cPrim := range opt.occur[pLit] {
		if !cPrim.isDeleted {
			//fmt.Printf("Check %s <%p> and %s <%p>\n", clause.String(opt), clause, cPrim.String(opt), cPrim)
			if opt.notEqual(clause, cPrim) && len(clause.vars) <= len(cPrim.vars) && opt.subset(clause, cPrim) {
				res = append(res, cPrim)
			}
		}
	}
	return res
}

func (opt *SimpleOptimizer) getAddedClauseCandidates(added *map[*Clause]struct{}, positiveSearch bool) map[*Clause]struct{} {
	for clause := range *added {
		for v := range clause.vars {
			if (positiveSearch && v > 0) {
				res := map[*Clause]struct{}{}
				for c := range opt.occur[v] {
					res[c] = struct{}{}
				}
				return res
			} else if (!positiveSearch && v < 0) {
				res := map[*Clause]struct{}{}
				for c := range opt.occur[v] {
					res[c] = struct{}{}
				}
				return res
			}
		}
	}
	return map[*Clause]struct{}{}
}

// Remove varID from clause
func (opt *SimpleOptimizer) strenghten(clause *Clause, varID sat_solver.CNFLiteral) error {
	//fmt.Printf("Strengthen clause: %s by %s\n", clause.String(opt), opt.vars.Reverse(varID))

	len1 := len(clause.vars)

	// Remove var from the clause
	delete(clause.vars, varID)
	delete(opt.occur[varID], clause)
	clause.Rehash()

	len2 := len(clause.vars)

	if len1 >= 1 && len2 == 1 {
		opt.singular[clause] = struct{}{}
	}

	if len1 == 1 && len2 == 0 {
		// Reverse change
		clause.vars[varID] = struct{}{}
		return sat_solver.NewUnsatError(NewUnsatReasonStrengthening(clause, varID, opt))
	}

	opt.strenghtened[clause] = struct{}{}
	for v := range clause.vars {
		opt.touched[v] = struct{}{}
	}

	// WHAT?
	// Remove empty clause
	//if len(clause.vars) == 0 {
	//	opt.removeClause(clause)
	//}

	//fmt.Printf("Strengthen result: %s\n", clause.String(opt))
	opt.validateState()
	return nil
}

func (opt *SimpleOptimizer) removeClause(clause *Clause) {
	//fmt.Printf("Remove clause: %s\n", clause.String(opt))
	clause.isDeleted = true
	delete(opt.clauses, clause)
	if len(clause.vars) == 1 {
		delete(opt.singular, clause)
	}
	for v := range opt.occur {
		delete(opt.occur[v], clause)
	}

	for v := range clause.vars {
		opt.touched[v] = struct{}{}
	}

	opt.validateState()
}

// Remove any clause subsumed by the first argument
func (opt *SimpleOptimizer) subsume(clause *Clause) {
	clausesToRemove := opt.findSubsumed(clause)
	for _, c := range clausesToRemove {
		opt.removeClause(c)
	}
}

func (opt *SimpleOptimizer) validateState() {
	return
	err := opt.checkIfStateIsValid()
	if err != nil {
		panic(err)
	}
}

func (opt *SimpleOptimizer) checkIfStateIsValid() error {

	validOccur := map[sat_solver.CNFLiteral]map[*Clause]struct{}{}
	validSingular := map[*Clause]struct{}{}

	// Check clauses
	for c := range opt.clauses {
		if !c.isDeleted {
			for v := range c.vars {
				if _, ok := opt.occur[v][c]; !ok {
					return fmt.Errorf("Variable %s occurs in %s but is not present in occur set.", opt.vars.Reverse(v), c.String(opt))
				} else {
					if _, ok := validOccur[v]; !ok {
						validOccur[v] = map[*Clause]struct{}{}
					}
					validOccur[v][c] = struct{}{}
				}
			}
			if len(c.vars) == 1 {
				if _, ok := opt.singular[c]; !ok {
					return fmt.Errorf("Clause %s is a unit clause but is not present in singular set.", c.String(opt))
				} else {
					validSingular[c] = struct{}{}
				}
			}
		}
	}

	for v := range opt.occur {
		for c := range opt.occur[v] {
			if _, ok := validOccur[v][c]; !ok {
				return fmt.Errorf("Variable %s is not present in cluase %s but it's present in the occur set.", opt.vars.Reverse(v), c.String(opt))
			}
		}
	}

	for c := range opt.singular {
		if _, ok := validSingular[c]; !ok {
			return fmt.Errorf("Clause %s is not a unit clause but is present in the singular set.", c.String(opt))
		}
	}

	for c := range opt.clauses {
		if c.isDeleted {
			return fmt.Errorf("Found deleted clause on clause list: %s", c.String(opt))
		}
	}

	return nil
}

func (opt *SimpleOptimizer) OptimizeTrivialTautologies() {
	for opt.tryOptimizeTrivialTautologies() {}
}

func (opt *SimpleOptimizer) tryOptimizeTrivialTautologies() bool {
	detectedChange := false
	for c := range opt.clauses {
		for v := range c.vars {
			if _, ok := opt.occur[-v][c]; ok {
				// Clause contains both phi and -phi
				opt.removeClause(c)
				detectedChange = true
			}
		}
	}
	return detectedChange
}


func (opt *SimpleOptimizer) PerformUnitPropagation() error {
	for {
		err, cont := opt.tryPerformUnitPropagation()
		if err != nil {
			return err
		}
		if !cont {
			 break
		}
	}
	return nil
}

func (opt *SimpleOptimizer) tryPerformUnitPropagation() (error, bool) {

	varToRemove := sat_solver.CNFLiteral(0)
	for c := range opt.singular {
		if len(c.vars) == 1 && !c.isDeleted {
			for varID := range c.vars {
				if _, ok := opt.visitedUnits[varID]; !ok {
					varToRemove = varID
					break
				}
			}
			break
		}
	}

	// Try with non-singular clauses
	//if varToRemove == 0 {
	//	for varID, _ := range opt.occur {
	//		if _, ok := opt.visitedUnits[varID]; !ok {
	//			varToRemove = varID
	//			break
	//		}
	//	}
	//}

	if varToRemove == 0 {
		return nil, false
	}

	opt.visitedUnits[varToRemove] = struct{}{}

	willTriggerUnsat := false
	for c := range opt.occur[-varToRemove] {
		if len(c.vars) == 1 {
			// Removing varToRemove will result in empty clause
			// So we will try with negated variable
			willTriggerUnsat = true
			varToRemove = -varToRemove
			break
		}
	}

	if willTriggerUnsat {
		for c := range opt.occur[-varToRemove] {
			if len(c.vars) == 1 {
				// We get empty clause for both values for the variable
				// So we throw UNSAT
				return sat_solver.NewUnsatError(NewUnsatReasonUP()), false
			}
		}
	}

	for c := range opt.occur[-varToRemove] {
		err := opt.strenghten(c, -varToRemove)
		if err != nil {
			return sat_solver.WrapError(err, "When performing unit propagation for variable %s (removing negation)", opt.vars.Reverse(varToRemove)), false
		}
	}
	for c := range opt.occur[varToRemove] {
		opt.removeClause(c)
	}

	return nil, true
}

func (opt *SimpleOptimizer) maybeEliminate(varID sat_solver.CNFLiteral) {
	if len(opt.occur[varID]) > 10 || len(opt.occur[-varID]) > 10 {
		return // Heuristic cut-off
	}
	opt.maybeClauseDistribute(varID)
}

func (opt *SimpleOptimizer) propagateToplevel() {
	// TODO: Implement
}

func (opt *SimpleOptimizer) cleanup() {
	v := true
	for v { _, v = opt.tryPerformUnitPropagation() }
	for opt.blockedClauseElimination() {}
}

func (opt *SimpleOptimizer) simplify() error {

	opt.singular = map[*Clause]struct{}{}
	for clause := range opt.clauses {
		if len(clause.vars) == 1 {
			opt.singular[clause] = struct{}{}
		}
	}

	/*
	 * Set of variables
	 * A variable is added to this set if it occurs in a clause being added, removed, or strengthened. Initially all variables are "touched"
	 */
	opt.touched = map[sat_solver.CNFLiteral]struct{}{}
	for clause := range opt.clauses {
		for v := range clause.vars {
			opt.touched[v] = struct{}{}
		}
	}

	/*
	 * Set of clauses
	 * When a clause is added to the SAT problem (e.g. by variable elimination), it is also added to this set.
	 * Initially all clauses are considered "added".
	 */
	opt.added = map[*Clause]struct{}{}
	for clause := range opt.clauses {
		opt.added[clause] = struct{}{}
	}

	/*
	 * Set of clauses
	 * When a clause is strengthened (one literal is removed, either by self-subsumption or toplevel propagation)
	 * it is added to this set. Initially the set is empty.
	 */
	opt.strenghtened = map[*Clause]struct{}{}

	for {
		// Subsumption

		//fmt.Printf("Iterate added %d\n", len(opt.added))

		S0 := opt.getAddedClauseCandidates(&opt.added,  true)
		for {
			//fmt.Printf("Iterate strenghtened\n")

			S1 := opt.getAddedClauseCandidates(&opt.added, false)
			for a := range opt.added {
				S1[a] = struct{}{}
			}
			for s := range opt.strenghtened {
				S1[s] = struct{}{}
			}
			// Clear Added and Strengthened
			opt.added = map[*Clause]struct{}{}
			opt.strenghtened = map[*Clause]struct{}{}

			// Loop
			for c := range S1 {
				err := opt.selfSubsume(c)
				if err != nil {
					return err
				}
			}
			// May strengthen/remove clauses
			opt.propagateToplevel()

			if len(opt.strenghtened) == 0 {
				break
			}
		}

		//fmt.Printf("Subsuming S0\n")
		for c := range S0 {
			if !c.isDeleted {
				opt.subsume(c)
			}
		}

		// Variable elimination

		//fmt.Printf("Variable elimination loop\n")
		for {
			//fmt.Printf("Eliminate variables\n")
			S := opt.touched
			opt.touched = map[sat_solver.CNFLiteral]struct{}{}
			for x := range S {
				opt.maybeEliminate(x)
			}
			if len(opt.touched) == 0 {
				break
			}
		}
		opt.cleanup()

		if len(opt.added) == 0 {
			break
		}
	}

	return nil
}

func (opt *SimpleOptimizer) selfSubsume(clause *Clause) error {
	for v := range clause.vars {
		subsumedBy := opt.findSubsumed(clause.negateClauseVar(v))
		for _, cPrim := range subsumedBy {
			err := opt.strenghten(cPrim, -v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (opt *SimpleOptimizer) blockedClauseElimination() bool {

	changeDetected := false
	for v, clausesWithV := range opt.occur {
		// Check if clauseWithV is blocked
		for clauseWithV := range clausesWithV {
			isBlocked := true

			//debugTraceClause := []*Clause{}
			//debugTraceClauseVarID := []int64{}

			if len(opt.occur[-v]) > 0 {
				for clauseWithNegV := range opt.occur[-v] {
					// Check if clauseWithV v clauseWithNegV is tautology
					isTautology := false
					for q := range clauseWithV.vars {
						if _, ok := opt.occur[-q][clauseWithNegV]; ok && q != v {

							//debugTraceClause = append(debugTraceClause, clauseWithNegV)
							//debugTraceClauseVarID = append(debugTraceClauseVarID, q)

							isTautology = true
							break
						}
					}
					if !isTautology {
						isBlocked = false
						break
					}
				}
				if isBlocked {
					// We can remove blocked clause
					/*fmt.Printf("Clause %s is blocked on %s:\n", clauseWithV.String(opt), opt.vars.Reverse(v))
					for i, c := range debugTraceClause {
						fmt.Printf("   by %s (var %s)\n", c.String(opt), opt.vars.Reverse(debugTraceClauseVarID[i]))
					}*/
					opt.removeClause(clauseWithV)
					changeDetected = true
				}
			}
		}
	}
	return changeDetected
}

func (opt *SimpleOptimizer) Brief() string {
	return opt.Formula().Brief()
}

func (opt *SimpleOptimizer) String() string {
	return opt.Formula().String()
}

func Optimize(formula *sat_solver.SATFormula, context *sat_solver.SATContext) (error, *sat_solver.SATFormula) {
	formRepr := formula.Formula()
	if f, ok := formRepr.(*sat_solver.CNFFormula); ok {

		hashVal := int64(0)

		bve := SimpleOptimizer{
			clauses: map[*Clause]struct{}{},
			occur:   map[sat_solver.CNFLiteral]map[*Clause]struct{}{},
			vars:    formula.Variables(),
			visitedUnits: map[sat_solver.CNFLiteral]struct{}{},
		}

		for _, clause := range f.Variables {
			clauseVars := map[sat_solver.CNFLiteral]struct{}{}
			for _, v := range clause {
				clauseVars[v] = struct{}{}
			}

			c := &Clause{
				vars:      clauseVars,
				isDeleted: false,
			}
			hashVal = 0
			for _, v := range clause {
				hashVal = hashVal | hashVarID(v)
				if _, ok := bve.occur[v]; ok {
					bve.occur[v][c] = struct{}{}
				} else {
					bve.occur[v] = map[*Clause]struct{}{
						c: {},
					}
				}
				if _, ok := bve.occur[-v]; !ok {
					bve.occur[-v] = map[*Clause]struct{}{}
				}
			}

			c.hash = hashVal
			bve.clauses[c] = struct{}{}
		}


		err, newContext := context.StartProcessing("Run simple optimizer","")
		if err != nil {
			return err, nil
		}
		err = bve.simplify()
		if err != nil {
			if v, ok := err.(*sat_solver.UnsatError); ok {
				f := bve.Formula()
				return nil, sat_solver.NewSATFormulaShortcut(f.Formula(), f.Variables(), nil, v)
			}
		}
		err = newContext.EndProcessingFormula(&bve)
		if err != nil {
			return err, nil
		}

		err, newContext = context.StartProcessing("Run simple postprocess","")
		if err != nil {
			return err, nil
		}
		bve.RemoveDanglingVariables()
		err = newContext.EndProcessingFormula(&bve)
		if err != nil {
			return err, nil
		}


		return nil, bve.Formula()
	}

	return nil, formula
}