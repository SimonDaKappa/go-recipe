package recipe

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

	SplitByOperation(tag string) ([]string, error)
	SplitExtras(tag string) (map[string]string, error)
	ParseOperation(opstr string) (LazyOperation, error)
	OrderOperations(lazyOps []LazyOperation) ([]LazyOperation, error)
}
