package recipe

import (
	"fmt"
	"reflect"
	"unsafe"
	// "unsafe"
)

var (
	ErrNotPointerKind   = fmt.Errorf("provided type is not pointer")
	ErrNotStructElem    = fmt.Errorf("provided type's elem is not struct")
	ErrWalkArgsMismatch = fmt.Errorf("walked arguments types do not match")
)

type Executor struct {
	reg     *OpRegistry
	builder *Builder
}

func NewExecutor(registry *OpRegistry, builder *Builder) *Executor {
	return &Executor{
		reg:     registry,
		builder: builder,
	}
}

func (exec *Executor) Execute(ctx *ExecContext, wt WalkType, walked []any, values []any) (any, error) {
	switch wt {
	case CombineWalk:
		return exec.ExecuteCombineWalk(ctx, walked)
	case ApplyWalk:
		return nil, exec.ExecuteApplyWalk(ctx, walked, values)
		// case TransformWalk:
		// 	return nil, exec.ExecuteTransformWalk(ctx, walked)
	default:
		return nil, fmt.Errorf("unknown walk type %d", wt)
	}
}

func (exec *Executor) prepareExecute(ctx *ExecContext, wt WalkType, walked []any) (*Recipe, error) {
	if len(walked) == 0 {
		return nil, fmt.Errorf("no walked arguments provided")
	}

	wet, err := exec.elemType(reflect.TypeOf(walked[0]))
	if err != nil {
		return nil, fmt.Errorf("resolving elem of walked type: %w", err)
	}

	rcp, err := exec.resolveRecipe(wet)
	if err != nil {
		return nil, fmt.Errorf("resolving recipe: %w", err)
	}

	if rcp.WalkType != wt {
		return nil, ErrWalkTypeMismatch
	}

	err = exec.applyContext(ctx, rcp)
	if err != nil {
		return nil, fmt.Errorf("applying exec context: %w", err)
	}

	if len(walked) > 1 {
		for i, w := range walked {
			if i == 0 {
				continue
			}

			iet, err := exec.elemType(reflect.TypeOf(w))
			if err != nil {
				return nil, fmt.Errorf("walked argument %d: %w", i, err)
			}

			if wet != iet {
				return nil, fmt.Errorf("walked argument %d: %w", i, ErrWalkArgsMismatch)
			}
		}
	}

	return rcp, nil
}

func (exec *Executor) applyContext(ctx *ExecContext, rcp *Recipe) error {
	if ctx == nil {
		return nil
	}

	switch rcp.WalkType {
	case CombineWalk:
		if ctx.CombinerOverride != nil {
			rcp.combiner = ctx.CombinerOverride
		}
	case ApplyWalk:
		if ctx.ApplierOverride != nil {
			rcp.applier = ctx.ApplierOverride
		}
	case TransformWalk:
		if ctx.TransformerOverride != nil {
			rcp.transformer = ctx.TransformerOverride
		}
	}

	return nil
}

// resolveRecipe ensures that the recipe for the given type is built and resolved.
//
// Takes reflect.TypeOf(Walked).Elem() as input, where Walked is a valid pointer to struct.
func (exec *Executor) resolveRecipe(wet reflect.Type) (*Recipe, error) {
	rcp, err := exec.builder.GetOrBuild(wet)
	if err != nil {
		return nil, err
	}

	if !rcp.resolved {
		err := exec.resolveTree(rcp.Root, rcp.Arity)
		if err != nil {
			return nil, fmt.Errorf("resolving exec tree: %w", err)
		}
	}

	rcp.resolved = true

	err = exec.builder.Set(wet, rcp)
	if err != nil {
		return nil, err
	}

	return rcp, nil
}

func (exec *Executor) resolveTree(eTree *ExecTree, arity OpArity) error {
	for _, lazyOp := range eTree.LazyOps {
		rOp, err := exec.reg.resolveOperation(lazyOp)
		if err != nil {
			return fmt.Errorf("field %s, resolving operation %s: %w", eTree.Name, lazyOp.Name, err)
		}

		if rOp.Op.Arity() != arity {
			return fmt.Errorf("field %s, operation %s arity %d does not match recipe arity %d", eTree.Name, lazyOp.Name, rOp.Op.Arity(), arity)
		}

		eTree.Operations = append(eTree.Operations, *rOp)
	}

	for _, child := range eTree.Children {
		err := exec.resolveTree(child, arity)
		if err != nil {
			return err
		}
	}

	return nil
}

func (exec *Executor) validKind(t reflect.Type) error {
	if t.Kind() != reflect.Pointer {
		return ErrNotPointerKind
	}

	if t.Elem().Kind() != reflect.Struct {
		return ErrNotStructElem
	}

	return nil
}

func (exec *Executor) elemType(t reflect.Type) (reflect.Type, error) {
	if err := exec.validKind(t); err != nil {
		return nil, err
	}

	return t.Elem(), nil
}

func (exec *Executor) extractChildPointers(eTree *ExecTree, wPtrs []unsafe.Pointer) []unsafe.Pointer {
	cPtrs := make([]unsafe.Pointer, len(wPtrs))
	for i, wPtr := range wPtrs {
		cPtrs[i] = eTree.structAddressor(wPtr)
	}
	return cPtrs
}

func (exec *Executor) extractFieldValues(eTree *ExecTree, wPtrs []unsafe.Pointer) []any {
	wFields := make([]any, len(wPtrs))
	for i, wPtr := range wPtrs {
		wFields[i] = eTree.fieldExtractor(wPtr)
	}
	return wFields
}

//--------------------------------------------------------------------------------
// Combine Walk
//  Performs a combine walk over variadic arity of walked structs,
//  combining results using the provided combiner.
//--------------------------------------------------------------------------------

func (exec *Executor) ExecuteCombineWalk(ctx *ExecContext, walked []any) (any, error) {
	rcp, err := exec.prepareExecute(ctx, CombineWalk, walked)
	if err != nil {
		return nil, fmt.Errorf("preparing combine execute: %w", err)
	}

	wPtrs := make([]unsafe.Pointer, len(walked))
	for i, w := range walked {
		wPtrs[i] = unsafe.Pointer(reflect.ValueOf(w).Pointer())
	}

	acc, err := exec.walkCombiner(rcp.combiner, rcp.Root, wPtrs)
	if err != nil {
		return nil, fmt.Errorf("executing combine walk: %w", err)
	}

	acc = rcp.combiner.Combine(rcp.combiner.Zero(), acc)
	return acc, nil
}

// walkCombiner is the internal implementation of the combine walk.
//
// wlPtrs: slice of unsafe.Pointer to the current struct level (child of root) being walked.
func (exec *Executor) walkCombiner(combiner Combiner, eTree *ExecTree, wPtrs []unsafe.Pointer) (any, error) {
	acc := combiner.Zero()

	// Struct node
	if eTree.hasChild() {
		for _, cTree := range eTree.Children {
			cPtrs := exec.extractChildPointers(cTree, wPtrs)

			res, err := exec.walkCombiner(combiner, cTree, cPtrs)
			if err != nil {
				return nil, fmt.Errorf("executing struct child %s: %w", cTree.Name, err)
			}
			acc = combiner.Combine(acc, res)
		}
		return acc, nil
	}

	// Leaf node
	if eTree.hasOperation() {
		wFields := exec.extractFieldValues(eTree, wPtrs)

		for _, operation := range eTree.Operations {
			res, err := operation.Op.Execute(operation.Opts, wFields...) // Must unpack slice
			if err != nil {
				// Handle opts here
				if false /*opts placehold*/ {

				} else {
					return nil, fmt.Errorf("executing operation %s on field %s: %w", operation.Name, eTree.Name, err)
				}
			}

			switch eTree.OpStrategy {
			case FirstSuccess:
				acc = res
				return acc, nil
			case AllOrNothing:
				acc = combiner.Combine(acc, res)
			default:
				return nil, fmt.Errorf("unknown multi-op strategy %d", eTree.OpStrategy)
			}
		}
	}

	return acc, nil
}

func (exec *Executor) ExecuteApplyWalk(ctx *ExecContext, walked []any, vals []any) error {
	rcp, err := exec.prepareExecute(ctx, ApplyWalk, walked)
	if err != nil {
		return fmt.Errorf("preparing apply execute: %w", err)
	}

	wPtrs := make([]unsafe.Pointer, len(walked))
	for i, w := range walked {
		wPtrs[i] = unsafe.Pointer(reflect.ValueOf(w).Pointer())
	}

	return exec.walkApplier(rcp.applier, rcp.Root, wPtrs, vals)
}

// walkApplier is the internal implementation of the apply walk.
//
// Applies results of operations to the walked structs using the provided applier.
//
// wlPtrs: slice of unsafe.Pointer to the current struct level (child of root) being walked.
func (exec *Executor) walkApplier(aplr Applier, eTree *ExecTree, wPtrs []unsafe.Pointer, vals []any) error {

	if eTree.hasChild() {
		for _, cTree := range eTree.Children {
			cPtrs := exec.extractChildPointers(cTree, wPtrs)

			err := exec.walkApplier(aplr, cTree, cPtrs, vals)
			if err != nil {
				return fmt.Errorf("executing struct child %s: %w", cTree.Name, err)
			}
		}
		return nil
	}

	if eTree.hasOperation() {
		for _, operation := range eTree.Operations {
			res, err := operation.Op.Execute(operation.Opts, vals...)
			if err != nil {
				// Handle opts here
				if false /*opts placehold*/ {

				} else {
					return fmt.Errorf("executing operation %s: %w", operation.Name, err)
				}
			}

			for _, wPtr := range wPtrs {
				err := aplr.Apply(wPtr, eTree.fieldOffset, eTree.fieldType, res)
				if err != nil {
					return fmt.Errorf("applying result to field %s: %w", eTree.Name, err)
				}
			}

			switch eTree.OpStrategy {
			case FirstSuccess:
				return nil
			case AllOrNothing:
				continue
			default:
				return fmt.Errorf("unknown multi-op strategy %d", eTree.OpStrategy)
			}
		}
	}

	return nil
}

// type TransformExecutor struct {
// 	executor
// }

// func (exec *TransformExecutor) ExecuteTransformWalk(ctx *ExecContext, walked []any) error {
// 	_, err := exec.prepareExecute(ctx, TransformWalk, walked)
// 	if err != nil {
// 		return fmt.Errorf("preparing transform execute: %w", err)
// 	}

// 	return nil
// }
