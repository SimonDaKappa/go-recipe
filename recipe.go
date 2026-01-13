package recipe

import (
	"reflect"
	"unsafe"
)

type Recipe struct {
	Root *ExecTree

	Arity OpArity

	// WalkType determines how the recipe is executed.
	// It also determines whether combiner/applier/transformer
	// are used.
	//
	// See: [RecipeWalkType]
	WalkType WalkType

	combiner    Combiner
	applier     Applier
	transformer Transformer
	resolved    bool
}

type ExecContext struct {
	CombinerOverride    Combiner
	ApplierOverride     Applier
	TransformerOverride Transformer
}

// ExecTree now supports multiple operations per field
type ExecTree struct {
	// Field name in the struct
	Field string

	fieldIdx    int          // Index in parent struct.Fields
	fieldOffset uintptr      // Offset in parent struct
	fieldType   reflect.Type // Type of the field
	fieldKind   reflect.Kind // Kind of the field

	// Lazy Operation References. During recipe resolving,
	// these are converted to ResolvedOperations.
	LazyOps []LazyOperation
	// Exec-time Resolved Operations. Populated during recipe
	// resolving. What is actually called during recipe
	// execution.
	Operations []ResolvedOperation
	// Strategy for multiple operations on this field.
	// Determines how multiple operations are executed.
	OpStrategy MultiOpStrategy

	// Tree structure for nested fields
	//
	// Empty if this recipe is not a struct
	Children []*ExecTree

	// fieldExtractor: pre-compiled field extractor to avoid reflect.NewAt overhead.
	// Takes a pointer to parent struct, returns field value as any
	//
	// Only non-nil if this is a leaf node (no Children)
	fieldExtractor func(structPtr unsafe.Pointer) any
	// structExtractor: pre-compiled struct extractor to avoid reflect.NewAt overhead.
	// Takes pointer to parent struct, returns pointer to child struct
	//
	// Only non-nil if this is a struct node (has Children)
	structExtractor func(structPtr unsafe.Pointer) unsafe.Pointer
}

func (t *ExecTree) isStruct() bool {
	return t.Children != nil && len(t.Children) > 0
}

func (t *ExecTree) hasOperations() bool {
	return t.Operations != nil && len(t.Operations) > 0
}

func (t *ExecTree) Strategy() MultiOpStrategy {
	return t.OpStrategy
}

func NoopExecTree(tree *ExecTree) error {
	return nil
}
