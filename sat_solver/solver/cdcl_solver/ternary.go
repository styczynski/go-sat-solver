package cdcl_solver

import "fmt"

/**
 * Boolean type with undefined state.
 */
type Ternary int8

/**
 * Ternary logic states
 */
const (
	TERNARY_TRUE      Ternary = 1
	TERNARY_FALSE             = 2
	TERNARY_UNDEFINED         = 3
)

/**
 * Check if the value is true.
 */
func (t Ternary) IsTrue() bool {
	return t == TERNARY_TRUE
}

/**
 * Check if the value is false.
 */
func (t Ternary) IsFalse() bool {
	return t == TERNARY_FALSE
}

/**
 * Negate value.
 */
func (t Ternary) Negate() Ternary {
	if t == TERNARY_FALSE {
		return TERNARY_TRUE
	} else if t == TERNARY_TRUE {
		return TERNARY_FALSE
	}
	return TERNARY_UNDEFINED
}

/**
 * Check if the value is undefined.
 */
func (t Ternary) IsUndefined() bool {
	return t == TERNARY_UNDEFINED
}

/**
 * Convert ternary value to boolean.
 * Note:
 *   If the value is undefined then the function will panic.
 */
func TernaryToBool(value Ternary) bool {
	if value == TERNARY_UNDEFINED {
		panic(fmt.Errorf("Tried to convert ternary undefined to bool"))
	}
	return value == TERNARY_TRUE
}

/**
 * Convert bool value to ternary.
 */
func BoolToTernary(value bool) Ternary {
	if value {
		return TERNARY_TRUE
	}
	return TERNARY_FALSE
}