package recipe

import (
	"encoding"
)

type Grammar interface {
	Key() string
	Description() string

	// Every recipe from this grammar uses the same walk type
	WalkType() WalkType
	// Combiner combines multiple operation results into final
	// output
	//
	// This is a default value, and can be overridden per-recipe,
	// in [StructExecutor.ExecuteCombinerWalk()]
	//
	// return nil, error if the walk type does not use a combiner
	Combiner() (Combiner, error)
	// Applier applies operation result to a field of the
	// walked struct.
	//
	// This is a default value, and can be overridden per-recipe,
	// in [StructExecutor.ExecuteApplierWalk()]
	//
	// return nil, error if the walk type does not use an applier
	Applier() (Applier, error)
	// Transformer transforms the walked struct after all
	// operations have been applied.
	//
	// This is a default value, and can be overridden per-recipe,
	// in [StructExecutor.ExecuteTransformerWalk()]
	//
	// return nil, error if the walk type does not use a transformer
	Transformer() (Transformer, error)

	// Split splits the tag value into a list of all operation strings
	// according to the grammar's structure and format.
	//
	// e.g., for a tag `<grammar_key>:op_str1,op_str2,op_str3`,
	// Split would return: []string{"op_str1", "op_str2", "op_str3"}
	//
	// Return an error if the tag does not conform to the grammar's
	// structure and format.
	Split(tag string) ([]string, error)

	// Parse parses an operation string into a LazyOperation
	//
	// e.g., for an op_str `bind=header;name=Authorization`,
	// Parse would return: LazyOperation{Name: "bind=header", ...}
	//
	// Return an error if the operation string does not conform to
	// the grammar's operation specification.
	Parse(opstr string) (LazyOperation, error)

	// Order orders a list of LazyOperations according to the
	// grammar's operation ordering rules.
	//
	// e.g., for a list of LazyOperations with names
	// []string{"op_str2", "op_str1", "op_str3"},
	// and lexicographical ordering, returns:
	// []string{"op_str1", "op_str2", "op_str3"}
	//
	// Return an error if the list of LazyOperations cannot be
	// ordered according to the grammar's operation ordering rules.
	Order(lazyOps []LazyOperation) ([]LazyOperation, error)
}

type GrammarStructure uint8

var (
	// List style must be flattened, i.e, no recursive KV parsing
	//
	// $$$TODO $$$SIMON: Add top-level pattern compiling
	StructureFlat GrammarStructure = 0

	// List style can be hierarchical, i.e, recursive KV parsing allowed
	//
	// $$$SIMON: Planned for future. v0.0.1 will not support hierarchical grammars.
	StructureHierarchy GrammarStructure = 1
)

//--------------------------------------------------------------------------------
//
//--------------------------------------------------------------------------------

type FlatGrammarFormat uint8

var (
	// `<grammar_key>:<op_str1><custom_delimiter><op_str2>...`\
	//
	// $$$TODO $$$SIMON: Add delimiter capture-group pattern compiling
	FlatFormatDelimited FlatGrammarFormat = 1

	// `<grammar_key>:[<op_str1>],[<op_str2>]`
	//
	// $$$TODO $$$SIMON: Add encloser capture-group pattern compiling
	FlatFormatEnclosed FlatGrammarFormat = 3
)

// ALWAYS len 2
//
// $$$TODO $$$SIMON: Add to flat grammar pattern compiler.
type FlatGrammarEncloser string

const (
	EncloserSquareBracket FlatGrammarEncloser = "[]"
	EncloserCurlyBrace    FlatGrammarEncloser = "{}"
	EncloserParenthesis   FlatGrammarEncloser = "()"
)

//--------------------------------------------------------------------------------
//
//--------------------------------------------------------------------------------

type HierarchyGrammarFormat uint8

const (
	// $$$SIMON: Planned for future. v0.0.1 will not support hierarchical grammars.
	// Will most likely need to revisit this to actually develop json specs.
	// `<grammar_key>:<op_json_object>` e.g.
	//
	//
	//  <grammar_key>: "{
	//    \"operations\" : [
	//      {\"name\": \"<op_name_1>\", \"mod1\": \"value1\", ...},
	//      {\"name\": \"<op_name_2>\", \"mod2\": \"value2\", ...}
	//    ]
	//  }"
	// Note: reflect ONLY tracks the depth of quotes if properly escaped in values.
	HierarchyFormatJSON HierarchyGrammarFormat = iota + 1
)

//--------------------------------------------------------------------------------
//
//--------------------------------------------------------------------------------

// $$$TODO $$$SIMON: Add arities to top-level grammar pattern compiler.
type GrammarArity uint8

const (
	// `<grammar_key>:<list_style><op_str><\list_style>`
	GrammarOpArityUnary GrammarArity = iota + 1

	// `<grammar_key>:<list_style><op_str1>,<op_str2>,...<\list_style>`
	GrammarOpArityVariadic
)

//--------------------------------------------------------------------------------
//
//--------------------------------------------------------------------------------

type ModifierUse uint8

const (
	// Modifier is used in the execution of the operation.
	//
	// e.g., in `bind=header,omitempty,default=3`, `omitempty` and `default`
	// are used by the [Executor] to handle empty values
	ModifierUseExecution ModifierUse = iota + 1

	// Modifier is used by the [Operation] itself during its processing.
	//
	// e.g., in `mask:"email,density=0.8"`, `density` could be used by the email
	// masking [Operation] to mask 80% of the email address characters.
	ModifierUseOperation
)

// $$$TODO $$$SIMON: Add modifier pattern compiling to operation pattern compiler.
type ModifierFormat uint8

const (
	// `<modifier_key>=<modifier_value>`
	//
	// # Can be used with any [ModifierValueType]
	ModFormatKV ModifierFormat = iota + 1

	// `<modifier_key>` (implies boolean true)
	//
	// # Can only be used with [ModTypeBool]
	ModFormatKeyOnly
)

// $$$SIMON: Are modifiers allowed to be quoted by builtin grammar builder?
type ModifierValueType uint

const (
	// ModTypeBool is a string that can be parsed as a boolean.
	// Used with [ModFormatKV] and [ModFormatKeyOnly]
	//
	// e.g., "true", "false", "1", "0", "yes", "no"
	ModTypeBool ModifierValueType = iota + 1

	// ModTypeInt is a string that can be parsed as an integer.
	// Used with [ModFormatKV]
	//
	// e.g., "42", "-1", "0"
	ModTypeInt

	// ModTypeUInt is a string that can be parsed as an unsigned integer.
	// Used with [ModFormatKV]
	//
	// e.g., "42", "0"
	ModTypeUInt

	// ModTypeFloat is a string that can be parsed as a float.
	// Used with [ModFormatKV]
	//
	// e.g., "3.14", "-0.001", "0.0"
	ModTypeFloat

	// ModTypeComplex is a string that can be parsed as a complex number.
	// Used with [ModFormatKV]
	//
	// e.g., "1+2i", "3-4i"
	ModTypeComplex

	// ModTypeString is any string.
	// Used with [ModFormatKV]
	//
	// e.g., "hello", "world", "any string"
	ModTypeString

	// ModTypeConverted is a special type indicating the value must be parsed
	// from string to a custom type using the provided ($$$SIMON $$$TODO)
	// reflective conversion functions. If the field implements
	// [encoding.TextUnmarshaler],
	// that will be used. Otherwise, a built-in parser for common types will be used.
	// Used with [ModFormatKV]
	//
	// e.g., custom types like time.Time, UUID, etc.
	ModTypeConverted ModifierValueType = 0xFF
)

//--------------------------------------------------------------------------------
//
//--------------------------------------------------------------------------------

type FlatOperationSpec struct {
	Key string
}

type FlatGrammarSpec struct {
	Format    FlatGrammarFormat
	ModFormat ModifierFormat

	// Used if Format == FlatFormatEnclosed
	Encloser FlatGrammarEncloser

	// Used if Format == FlatFormatDelimited
	Delimiter string

	// Used only if a specific set of operations need additional
	// specification beyond the default modifier format.
	//
	// Un-Keyed operations use default modifier formats, and
	// function under lazy resolution as usual.
	CustomOpSpecs map[string]FlatOperationSpec
}

type HierarchyGrammarSpec struct {
	Format HierarchyGrammarFormat
}

type GrammarBuilder struct {
	Structure GrammarStructure
	Arity     GrammarArity

	// Used if Structure == StructureFlat
	FlatSpec FlatGrammarSpec

	// Used if Structure == StructureHierarchy
	HierarchySpec HierarchyGrammarSpec
}

func NewGrammarBuilder() GrammarBuilder {
	return GrammarBuilder{}
}

func (builder GrammarBuilder) SetStructure(structure GrammarStructure) GrammarBuilder {
	builder.Structure = structure
	return builder
}

func (builder GrammarBuilder) SetOpArity(arity GrammarArity) GrammarBuilder {
	builder.Arity = arity
	return builder
}

func (builder GrammarBuilder) Compile() Grammar { return nil }
