package sat_solver

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

type CNFLiteral int64

const CNF_UNDEFINED = CNFLiteral(0)

type CNFClause []CNFLiteral

type CNFFormula struct {
	// -1 is false
	// 1 is true
	// x > 1 is a variable with unique id = x
	// x < 1 is a variable with unique id = -x
	Variables []CNFClause
}

func (clause CNFClause) Copy() CNFClause {
	newClause := make(CNFClause, len(clause))
	copy(newClause, clause)
	return newClause
}

func (f *CNFFormula) AST(vars *SATVariableMapping) *Formula {
	var currentRet *Formula = nil
	for _, clause := range f.Variables {
		var currentClause *Formula = nil
		for _, v := range clause {
			if currentClause == nil {
				if v < 0 {
					currentClause = MakeNot(MakeVar(vars.Reverse(v)))
				} else {
					currentClause = MakeVar(vars.Reverse(v))
				}
			} else {
				if v < 0 {
					currentClause = MakeOr(currentClause, MakeNot(MakeVar(vars.Reverse(v))))
				} else {
					currentClause = MakeOr(currentClause, MakeVar(vars.Reverse(v)))
				}
			}
		}
		if currentRet == nil {
			currentRet = currentClause
		} else {
			currentRet = MakeAnd(currentRet, currentClause)
		}
	}
	return currentRet
}

type UnsatReasonCNFNormalization struct {}

func NewUnsatReasonCNFNormalization() *UnsatReasonCNFNormalization {
	return &UnsatReasonCNFNormalization{}
}

func (reason *UnsatReasonCNFNormalization) Describe() string {
	return fmt.Sprintf("Empty clause detected when normalizing CNF formula.")
}

func (f *CNFFormula) NormalizeVars(vars *SATVariableMapping) (error, *SATVariableMapping, int) {
	newVars := &SATVariableMapping{
		names:    map[string]CNFLiteral{},
		reverse:  map[CNFLiteral]string{},
		uniqueID: 1,
		freshVarNameID: vars.freshVarNameID,
	}
	newMapping := map[CNFLiteral]CNFLiteral{}

	for i, clause := range f.Variables {
		if len(clause) == 0 {
			return NewUnsatError(NewUnsatReasonCNFNormalization()), nil, 0
		}
		for j, v := range clause {
			varID := v
			neg := false
			if varID < 0 {
				varID = -varID
				neg = true
			}
			if entry, ok := newMapping[varID]; ok {
				if neg {
					f.Variables[i][j] = -entry
				} else {
					f.Variables[i][j] = entry
				}
			} else {
				newID := newVars.uniqueID
				newMapping[varID] = newID
				if neg {
					f.Variables[i][j] = -newID
				} else {
					f.Variables[i][j] = newID
				}

				newVars.reverse[newID] = vars.reverse[varID]
				newVars.names[vars.reverse[varID]] = newID

				newVars.uniqueID++
			}
		}
	}

	return nil, newVars, len(newVars.names)
}

func (f *CNFFormula) Evaluate(vars []bool) bool {
	for _, clause := range f.Variables {
		isClauseOk := false
		for _, v := range clause {
			if v > 0 && vars[v-1] {
				isClauseOk = true
				break
			} else if v < 0 && !vars[-v-1] {
				isClauseOk = true
				break
			}
		}
		if !isClauseOk {
			return false
		}
	}
	return true
}

func (f *CNFFormula) AndWith(e *CNFFormula) {
	f.Variables = append(f.Variables, e.Variables...)
}

func (f *CNFFormula) MulWith(e *CNFFormula) {
	lenF := len(f.Variables)
	lenE := len(e.Variables)
	newVariables := make([]CNFClause, 0, lenF * lenE)
	shouldAddE := true
	shouldAddF := true

	for _, clauseF := range f.Variables {
		shouldAddF = true
		for _, varF := range clauseF {
			if varF == 1 {
				shouldAddF = false
				break
			}
		}
		if !shouldAddF {
			continue
		}
		for _, clauseE := range e.Variables {
			shouldAddE = true
			for _, varE := range clauseE {
				if varE == 1 {
					shouldAddE = false
					break
				}
			}
			if shouldAddE {
				newVariables = append(newVariables, append(clauseF, clauseE...))
			}
		}
	}
	f.Variables = newVariables
}

func (f *CNFFormula) Measure() *SATFormulaStatistics {
	varIDs := map[CNFLiteral]struct{}{}
	clauseCount := int64(len(f.Variables))
	longestClause := int64(0)
	shortestClause := int64(math.MaxInt64)
	clauseLenSum := int64(0)
	for _, clause := range f.Variables {
		l := int64(len(clause))
		if l > longestClause {
			longestClause = l
		}
		if l < shortestClause {
			shortestClause = l
		}
		clauseLenSum += l
		for _, varID := range clause {
			if varID > 0 {
				varIDs[varID] = struct{}{}
			} else {
				varIDs[-varID] = struct{}{}
			}
		}
	}
	varCount := int64(0)
	for range varIDs {
		varCount++
	}
	return &SATFormulaStatistics{
		variableCount:    varCount,
		clauseCount:      clauseCount,
		clauseLenSum:     clauseLenSum,
		clauseDepth:      2,
		clauseComplexity: clauseLenSum,
		isCNF: true,
	}
}

func (f *CNFFormula) SaveDIMACSCNF(writer io.Writer, varNames *SATVariableMapping) error {
	vars := make([]CNFClause, len(f.Variables))
	variableRemap := map[CNFLiteral]CNFLiteral{}
	variableNames := map[CNFLiteral]string{}
	freeID := CNFLiteral(1)
	for i, clause := range f.Variables {
		vars[i] = make(CNFClause, len(clause))
		for j, v := range clause {
			if v == 1 || v == -1 || v == 0 {
				return fmt.Errorf("CNF formula cannot contain T/F.")
			}
			if v < 0 {
				if k, ok := variableRemap[-v]; ok {
					vars[i][j] = -k
				} else {
					variableRemap[-v] = freeID
					vars[i][j] = -freeID
					variableNames[-v] = varNames.Reverse(-v)
					freeID++
				}
			} else {
				if k, ok := variableRemap[v]; ok {
					vars[i][j] = k
				} else {
					variableRemap[v] = freeID
					variableNames[v] = varNames.Reverse(v)
					vars[i][j] = freeID
					freeID++
				}
			}
		}
	}

	for v, name := range variableNames {
		if varNames.IsFounderVariable(v) {
			_, err := writer.Write([]byte(fmt.Sprintf("c  %d => Variable \"%s\"\n", v, name)))
			if err != nil {
				return err
			}
		}
	}

	_, err := writer.Write([]byte(fmt.Sprintf("p cnf %d %d\n", len(variableRemap), len(vars))))
	if err != nil {
		return err
	}
	for _, clause := range vars {
		for _, v := range clause {
			if v == 0 {
				return fmt.Errorf("Detected 0 during converting to CNF file format.")
			}
			_, err = writer.Write([]byte(fmt.Sprintf("%d ", v)))
			if err != nil {
				return err
			}
		}
		_, err = writer.Write([]byte("0\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *CNFFormula) SaveDIMACSCNFToFile(filePath string, vars *SATVariableMapping) error {
	outputFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	err = f.SaveDIMACSCNF(outputFile, vars)
	if err != nil {
		return err
	}
	return outputFile.Close()
}

func (literal CNFLiteral) DebugString() string {
	v := literal
	if (v == 1) {
		return "True"
	} else if (v == -1) {
		return "False"
	} else if (v > 0) {
		return strconv.Itoa(int(v));
	} else {
		return "-" + strconv.Itoa(int(-v))
	}
}

func LiteralSliceString(literals []CNFLiteral, vars *SATVariableMapping) string {
	partialResult := make([]string, len(literals))
	for i, id := range literals {
		partialResult[i] = id.String(vars)
	}
	return "[" + strings.Join(partialResult, "; ") + "]"
}

func (clause CNFClause) DebugString() string {
	partialResult := make([]string, len(clause))
	for i, id := range clause {
		partialResult[i] = id.DebugString()
	}
	return "(" + strings.Join(partialResult, " v ") + ")"
}

func (literal CNFLiteral) String(vars *SATVariableMapping) string {
	v := literal
	if (v == 1) {
		return "True"
	} else if (v == -1) {
		return "False"
	} else if (v > 0) {
		return trimVarQuotes(vars.reverse[v])
	} else {
		return "-" + trimVarQuotes(vars.reverse[-v])
	}
}

func (clause CNFClause) String(vars *SATVariableMapping) string {
	partialResult := make([]string, len(clause))
	for i, id := range clause {
		partialResult[i] = id.String(vars)
	}
	return "(" + strings.Join(partialResult, " v ") + ")"
}

func (f *CNFFormula) String(vars *SATVariableMapping) string {
	result := make([]string, len(f.Variables))
	for j, clause := range f.Variables {
		result[j] = clause.String(vars)
	}
	return strings.Join(result, "^")
}