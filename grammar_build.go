package recipe

import (
	"fmt"
	"regexp"
)

var (
	// Matches individual recipes: [bind=json,key=user_id,omitempty,default=0,...]
	// Captures everything inside the brackets
	recipePattern = regexp.MustCompile(`\[([^\]]+)\]`)

	// Matches key=value pairs
	// Group 1: key
	// Group 2: quoted value (double quotes) - content without quotes
	// Group 3: quoted value (single quotes) - content without quotes
	// Group 4: unquoted value
	kvPattern = regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)=(?:"((?:[^"\\]|\\.)*)"|'((?:[^'\\]|\\.)*)'|([^,\]]*))`)
)

type GrammarStructure uint8

const (
	// GStructureInvalid indicates an invalid grammar structure.
	//
	// Catches zero-value errors.
	GStructureInvalid GrammarStructure = iota

	// List style must be flattened, i.e, no recursive KV parsing
	//
	// $$$TODO $$$SIMON: Add top-level pattern compiling
	GStructureFlat

	// List style can be hierarchical, i.e, recursive KV parsing allowed
	//
	// $$$SIMON: Planned for future. v0.0.1 will not support hierarchical grammars.
	GStructureHierarchy
)

type FlatGrammarTemplate uint8

const (
	// FGTInlineInvalid indicates an invalid flat grammar template.
	//
	// Catches zero-value errors.
	FTemplateInvalid FlatGrammarTemplate = iota

	// FTemplateInlineComma indicates a flat grammar template with inline comma separation.
	//
	// e.g., `<grammar_key>:"<op_str1>,<op_str2>,..."`
	// <op_str>:
	//  `<operation_key>`
	//
	// This is a very simple template, and does not allow for any [Modifier]'s to be
	// used, as there is no way to escape commas in modifier values.
	FTemplateInlineComma

	// FTemplateInlineSemi indicates a flat grammar template with inline semicolon separation.
	// e.g., `<grammar_key>:"<op_str1>;<op_str2>;..."`
	//
	// This simple template allows for modifiers to be used, as semicolons are less likely
	// to appear in modifier values.
	FTemplateInlineSemi

	// FTemplateInlinePipe indicates a flat grammar template with inline pipe separation.
	// e.g., `<grammar_key>:"<op_str1>|<op_str2>|..."`
	//
	// This simple template allows for modifiers to be used, as pipes are less likely
	// to appear in modifier values.
	FTemplateInlinePipe

	// FTemplatePairSquare indicates a flat grammar template with square bracket pairs around
	// each operation string.
	// e.g., `<grammar_key>:"[<op_str1>][<op_str2>]..."`
	//
	// This template allows for complex modifier values, as the brackets can be used
	// to clearly delineate each operation string.
	FTemplatePairSquare

	// FTemplatePairCurly indicates a flat grammar template with curly brace pairs around
	// each operation string.
	// e.g., `<grammar_key>:"{<op_str1>}{<op_str2>}..."`
	//
	// This template allows for complex modifier values, as the braces can be used
	// to clearly delineate each operation string.
	FTemplatePairCurly

	// FTemplatePairParen indicates a flat grammar template with parenthesis pairs around
	// each operation string.
	// e.g., `<grammar_key>:"(<op_str1>)(<op_str2>)..."`
	//
	// This template allows for complex modifier values, as the parentheses can be used
	// to clearly delineate each operation string.
	FTemplatePairParen
)

type FlatGrammarFormat uint8

const (
	// FFormatInvalid indicates an invalid flat grammar format.
	//
	// Catches zero-value errors.
	FFormatInvalid FlatGrammarFormat = iota

	// `<grammar_key>:<op_str1><custom_delimiter><op_str2>...`\
	//
	// $$$TODO $$$SIMON: Add delimiter capture-group pattern compiling
	FFormatDelimited

	// `<grammar_key>:[<op_str1>],[<op_str2>]`
	//
	// $$$TODO $$$SIMON: Add encloser capture-group pattern compiling
	FFormatEnclosed
)

type FlatGrammarSeparator string

const (
	FSepPairSquare   FlatGrammarSeparator = "[]"
	FSepPairCurl     FlatGrammarSeparator = "{}"
	FSepPairParen    FlatGrammarSeparator = "()"
	FSepInlineCommma FlatGrammarSeparator = ","
	FSepInlinePipe   FlatGrammarSeparator = "|"
	FSepInlineSemi   FlatGrammarSeparator = ";"
)

type HierarchyGrammarTemplate uint8

const (
	//
	HTemplateInvalid HierarchyGrammarTemplate = iota
	// HGTJson Provides a template grammar with all formatting
	// expected to be in JSON.
	HTemplateJSON
)

type HierarchyGrammarFormat uint8

const (
	// HFormatInvalid indicates an invalid hierarchy grammar format.
	//
	// Catches zero-value errors.
	HFormatInvalid HierarchyGrammarFormat = iota

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
	HFormatJSON
)

// $$$TODO $$$SIMON: Add arities to top-level grammar pattern compiler.
type GrammarArity uint8

const (
	// GrammarArityInvalid indicates an invalid grammar arity.
	//
	// Catches zero-value errors.
	GArityInvalid GrammarArity = iota

	// `<grammar_key>:<list_style><op_str><\list_style>`
	GArityUnary

	// `<grammar_key>:<list_style><op_str1>,<op_str2>,...<\list_style>`
	GArityVariadic
)

func (ga GrammarArity) String() string {
	switch ga {
	case GArityUnary:
		return "Unary"
	case GArityVariadic:
		return "Variadic"
	default:
		return "Unknown"
	}
}

type ModUse uint8

const (
	// ModUseInvalid indicates an invalid modifier use.
	//
	// Catches zero-value errors.
	ModUseInvalid ModUse = iota

	// Modifier is used in the execution of the operation.
	//
	// e.g., in `bind=header,omitempty,default=3`, `omitempty` and `default`
	// are used by the [Executor] to handle empty values
	ModUseExec

	// Modifier is used by the [Operation] itself during its processing.
	//
	// e.g., in `mask:"email,density=0.8"`, `density` could be used by the email
	// masking [Operation] to mask 80% of the email address characters.
	ModUseOp
)

func (mu ModUse) String() string {
	switch mu {
	case ModUseExec:
		return "Execution"
	case ModUseOp:
		return "Operation"
	default:
		return "Unknown"
	}
}

// $$$TODO $$$SIMON: Add modifier pattern compiling to operation pattern compiler.
type ModType uint8

const (
	// ModTypeInvalid indicates an invalid modifier type.
	//
	// Catches zero-value errors.
	ModTypeInvalid ModType = iota

	// `<modifier_key><mod_kv_delim><modifier_value>`
	//
	// Can be used with any [ModKind]
	ModTypeKeyVal

	// `<modifier_key>` (implies boolean true)
	//
	// Can only be used with [ModKindBool]
	ModTypeKeyOnly
)

type ModKVDelim rune

const (
	ModKVDelimEqual     ModKVDelim = '='
	ModKVDelimColon     ModKVDelim = ':'
	ModKVDelimSemiColon ModKVDelim = ';'
)

func (mf ModType) String() string {
	switch mf {
	case ModTypeKeyVal:
		return "Key-Value"
	case ModTypeKeyOnly:
		return "Key-Only"
	default:
		return "Unknown"
	}
}

// $$$SIMON: Are modifiers allowed to be quoted by builtin grammar builder?
type ModKind uint

const (
	// ModKindInvalid indicates an invalid modifier kind.
	//
	// Catches zero-value errors.
	ModKindInvalid ModKind = iota

	// ModKindBool is a string that can be parsed as a boolean.
	// Used with [ModTypeKeyVal] and [ModTypeKeyOnly]
	//
	// e.g., "true", "false", "1", "0", "yes", "no"
	ModKindBool

	// ModKindInt is a string that can be parsed as an integer.
	// Used with [ModTypeKeyVal]
	//
	// e.g., "42", "-1", "0"
	ModKindInt

	// ModKindUInt is a string that can be parsed as an unsigned integer.
	// Used with [ModTypeKeyVal]
	//
	// e.g., "42", "0"
	ModKindUInt

	// ModKindFloat is a string that can be parsed as a float.
	// Used with [ModTypeKeyVal]
	//
	// e.g., "3.14", "-0.001", "0.0"
	ModKindFloat

	// ModKindComplex is a string that can be parsed as a complex number.
	// Used with [ModTypeKeyVal]
	//
	// e.g., "1+2i", "3-4i"
	ModKindComplex

	// ModKindString is any string.
	// Used with [ModTypeKeyVal]
	//
	// e.g., "hello", "world", "any string"
	ModKindString

	// ModKindConverted is a special type indicating the value must be parsed
	// from string to a custom type using the provided ($$$SIMON $$$TODO)
	// reflective conversion functions. If the field implements
	// [encoding.TextUnmarshaler],
	// that will be used. Otherwise, a built-in parser for common types will be used.
	// Used with [ModTypeKeyVal]
	//
	// e.g., custom types like time.Time, UUID, etc.
	ModKindConverted ModKind = 0xFF
)

func (mk ModKind) String() string {
	switch mk {
	case ModKindBool:
		return "Bool"
	case ModKindInt:
		return "Int"
	case ModKindUInt:
		return "UInt"
	case ModKindFloat:
		return "Float"
	case ModKindComplex:
		return "Complex"
	case ModKindString:
		return "String"
	case ModKindConverted:
		return "Converted"
	default:
		return "Unknown"
	}
}

type ModifierSpec struct {
	modkey string
	use    ModUse
	kind   ModKind
	format ModType
}

type OpModifierSpec struct {
	ModifierSpec
	overrideShared bool
}

type OperationSpec struct {
	opkey    string
	modSpecs map[string]OpModifierSpec
}

type GrammarBuildStage uint8

const (
	StageFormatValidation GrammarBuildStage = iota + 1
	StagePatternCompilation
	StageModifierValidation
	StageOperationValidation
	StageFinalization
)

func (s GrammarBuildStage) String() string {
	switch s {
	case StageFormatValidation:
		return "Format Validation"
	case StagePatternCompilation:
		return "Pattern Compilation"
	case StageModifierValidation:
		return "Modifier Validation"
	case StageOperationValidation:
		return "Operation Validation"
	case StageFinalization:
		return "Finalization"
	default:
		return "Unknown Stage"
	}
}

type GrammarBuildError struct {
	Stage  GrammarBuildStage
	Reason string
	Value  any
}

func (e GrammarBuildError) Error() string {
	return fmt.Sprintf("Grammar Build Stage %s: %s (Value: %v)", e.Stage, e.Reason, e.Value)
}

type GrammarConfig interface {

	// SetKey sets the unique key for the grammar
	//
	// This is the key that will be used in struct tags to
	// identify this grammar.
	//
	// e.g., <field> <type> `<key>:<content>`
	SetKey(key string) GrammarConfig

	// SetDescription sets a human-readable description for the grammar
	//
	// This is optional, but can be useful for documentation
	// and debugging purposes.
	//
	// e.g., "JSON Binding Grammar"
	SetDescription(desc string) GrammarConfig

	// SetWalkType sets the default walk type for the grammar
	//
	// This determines how the grammar will be used during
	// execution.
	//
	// e.g., [CombineWalk], [ApplyWalk], [TransformWalk]
	SetWalkType(walkType WalkType) GrammarConfig

	// SetCombiner sets the combiner function for the grammar
	//
	// Only used if the walk type is [CombineWalk]
	//
	// See: [Combiner]
	SetCombiner(combiner Combiner) GrammarConfig

	// SetApplier sets the applier function for the grammar
	//
	// Only used if the walk type is [ApplyWalk]
	//
	// See: [Applier]
	SetApplier(applier Applier) GrammarConfig

	// SetTransformer sets the transformer function for the grammar
	//
	// Only used if the walk type is [TransformWalk]
	//
	// See: [Transformer]
	SetTransformer(transformer Transformer) GrammarConfig

	// SetOpArity sets the operation arity for the grammar
	//
	// e.g., [OpArityUnary], [OpArityVariadic]
	SetOpArity(arity OpArity) GrammarConfig

	// SetMaxOperations sets the maximum number of operations
	// allowed per field for the grammar
	//
	// e.g., 1, 2, 3, ...
	SetMaxOperations(maxOps uint8) GrammarConfig

	// SetMultiOpStrategy sets the multi-operation strategy
	// for the grammar
	//
	// e.g., [MultiOpStrategyOrdered], [MultiOpStrategyUnordered]
	SetMultiOpStrategy(strategy MultiOpStrategy) GrammarConfig

	// SetFlatStructure sets the grammar structure to flat
	//
	// Returns a [FlatGrammarConfig] for further flat-specific
	// configuration.
	//
	// e.g., `<grammar_key>:<op_str1>,<op_str2>,...`
	SetFlatStructure() FlatGrammarConfig

	// SetHierarchyStructure sets the grammar structure to hierarchical
	//
	// Returns a [HierarchyGrammarConfig] for further
	// hierarchy-specific configuration.
	//
	// e.g., `<grammar_key>: { \"operations\": [ {...}, {...} ] }`

	SetHierarchyStructure() HierarchyGrammarConfig
	// AddSharedModifier adds a shared modifier specification
	// to the grammar
	//
	// Shared modifiers can be used with any operation in the grammar
	//
	// e.g., `omitempty`, `default`, etc.
	AddSharedModifier(modkey string, fmt ModType, use ModUse, kind ModKind) GrammarConfig

	// SetKVModifierDelim sets the key-value delimiter for
	// key-value modifiers in the grammar
	//
	// e.g., `=`, `:`, `;`, etc.
	SetKVModifierDelim(delim ModKVDelim) GrammarConfig

	// AddOperation adds a new operation to the grammar
	//
	// This is the main way to define the operations that the grammar
	// supports.
	//
	// e.g., `bind_http_json`, `mask_email`, `validate_int`, etc.
	AddOperation(opkey string) GrammarConfig

	// AddOperationModifier adds a modifier for only a specific operation
	//
	// This allows for operation-specific modifier specifications,
	// overriding shared modifiers if necessary.
	//
	// e.g., for operation `bind_http`, modifier `default` may be
	// a string, but for operation `bind_int`, modifier `default`
	// may be an integer.
	AddOperationModifier(opkey string, modkey string, fmt ModType, use ModUse, kind ModKind, override bool) GrammarConfig

	// SetExecFSM sets the execution FSM for the grammar
	//
	// This allows for custom execution control during grammar
	// execution.
	//
	// See: [WalkStateMachine]
	SetExecFSM(fsm WalkStateMachine) GrammarConfig

	// WrapExecFSM wraps the existing execution FSM with
	// the provided wrapper FSM
	//
	// If callBefore is true, the wrapper is called before
	// the existing FSM. Otherwise, it is called after.
	//
	// Warning: This can be called multiple times, resulting
	// in a chain of FSMs. Be careful to avoid infinite loops.
	//
	// See: [WalkStateMachine]
	WrapExecFSM(wrapper WalkStateMachine, callBefore bool) GrammarConfig

	// Build builds the grammar from the configuration
	//
	// Returns the built [Grammar] or an error if the
	// configuration is invalid.
	Build() (Grammar, error)
}

type FlatGrammarConfig interface {
	GrammarConfig
	SetFormat(fmt FlatGrammarFormat, sep FlatGrammarSeparator) FlatGrammarConfig
}

type HierarchyGrammarConfig interface {
	GrammarConfig
	SetFormat(fmt HierarchyGrammarFormat) HierarchyGrammarConfig
}

type grammarConfig struct {
	// Common fields
	key         string
	desc        string
	walkType    WalkType
	combiner    Combiner
	applier     Applier
	transformer Transformer

	// Execution control
	mstrategy MultiOpStrategy
	walkFSM   WalkStateMachine

	// Modifiers
	sharedMods map[string]ModifierSpec
	mkvDelim   ModKVDelim

	// Operations
	opArity    OpArity
	maxOps     uint8
	operations map[string]OperationSpec
}

type flatGrammarConfig struct {
	grammarConfig
	format    FlatGrammarFormat
	separator FlatGrammarSeparator
}

type hierarchyGrammarConfig struct {
	grammarConfig
	format HierarchyGrammarFormat
}

func NewGrammarConfig() GrammarConfig {
	return &grammarConfig{}
}

func (cfg *grammarConfig) SetKey(key string) GrammarConfig {
	cfg.key = key
	return cfg
}

func (cfg *grammarConfig) SetDescription(desc string) GrammarConfig {
	cfg.desc = desc
	return cfg
}

func (cfg *grammarConfig) SetWalkType(walkType WalkType) GrammarConfig {
	cfg.walkType = walkType
	return cfg
}

func (cfg *grammarConfig) SetCombiner(combiner Combiner) GrammarConfig {
	cfg.combiner = combiner
	return cfg
}

func (cfg *grammarConfig) SetApplier(applier Applier) GrammarConfig {
	cfg.applier = applier
	return cfg
}

func (cfg *grammarConfig) SetTransformer(transformer Transformer) GrammarConfig {
	cfg.transformer = transformer
	return cfg
}

func (cfg *grammarConfig) SetOpArity(arity OpArity) GrammarConfig {
	cfg.opArity = arity
	return cfg
}

func (cfg *grammarConfig) SetMultiOpStrategy(strategy MultiOpStrategy) GrammarConfig {
	cfg.mstrategy = strategy
	return cfg
}

func (cfg *grammarConfig) AddSharedModifier(modkey string, fmt ModType, use ModUse, kind ModKind) GrammarConfig {

	if cfg.sharedMods == nil {
		cfg.sharedMods = make(map[string]ModifierSpec)
	}

	spec, exists := cfg.sharedMods[modkey]
	if !exists {
		spec = ModifierSpec{
			modkey: modkey,
			format: fmt,
			use:    use,
			kind:   kind,
		}
	} else {
		spec.format = fmt
		spec.use = use
		spec.kind = kind
	}

	return cfg
}

func (cfg *grammarConfig) SetKVModifierDelim(delim ModKVDelim) GrammarConfig {
	cfg.mkvDelim = delim
	return cfg
}

func (cfg *grammarConfig) AddOperation(opkey string) GrammarConfig {
	if cfg.operations == nil {
		cfg.operations = make(map[string]OperationSpec)
	}

	_, exists := cfg.operations[opkey]
	if !exists {
		opSpec := OperationSpec{
			opkey:    opkey,
			modSpecs: make(map[string]OpModifierSpec),
		}
		cfg.operations[opkey] = opSpec
	}

	return cfg
}

func (cfg *grammarConfig) SetMaxOperations(maxOps uint8) GrammarConfig {
	cfg.maxOps = maxOps
	return cfg
}

func (cfg *grammarConfig) AddOperationModifier(
	opkey string,
	modkey string,
	fmt ModType,
	use ModUse,
	kind ModKind,
	override bool,
) GrammarConfig {
	if cfg.operations == nil {
		cfg.operations = make(map[string]OperationSpec)
	}

	opSpec, exists := cfg.operations[opkey]
	if !exists {
		opSpec = OperationSpec{
			opkey:    opkey,
			modSpecs: make(map[string]OpModifierSpec),
		}
	}

	modSpec, exists := opSpec.modSpecs[modkey]
	if !exists {
		modSpec = OpModifierSpec{
			ModifierSpec: ModifierSpec{
				modkey: modkey,
				use:    use,
				kind:   kind,
				format: fmt,
			},
			overrideShared: override,
		}
	} else {
		modSpec.use = use
		modSpec.kind = kind
		modSpec.format = fmt
		modSpec.overrideShared = override
	}

	opSpec.modSpecs[modkey] = modSpec
	cfg.operations[opkey] = opSpec
	return cfg
}

func (cfg *grammarConfig) SetExecFSM(controller WalkStateMachine) GrammarConfig {
	cfg.walkFSM = controller
	return cfg
}

func (cfg *grammarConfig) WrapExecFSM(walkFSM WalkStateMachine, callBefore bool) GrammarConfig {
	previous := cfg.walkFSM
	if previous == nil {
		previous = func(val any, err error, current ExecState) ExecState {
			return current
		}
	}

	cfg.walkFSM = func(val any, err error, current ExecState) ExecState {
		if callBefore {
			current = walkFSM(val, err, current)
		}

		current = previous(val, err, current)

		if !callBefore {
			current = walkFSM(val, err, current)
		}

		return current
	}

	return cfg
}

func (cfg *grammarConfig) SetFlatStructure() FlatGrammarConfig {
	return &flatGrammarConfig{
		grammarConfig: *cfg,
	}
}

func (cfg *grammarConfig) SetHierarchyStructure() HierarchyGrammarConfig {
	return &hierarchyGrammarConfig{
		grammarConfig: *cfg,
	}
}

func (cfg *grammarConfig) Build() (Grammar, error) {
	return nil, &GrammarBuildError{
		Stage:  StageFinalization,
		Reason: "Cannot build unstructured grammar; must specify structure",
		Value:  nil,
	}
}

func (cfg *flatGrammarConfig) SetFormat(fmt FlatGrammarFormat, sep FlatGrammarSeparator) FlatGrammarConfig {
	cfg.format = fmt
	cfg.separator = sep
	return cfg
}

func (cfg *flatGrammarConfig) validate() error {
	switch cfg.format {
	case FFormatDelimited:
		if len(cfg.separator) != 1 {
			return GrammarBuildError{
				Stage:  StageFormatValidation,
				Reason: "Invalid separator length for delimited flat format (want 1)",
				Value:  len(cfg.separator),
			}
		}
	case FFormatEnclosed:
		if len(cfg.separator) != 2 {
			return GrammarBuildError{
				Stage:  StageFormatValidation,
				Reason: "Invalid separator length for enclosed flat format (want 2)",
				Value:  len(cfg.separator),
			}
		}
	default:
		return GrammarBuildError{
			Stage:  StageFormatValidation,
			Reason: "Unknown flat grammar format",
			Value:  cfg.format,
		}
	}

	return nil
}

func (cfg *flatGrammarConfig) Build() (Grammar, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// $$$TODO $$$SIMON: Implement flat grammar building.
	return nil, nil
}

// $$$TODO $$$SIMON: Implement hierarchy grammar building.
func (cfg *hierarchyGrammarConfig) SetFormat(fmt HierarchyGrammarFormat) HierarchyGrammarConfig {
	panic("not implemented")
}

func (cfg *hierarchyGrammarConfig) validateStructure() error {
	// $$$TODO $$$SIMON: Implement hierarchy grammar format and separator compatibility validation.
	panic("not implemented")
}

func (cfg *hierarchyGrammarConfig) Build() (Grammar, error) {
	// $$$TODO $$$SIMON: Implement hierarchy grammar building.
	panic("not implemented")
}

//--------------------------------------------------------------------------------
// Grammar Templates ($$$SIMON: Are these useful?)
//--------------------------------------------------------------------------------

const (
	DefaultMaxGrammarOps uint8 = 4
)

func NewFlatConfigFromTemplate(template FlatGrammarTemplate) FlatGrammarConfig {
	var (
		flatFmt FlatGrammarFormat
		sep     FlatGrammarSeparator
	)

	switch template {
	case FTemplateInlineComma:
		flatFmt = FFormatDelimited
		sep = FSepInlineCommma
	case FTemplateInlineSemi:
		flatFmt = FFormatDelimited
		sep = FSepInlineSemi
	case FTemplatePairSquare:
		flatFmt = FFormatEnclosed
		sep = FSepPairSquare
	case FTemplatePairCurly:
		flatFmt = FFormatEnclosed
		sep = FSepPairCurl
	case FTemplatePairParen:
		flatFmt = FFormatEnclosed
		sep = FSepPairParen
	default:
		return nil
	}

	return newDefaultGrammarConfig().
		SetFlatStructure().
		SetFormat(flatFmt, sep)
}

func NewHierarchyConfigFromTemplate(template HierarchyGrammarTemplate) HierarchyGrammarConfig {
	switch template {
	case HTemplateJSON:
		return newDefaultGrammarConfig().
			SetHierarchyStructure().
			SetFormat(HFormatJSON)
	default:
		return nil
	}
}

func newDefaultGrammarConfig() GrammarConfig {
	return NewGrammarConfig().
		SetMaxOperations(DefaultMaxGrammarOps).
		AddSharedModifier("omitnil", ModTypeKeyOnly, ModUseExec, ModKindBool).
		AddSharedModifier("omitempty", ModTypeKeyOnly, ModUseExec, ModKindBool).
		AddSharedModifier("omiterr", ModTypeKeyOnly, ModUseExec, ModKindBool).
		AddSharedModifier("default", ModTypeKeyVal, ModUseExec, ModKindConverted)
}
