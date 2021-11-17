package sat_solver

import "fmt"

type SATVariableMapping struct {
	names map[string]CNFLiteral
	reverse map[CNFLiteral]string
	uniqueID CNFLiteral
	freshVarNameID uint64
}

func NewSATVariableMapping() *SATVariableMapping {
	return &SATVariableMapping{
		names:    map[string]CNFLiteral{},
		reverse:  map[CNFLiteral]string{},
		uniqueID: 2,
		freshVarNameID: 1,
	}
}

func (vars *SATVariableMapping) GetAllVariables() []CNFLiteral {
	ret := make([]CNFLiteral, len(vars.reverse))
	i := 0
	for v := range vars.reverse {
		ret[i] = v
		i++
	}
	return ret
}

func (vars *SATVariableMapping) IsFounderVariable(id CNFLiteral) bool {
		s := ""
		if id < 0 {
			s = vars.reverse[-id]
		} else {
			s = vars.reverse[id]
		}
		if len(s) >= 2 {
			if s[0] == '[' && s[len(s)-1] == ']' {
				return false
			}
			if s[0] == '"' && s[len(s)-1] == '"' {
				return true
			}
		}
		return true
}

func (vars *SATVariableMapping) Reverse(id CNFLiteral) string {
	if id < 0 {
		return fmt.Sprintf("-%s", trimVarQuotes(vars.reverse[-id]))
	}
	return trimVarQuotes(vars.reverse[id])
}

func (vars *SATVariableMapping) Fresh() (string, CNFLiteral) {
	newVarNameID := vars.freshVarNameID
	newID := vars.uniqueID
	name := fmt.Sprintf("[%d]", newVarNameID)
	vars.uniqueID++
	vars.freshVarNameID++
	vars.names[name] = newID
	vars.reverse[newID] = name
	return name, newID
}

func (vars *SATVariableMapping) Get(name string) CNFLiteral {
	if id, ok := vars.names[name]; ok {
		return id
	}
	newID := vars.uniqueID
	vars.uniqueID++
	vars.names[name] = newID
	vars.reverse[newID] = name
	return newID
}
