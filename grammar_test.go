package recipe

var (
	simpleStringGrammarKey       = "simple_string_grammar"
	simpleBoolCombinerGrammarKey = "simple_bool_grammar"
)

type test_BaseGrammar struct {
	combiner  Combiner
	applier   Applier
	transform Transformer
	key       string
	desc      string
	walkType  WalkType
}

// SplitByOperation: Must use embedding structs to implement
func (tg test_BaseGrammar) SplitByOperation(tag string) ([]string, error) {
	panic("not implemented")
}

// SplitExtras: Must use embedding structs to implement
func (tg test_BaseGrammar) SplitExtras(tag string) (map[string]string, error) {
	panic("not implemented")
}

// ParseOperation: Must use embedding structs to implement
func (tg test_BaseGrammar) ParseOperation(opstr string) (LazyOperation, error) {
	panic("not implemented")
}

// OrderOperations: Must use embedding structs to implement
func (tg test_BaseGrammar) OrderOperations(lazyOps []LazyOperation) ([]LazyOperation, error) {
	panic("not implemented")
}

func (tg test_BaseGrammar) Key() string                       { return tg.key }
func (tg test_BaseGrammar) Description() string               { return tg.desc }
func (tg test_BaseGrammar) WalkType() WalkType                { return tg.walkType }
func (tg test_BaseGrammar) Combiner() (Combiner, error)       { return tg.combiner, nil }
func (tg test_BaseGrammar) Applier() (Applier, error)         { return tg.applier, nil }
func (tg test_BaseGrammar) Transformer() (Transformer, error) { return tg.transform, nil }
func (tg test_BaseGrammar) SetKey(k string)                   { tg.key = k }
func (tg test_BaseGrammar) SetDescription(d string)           { tg.desc = d }
func (tg test_BaseGrammar) SetWalkType(w WalkType)            { tg.walkType = w }
func (tg test_BaseGrammar) SetCombiner(c Combiner)            { tg.combiner = c }
func (tg test_BaseGrammar) SetApplier(a Applier)              { tg.applier = a }
func (tg test_BaseGrammar) SetTransformer(t Transformer)      { tg.transform = t }

// Simple String Combiner Grammar:
//   - tag key = "simple_string_grammar"
//   - Only one operation per tag
//   - No options ([FirstSuccess] strategy only)
type test_SimpleStringGrammar struct{ test_BaseGrammar }

// Simple Grammar:
//   - tag key = "simple_bool_grammar"
//   - Only one operation per tag
//   - No options ([FirstSuccess] strategy only)
type test_SimpleBoolGrammar struct{ test_BaseGrammar }

// Simple Grammar:
//   - tag key = "simple_self_field_grammar"
//   - Only one operation per tag
//   - No options ([FirstSuccess] strategy only)
type test_SimpleSelfFieldGrammar struct{ test_BaseGrammar }
