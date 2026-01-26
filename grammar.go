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
	// return nil, error if the walk type does not use a combiner
	Combiner() (Combiner, error)

	// Applier applies operation result to a field of the
	// walked struct.
	//
	// return nil, error if the walk type does not use an applier
	Applier() (Applier, error)

	// Transformer transforms the walked struct after all
	// operations have been applied.
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
