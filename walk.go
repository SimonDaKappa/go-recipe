package recipe

import (
	"fmt"
	"reflect"
)

var (
	ErrWalkFuncMismatch = fmt.Errorf("walk function does not match walk type requirements")
	ErrWalkTypeMismatch = fmt.Errorf("recipe walk type does not match executor walk type")
)

type WalkType uint8

const (
	// CombineWalk: Operations produce values that get combined via a Combiner.
	//
	// CombineWalk operations are required to depend on n instances of the
	// walked struct, where n is the arity of the operation.
	//
	// Example: Validation (combine bools), Masking (combine strings)
	//
	// See: [Combiner]
	CombineWalk WalkType = iota + 1

	// ApplyWalk: Operations produce values that get applied to the current
	// exec tree's level corresponding field in the walked struct.
	//
	// ApplyWalk operations are required to not depend on the walked struct
	// directly, but instead return the value to be applied from the n arguments,
	// of the operation, where n is the arity of the operation.
	//
	// Example: Parsing (*http.Request to struct), Mapping (struct to struct)
	//
	// See: [Applier]
	ApplyWalk

	// TransformWalk: Operations transform the walked struct in place.
	//
	// TransformWalk operations are required to depend on n instances of the
	// walked struct, where n is the arity of the operation.
	//
	// Example: Normalization (struct field value normalization), Sanitization
	//
	// See: [Transformer]
	TransformWalk
)

// Combiner coalesces multiple operation results into final output
// Used for RecipeWalk recipes.
type Combiner interface {
	// Initialize with zero value.
	Zero() any

	// Combine accumulator with new result.
	Combine(acc, result any) any
}

type BoolAndCombiner struct{}

func (c BoolAndCombiner) Zero() any { return true }
func (c BoolAndCombiner) Combine(acc, result any) any {
	return acc.(bool) && result.(bool)
}

type StringConcatCombiner struct{}

func (c StringConcatCombiner) Zero() any { return "" }
func (c StringConcatCombiner) Combine(acc, result any) any {
	if acc.(string) == "" {
		return result.(string)
	}
	return acc.(string) + result.(string)
}

// Applier applies operation result to a field of the walked struct.
//
// Used for AppliedWalk recipes.
type Applier interface {
	// Apply result to field at given offset in walked struct.
	Apply(walked any, fieldOffset uintptr, fieldType reflect.Type, value any) error
}

type ReflectSetterApplier struct{}

func (a ReflectSetterApplier) Apply(walked any, fieldOffset uintptr, fieldType reflect.Type, value any) error {
	return setField(walked, fieldOffset, fieldType, value)
}

func setField(walked any, fieldOffset uintptr, fieldType reflect.Type, value any) error {
	// TODO $$$SIMON
	return nil
}

type Transformer interface {
	Transform(walked any) error
}

type IntNormalizerTransformer struct{}

func (t IntNormalizerTransformer) Transform(walked any) error {
	// TODO $$$SIMON
	return nil
}
