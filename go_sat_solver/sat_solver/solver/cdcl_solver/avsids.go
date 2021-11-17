package cdcl_solver

/**
 * This file provides utilities for managing AVSIDS (Adaptive Variable State Independent Decaying Sum).
 *
 * For more details please see the minisat C++ code with Adaptive ASIDS support on github (vsids/adaptvsids/ directory):
 *    https://github.com/JLiangWaterloo/vsids
 * And the paper about VSIDS and Adaptive VSIDS:
 *    https://arxiv.org/pdf/1506.08905.pdf
 *
 * The basic idea is to have a score for each literal.
 * After the solver learns a clause the scores for each literal in that clause are incremented by varInc.
 * After each conflict decay occurs and each literal from the formula has its score multiplied by an alpha factor,
 * where 0 < alpha < 1.
 * The selection algorithm chooses variables to decision using the scores and takes literal with the highest score.
 * The detailed explanation why this is desirable is contained within the linked scientific paper.
 * The more advanced idea is to capture some heuristics like LBD (Literal blocks distance) and use different
 * decay rates for variables based on that.
 */

import (
	"container/heap"
	"github.com/styczynski/go-sat-solver/sat_solver"
)

type AVSIDS struct {
	// LBD is Literal blocks distance a heuristic value used to control the decay of variables
	// Some research about LBD is done here: https://www.ijcai.org/Proceedings/09/Papers/074.pdf
	lbdSeen     map[int]struct{}
	lbdEmaDecay float64
	lbdEma      float64

	// Scores for literals
	activity       map[sat_solver.CNFLiteral]float64

	// Scores for clauses
	// Those can be used to remove old learned clauses, but this feature is not yet implemented
	clauseActivity map[*sat_solver.CNFClause]float64

	// Ratios of decay
	varDecay       float64
	varThreshDecay float64
	clauseDecay    float64

	// Incrementation values for scores
	varInc    float64
	clauseInc float64

	// Priority queue with literals sorted by the activity score
	varOrderHeap *LiteralPriorityQueue
}

/**
 * Init AVSIDS variables.
 */
func (solver *CDCLSolver) avsidsInit() {
	/*
	 * Setup initial decay rates.
	 * This values are taken from a paper but sure, you can change that and see what happens!
	 */
	solver.varDecay = 0.85
	solver.varThreshDecay = 0.99
	solver.lbdEmaDecay = 0.95
	solver.clauseDecay = 0.99
	solver.varInc = 1
	solver.clauseInc = 1
	solver.lbdEma = 0

	// No literal was seen yet
	solver.lbdSeen = map[int]struct{}{}

	// Set initial scores
	solver.varOrderHeap = NewLiteralPriorityQueue(solver)
	vars := solver.vars.GetAllVariables()
	solver.activity = map[sat_solver.CNFLiteral]float64{}
	for _, v := range vars {
		if v < 0 {
			v = -v
		}
		solver.activity[v] = 0
	}
}

/**
 * Return recommended literal for decision based on AVSIDS heuristics.
 */
func (solver *CDCLSolver) avsidsSuggestSelect() (sat_solver.CNFLiteral, bool) {
	for solver.varOrderHeap.Len() > 0 {
		lit := heap.Pop(solver.varOrderHeap).(*PQLitItem)
		if lit.value < 0 {
			lit.value = -lit.value
		}
		if _, ok := solver.currentAssignment[lit.value]; !ok {
			return lit.value, true
		}
	}
	return sat_solver.CNF_UNDEFINED, false
}

/**
 * Increment scores for a clause.
 * Note: Clause scores are not yet handled correctly.
 *       Activity for a clause can be used to filter out old learned clauses.
 */
func (solver *CDCLSolver) avsidsBumpClauseActivity(clause *sat_solver.CNFClause) {
	// TODO: Capture clause scores here and use them to filter out learned clauses
	//solver.clauseActivity[clause] += solver.clauseInc
	//if solver.clauseActivity[clause] > 1e20 {
	//	for c := range solver.clauseActivity {
	//		solver.clauseActivity[c] *= 1e-20
	//	}
	//	solver.clauseInc *= 1e-20
	//}
}

/**
 * Decay scores for clauses.
 * Note: Clause scores are not yet handled correctly.
 *       Activity for a clause can be used to filter out old learned clauses.
 */
func (solver *CDCLSolver) avsidsDecayClauseActivity() {
	// TODO: Capture clauses scores here and use them to filter out learned clauses
	// solver.clauseInc *= (1 / solver.clauseDecay);
}

/**
 * Increment scores for the given literal.
 */
func (solver *CDCLSolver) avsidsBumpVarActivity(literal sat_solver.CNFLiteral) {
	if literal  < 0 {
		literal  = -literal
	}
	if _, ok := solver.activity[literal]; ok {
		solver.activity[literal] += solver.varInc
		if solver.activity[literal] > 1e100 {
			for varID := range solver.activity {
				solver.activity[varID] *= 1e-100
			}
			solver.varInc *= 1e-100
		}
	}
	solver.varOrderHeap.Update(literal)
}

/**
 * Decay activity incrementation for all variables.
 */
func (solver *CDCLSolver) avsidsDecayVarActivity(factor float64) {
	solver.varInc *= 1 / factor
}

/**
 * Handle new learned clause.
 */
func (solver *CDCLSolver) avsidsClauseLearnt(clause *sat_solver.CNFClause) {
	solver.avsidsBumpClauseActivity(clause)

	lbdVal := solver.lbd(*clause)
	solver.lbdEma = solver.lbdEmaDecay * solver.lbdEma + (1 - solver.lbdEmaDecay) * lbdVal;
	if lbdVal >= solver.lbdEma {
		solver.avsidsDecayVarActivity(solver.varDecay)
	} else {
		solver.avsidsDecayVarActivity(solver.varThreshDecay)
	}
	solver.avsidsDecayClauseActivity()
}

/**
 * Calculate LBD value for a literal.
 */
func (solver *CDCLSolver) lbd(clause sat_solver.CNFClause) float64 {
	lbd := float64(0)
	for _, lit := range clause {
		litVar := lit
		if litVar < 0 {
			litVar = -litVar
		}
		level := solver.getDecisionLevelForVar(litVar)
		if _, ok := solver.lbdSeen[level]; ok {
			solver.lbdSeen[level] = struct{}{}
			lbd++
		}
	}
	for _, lit := range clause {
		litVar := lit
		if litVar < 0 {
			litVar = -litVar
		}
		level := solver.getDecisionLevelForVar(litVar)
		solver.lbdSeen[level] = struct{}{}
	}
	return lbd
}