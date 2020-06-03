package cdcl_solver

import (
	"github.com/go-sat-solver/sat_solver"
)

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
		traceLiteral := solver.assignmentTrace[traceIndex]
		if traceLiteral < 0 {
			traceLiteral = -traceLiteral
		}
		for !solver.visited[traceLiteral] {
			traceIndex = traceIndex-1
			traceLiteral = solver.assignmentTrace[traceIndex]
			if traceLiteral < 0 {
				traceLiteral = -traceLiteral
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
