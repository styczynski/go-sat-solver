package cdcl_solver

/**
 * This file provides learning mechanism implementation for a CDCL solver.
 * You can read more about how CDCL learns new clauses here:
 *   https://en.wikipedia.org/wiki/Conflict-driven_clause_learning
 *
 * You can also see the Minisat C++ code:
 *   Solver::analyze  https://github.com/niklasso/minisat/blob/master/minisat/core/Solver.cc#L296
 *
 */

import (
	"github.com/go-sat-solver/sat_solver"
)

type SolverLearnState struct {
	// Learned clause is stored here
	currentLearnedClause   sat_solver.CNFClause
	// Map of visited literals is used only in learnClause() to prevent updating literals twice
	visited                map[sat_solver.CNFLiteral]bool
}

/**
 * Having a clause that caused a conflict to arise, this function updates currentLearnedClause
 * and returns the decision level that the solver should use to jump backwards.
 */
func (solver *CDCLSolver) learnClause(conflictingClause sat_solver.CNFClause) int {
	literalsLeft := 0
	solver.currentLearnedClause = solver.currentLearnedClause[:1]
	traceLiteral := sat_solver.CNF_UNDEFINED
	traceIndex := len(solver.assignmentTrace) - 1
	for {
		learnedClauseStartIndex := 0
		if traceLiteral == sat_solver.CNF_UNDEFINED {
			learnedClauseStartIndex = 0
		} else {
			learnedClauseStartIndex = 1
		}

		for _, learnedClauseLiteral := range conflictingClause[learnedClauseStartIndex:] {
			learnedVar := learnedClauseLiteral
			if learnedVar < 0 {
				learnedVar = -learnedVar
			}
			learnedVarLevel := solver.getDecisionLevelForVar(learnedVar)
			if !solver.visited[learnedVar] && learnedVarLevel > 0 {
				solver.avsidsBumpVarActivity(learnedVar);
				if learnedVarLevel >= solver.getDecisionLevel() {
					literalsLeft++
				} else {
					solver.currentLearnedClause = append(solver.currentLearnedClause, learnedClauseLiteral)
				}
				solver.visited[learnedVar] = true
			}
		}

		/**
		 * This loop gets the last visited literal on a trace
		 */
		traceVisitedLiteral := solver.assignmentTrace[traceIndex]
		if traceVisitedLiteral < 0 {
			traceVisitedLiteral = -traceVisitedLiteral
		}
		for !solver.visited[traceVisitedLiteral] {
			traceIndex = traceIndex-1
			traceVisitedLiteral = solver.assignmentTrace[traceIndex]
			if traceVisitedLiteral < 0 {
				traceVisitedLiteral = -traceVisitedLiteral
			}
		}
		// But go back one step more
		traceIndex = traceIndex-1

		traceLiteral = solver.assignmentTrace[traceIndex+1]
		traceLiteralVar := traceLiteral
		if traceLiteralVar < 0 {
			traceLiteralVar = -traceLiteralVar
		}
		solver.visited[traceLiteralVar] = false
		conflictingClause = solver.varsInfo[traceLiteralVar].reasonClause

		if literalsLeft <= 1 {
			break
		} else {
			literalsLeft = literalsLeft - 1
		}
	}
	solver.currentLearnedClause[0] = -traceLiteral

	/**
	 * Detect the decision level to jump to
	 */
	jumpToDecisionLevel := 0
	if len(solver.currentLearnedClause) > 1 {
		maxDecisionLevelVarIndex := 1
		learnedVarMax := solver.currentLearnedClause[maxDecisionLevelVarIndex]
		if learnedVarMax < 0 {
			learnedVarMax = -learnedVarMax
		}
		maxLevel := solver.getDecisionLevelForVar(learnedVarMax)
		for i, learnedVar := range solver.currentLearnedClause[2:] {
			if learnedVar < 0 {
				learnedVar = -learnedVar
			}
			varLevel := solver.getDecisionLevelForVar(learnedVar);
			if varLevel > maxLevel {
				maxLevel = varLevel
				maxDecisionLevelVarIndex = i + 2
			}
		}
		maxLiteral := solver.currentLearnedClause[maxDecisionLevelVarIndex]
		solver.currentLearnedClause[maxDecisionLevelVarIndex] = solver.currentLearnedClause[1]
		solver.currentLearnedClause[1] = maxLiteral
		jumpToDecisionLevel = maxLevel
	}

	for _, literal := range solver.currentLearnedClause {
		literalVar := literal
		if literalVar < 0 {
			literalVar = -literalVar
		}
		solver.visited[literalVar] = false
	}

	return jumpToDecisionLevel
}
