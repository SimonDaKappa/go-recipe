package recipe

import (
	"fmt"
	"regexp"
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

type baseGrammarData struct {
	key         string
	description string

	walkType    WalkType
	combiner    Combiner
	applier     Applier
	transformer Transformer
}

func (bgd *baseGrammarData) Key() string {
	return bgd.key
}

func (bgd *baseGrammarData) Description() string {
	return bgd.description
}

func (bgd *baseGrammarData) WalkType() WalkType {
	return bgd.walkType
}

func (bgd *baseGrammarData) Combiner() (Combiner, error) {
	if bgd.walkType == CombineWalk {
		return bgd.combiner, nil
	}
	return nil, fmt.Errorf("grammar walk type %s does not use a combiner", bgd.walkType)
}

func (bgd *baseGrammarData) Applier() (Applier, error) {
	if bgd.walkType == ApplyWalk {
		return bgd.applier, nil
	}
	return nil, fmt.Errorf("grammar walk type %s does not use an applier", bgd.walkType)
}

func (bgd *baseGrammarData) Transformer() (Transformer, error) {
	if bgd.walkType == TransformWalk {
		return bgd.transformer, nil
	}
	return nil, fmt.Errorf("grammar walk type %s does not use a transformer", bgd.walkType)
}

type FlatGrammar struct {
	baseGrammarData

	TagPattern       *regexp.Regexp
	OperationPattern *regexp.Regexp
}

func (fg FlatGrammar) Split(tag string) ([]string, error) {
	// $$$TODO $$$SIMON: Implement flat grammar splitting according to format and separator.
	return nil, nil
}

func (fg FlatGrammar) Parse(opstr string) (LazyOperation, error) {
	// $$$TODO $$$SIMON: Implement flat grammar operation parsing according to operation pattern.
	return LazyOperation{}, nil
}

func (fg FlatGrammar) Order(lazyOps []LazyOperation) ([]LazyOperation, error) {
	// $$$TODO $$$SIMON: Implement flat grammar operation ordering according to grammar rules.
	return nil, nil
}

var (
	__ctc__FlatGrammar_impl_Grammar      Grammar = (*FlatGrammar)(nil)
	__ctc__HierarchyGrammar_impl_Grammar Grammar = (*HierarchyGrammar)(nil)
)

type HierarchyGrammar struct {
	baseGrammarData
}

func (hg HierarchyGrammar) Split(tag string) ([]string, error) {
	// $$$TODO $$$SIMON: Implement hierarchy grammar splitting according to format and separator.
	return nil, nil
}

func (hg HierarchyGrammar) Parse(opstr string) (LazyOperation, error) {
	// $$$TODO $$$SIMON: Implement hierarchy grammar operation parsing according to operation pattern.
	return LazyOperation{}, nil
}

func (hg HierarchyGrammar) Order(lazyOps []LazyOperation) ([]LazyOperation, error) {
	// $$$TODO $$$SIMON: Implement hierarchy grammar operation ordering according to grammar rules.
	return nil, nil
}

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
type FlatGrammarSeparator string

const (
	PairSepSquare      FlatGrammarSeparator = "[]"
	PairSepCurly       FlatGrammarSeparator = "{}"
	PairSepParen       FlatGrammarSeparator = "()"
	InlineSepComma     FlatGrammarSeparator = ","
	InlineSepPipe      FlatGrammarSeparator = "|"
	InlineSepSemicolon FlatGrammarSeparator = ";"
)

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

// $$$TODO $$$SIMON: Add arities to top-level grammar pattern compiler.
type GrammarArity uint8

const (
	// `<grammar_key>:<list_style><op_str><\list_style>`
	GrammarArityUnary GrammarArity = iota + 1

	// `<grammar_key>:<list_style><op_str1>,<op_str2>,...<\list_style>`
	GrammarArityVariadic
)

func (ga GrammarArity) String() string {
	switch ga {
	case GrammarArityUnary:
		return "Unary"
	case GrammarArityVariadic:
		return "Variadic"
	default:
		return "Unknown"
	}
}

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

func (mu ModifierUse) String() string {
	switch mu {
	case ModifierUseExecution:
		return "Execution"
	case ModifierUseOperation:
		return "Operation"
	default:
		return "Unknown"
	}
}

// $$$TODO $$$SIMON: Add modifier pattern compiling to operation pattern compiler.
type ModifierFormat uint8

const (
	// `<modifier_key>=<modifier_value>`
	//
	// Can be used with any [ModifierKind]
	ModFormatKVOnly ModifierFormat = iota + 1

	// `<modifier_key>` (implies boolean true)
	//
	// # Can only be used with [ModKindBool]
	ModFormatKeyOnly

	// Mix of both KV and Key-Only modifiers.
	//
	// Can be used with any [ModifierKind], but Key-Only
	// modifiers must be of type [ModKindBool]
	ModFormatMixed ModifierFormat = 0xFF
)

func (mf ModifierFormat) String() string {
	switch mf {
	case ModFormatKVOnly:
		return "Key-Value"
	case ModFormatKeyOnly:
		return "Key-Only"
	default:
		return "Unknown"
	}
}

// $$$SIMON: Are modifiers allowed to be quoted by builtin grammar builder?
type ModifierKind uint

const (
	// ModKindBool is a string that can be parsed as a boolean.
	// Used with [ModFormatKV] and [ModFormatKeyOnly]
	//
	// e.g., "true", "false", "1", "0", "yes", "no"
	ModKindBool ModifierKind = iota + 1

	// ModKindInt is a string that can be parsed as an integer.
	// Used with [ModFormatKV]
	//
	// e.g., "42", "-1", "0"
	ModKindInt

	// ModKindUInt is a string that can be parsed as an unsigned integer.
	// Used with [ModFormatKV]
	//
	// e.g., "42", "0"
	ModKindUInt

	// ModKindFloat is a string that can be parsed as a float.
	// Used with [ModFormatKV]
	//
	// e.g., "3.14", "-0.001", "0.0"
	ModKindFloat

	// ModKindComplex is a string that can be parsed as a complex number.
	// Used with [ModFormatKV]
	//
	// e.g., "1+2i", "3-4i"
	ModKindComplex

	// ModKindString is any string.
	// Used with [ModFormatKV]
	//
	// e.g., "hello", "world", "any string"
	ModKindString

	// ModKindConverted is a special type indicating the value must be parsed
	// from string to a custom type using the provided ($$$SIMON $$$TODO)
	// reflective conversion functions. If the field implements
	// [encoding.TextUnmarshaler],
	// that will be used. Otherwise, a built-in parser for common types will be used.
	// Used with [ModFormatKV]
	//
	// e.g., custom types like time.Time, UUID, etc.
	ModKindConverted ModifierKind = 0xFF
)

func (mk ModifierKind) String() string {
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
	use    ModifierUse
	kind   ModifierKind
}

type OperationSpec struct {
	opkey    string
	modSpecs map[string]ModifierSpec
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
	SetKey(key string) GrammarConfig
	SetDescription(desc string) GrammarConfig
	SetWalkType(walkType WalkType) GrammarConfig
	SetCombiner(combiner Combiner) GrammarConfig
	SetApplier(applier Applier) GrammarConfig
	SetTransformer(transformer Transformer) GrammarConfig
	SetArity(arity GrammarArity) GrammarConfig
	SetModifierFormat(format ModifierFormat) GrammarConfig
	SetSharedModifier(modkey string, use ModifierUse, kind ModifierKind) GrammarConfig
	SetCustomModifier(opkey string, modkey string, use ModifierUse, kind ModifierKind) GrammarConfig
	SetFlatStructure() FlatGrammarConfig
	SetHierarchyStructure() HierarchyGrammarConfig
}

type FlatGrammarConfig interface {
	GrammarConfig
	SetFormat(fmt FlatGrammarFormat, sep FlatGrammarSeparator) FlatGrammarConfig
	Build() (Grammar, error)
}

type HierarchyGrammarConfig interface {
	GrammarConfig
	SetFormat(fmt HierarchyGrammarFormat) HierarchyGrammarConfig
	Build() (Grammar, error)
}

type grammarConfig struct {
	// Common fields
	key         string
	desc        string
	walkType    WalkType
	combiner    Combiner
	applier     Applier
	transformer Transformer

	arity      GrammarArity
	modformat  ModifierFormat
	sharedMods map[string]ModifierSpec

	// Used only if a specific set of operations need additional
	// specification beyond the default modifier format.
	//
	// Un-Keyed operations use default modifier formats, and
	// function under lazy resolution as usual.
	customOpSpecs map[string]OperationSpec
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

func NewGrammarConfig() *grammarConfig {
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

func (cfg *grammarConfig) SetArity(arity GrammarArity) GrammarConfig {
	cfg.arity = arity
	return cfg
}

func (cfg *grammarConfig) SetModifierFormat(format ModifierFormat) GrammarConfig {
	cfg.modformat = format
	return cfg
}

func (cfg *grammarConfig) SetSharedModifier(modkey string, use ModifierUse, kind ModifierKind) GrammarConfig {

	if cfg.sharedMods == nil {
		cfg.sharedMods = make(map[string]ModifierSpec)
	}

	spec, ok := cfg.sharedMods[modkey]
	if !ok {
		spec = ModifierSpec{
			modkey: modkey,
			use:    use,
			kind:   kind,
		}
	} else {
		spec.use = use
		spec.kind = kind
	}

	return cfg
}

func (cfg *grammarConfig) SetCustomModifier(opkey string, modkey string, use ModifierUse, kind ModifierKind) GrammarConfig {
	if cfg.customOpSpecs == nil {
		cfg.customOpSpecs = make(map[string]OperationSpec)
	}

	opSpec, ok := cfg.customOpSpecs[opkey]
	if !ok {
		opSpec = OperationSpec{
			opkey:    opkey,
			modSpecs: make(map[string]ModifierSpec),
		}
	}

	modSpec, ok := opSpec.modSpecs[modkey]
	if !ok {
		modSpec = ModifierSpec{
			modkey: modkey,
			use:    use,
			kind:   kind,
		}
	} else {
		modSpec.use = use
		modSpec.kind = kind
	}

	opSpec.modSpecs[modkey] = modSpec
	cfg.customOpSpecs[opkey] = opSpec
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
	case FlatFormatDelimited:
		if len(cfg.separator) != 1 {
			return GrammarBuildError{
				Stage:  StageFormatValidation,
				Reason: "Invalid separator length for delimited flat format (want 1)",
				Value:  len(cfg.separator),
			}
		}
	case FlatFormatEnclosed:
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
