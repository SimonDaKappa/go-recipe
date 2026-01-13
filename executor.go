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

	recipe, err := exec.resolveRecipe(wet)
	if err != nil {
		return nil, fmt.Errorf("resolving recipe: %w", err)
	}

	if recipe.WalkType != wt {
		return nil, ErrWalkTypeMismatch
	}

	err = exec.applyContext(ctx, recipe)
	if err != nil {
		return nil, fmt.Errorf("applying exec context: %w", err)
	}

	if len(walked) > 1 {
		for i, walkedArg := range walked {
			if i == 0 {
				continue
			}

			iet, err := exec.elemType(reflect.TypeOf(walkedArg))
			if err != nil {
				return nil, fmt.Errorf("walked argument %d: %w", i, err)
			}

			if wet != iet {
				return nil, fmt.Errorf("walked argument %d: %w", i, ErrWalkArgsMismatch)
			}
		}
	}

	return recipe, nil
}

func (exec *Executor) applyContext(ctx *ExecContext, r *Recipe) error {
	if ctx == nil {
		return nil
	}

	switch r.WalkType {
	case CombineWalk:
		if ctx.CombinerOverride != nil {
			r.combiner = ctx.CombinerOverride
		}
	case ApplyWalk:
		if ctx.ApplierOverride != nil {
			r.applier = ctx.ApplierOverride
		}
	case TransformWalk:
		if ctx.TransformerOverride != nil {
			r.transformer = ctx.TransformerOverride
		}
	}

	return nil
}

// resolveRecipe ensures that the recipe for the given type is built and resolved.
//
// Takes reflect.TypeOf(Walked).Elem() as input, where Walked is a valid pointer to struct.
func (exec *Executor) resolveRecipe(wet reflect.Type) (*Recipe, error) {
	recipe, err := exec.builder.GetOrBuild(wet)
	if err != nil {
		return nil, err
	}

	if !recipe.resolved {
		err := exec.resolveTree(recipe.Root, recipe.Arity)
		if err != nil {
			return nil, fmt.Errorf("resolving exec tree: %w", err)
		}
	}

	recipe.resolved = true

	err = exec.builder.Set(wet, recipe)
	if err != nil {
		return nil, err
	}

	return recipe, nil
}

func (exec *Executor) resolveTree(etree *ExecTree, arity OpArity) error {
	for _, lazyOp := range etree.LazyOps {
		rOp, err := exec.reg.resolveOperation(lazyOp)
		if err != nil {
			return fmt.Errorf("field %s, resolving operation %s: %w", etree.Name, lazyOp.Name, err)
		}

		if rOp.Op.Arity() != arity {
			return fmt.Errorf("field %s, operation %s arity %d does not match recipe arity %d", etree.Name, lazyOp.Name, rOp.Op.Arity(), arity)
		}

		etree.Operations = append(etree.Operations, *rOp)
	}

	for _, child := range etree.Children {
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

func (exec *Executor) extractChildPointers(etree *ExecTree, wlPtrs []unsafe.Pointer) []unsafe.Pointer {
	wcPtrs := make([]unsafe.Pointer, len(wlPtrs))
	for i, wlPtr := range wlPtrs {
		wcPtrs[i] = etree.structAddressor(wlPtr)
	}
	return wcPtrs
}

func (exec *Executor) extractFieldValues(etree *ExecTree, wlPtrs []unsafe.Pointer) []any {
	wlFields := make([]any, len(wlPtrs))
	for i, structPtr := range wlPtrs {
		wlFields[i] = etree.fieldExtractor(structPtr)
	}
	return wlFields
}

//--------------------------------------------------------------------------------
// Combine Walk
//  Performs a combine walk over variadic arity of walked structs,
//  combining results using the provided combiner.
//--------------------------------------------------------------------------------

func (exec *Executor) ExecuteCombineWalk(ctx *ExecContext, walked []any) (any, error) {

	recipe, err := exec.prepareExecute(ctx, CombineWalk, walked)
	if err != nil {
		return nil, fmt.Errorf("preparing combine execute: %w", err)
	}

	wlPointers := make([]unsafe.Pointer, len(walked))
	for i, w := range walked {
		wlPointers[i] = unsafe.Pointer(reflect.ValueOf(w).Pointer())
	}

	acc, err := exec.walkCombiner(recipe.combiner, recipe.Root, wlPointers)
	if err != nil {
		return nil, fmt.Errorf("executing combine walk: %w", err)
	}

	zero := recipe.combiner.Zero()
	acc = recipe.combiner.Combine(zero, acc)

	return acc, nil
}

// walkCombiner is the internal implementation of the combine walk.
//
// wlPtrs: slice of unsafe.Pointer to the current struct level (child of root) being walked.
func (exec *Executor) walkCombiner(combiner Combiner, etree *ExecTree, wlPtrs []unsafe.Pointer) (any, error) {
	acc := combiner.Zero()

	// Struct node
	if etree.hasChild() {
		for _, child := range etree.Children {
			wcPtrs := exec.extractChildPointers(child, wlPtrs)

			res, err := exec.walkCombiner(combiner, child, wcPtrs)
			if err != nil {
				return nil, fmt.Errorf("executing struct child %s: %w", child.Name, err)
			}
			acc = combiner.Combine(acc, res)
		}
		return acc, nil
	}

	// Leaf node
	if etree.hasOperation() {
		wlFields := exec.extractFieldValues(etree, wlPtrs)

		for _, operation := range etree.Operations {
			res, err := operation.Op.Execute(operation.Opts, wlFields...) // Must unpack slice
			if err != nil {
				// Handle opts here
				if false /*opts placehold*/ {

				} else {
					return nil, fmt.Errorf("executing operation %s on field %s: %w", operation.Name, etree.Name, err)
				}
			}

			switch etree.OpStrategy {
			case FirstSuccess:
				acc = res
				return acc, nil
			case AllOrNothing:
				acc = combiner.Combine(acc, res)
			default:
				return nil, fmt.Errorf("unknown multi-op strategy %d", etree.OpStrategy)
			}
		}
	}

	return acc, nil
}

func (exec *Executor) ExecuteApplyWalk(ctx *ExecContext, walked []any, values []any) error {
	recipe, err := exec.prepareExecute(ctx, ApplyWalk, walked)
	if err != nil {
		return fmt.Errorf("preparing apply execute: %w", err)
	}

	wlStructPtrs := make([]unsafe.Pointer, len(walked))
	for i, w := range walked {
		wlStructPtrs[i] = unsafe.Pointer(reflect.ValueOf(w).Pointer())
	}

	return exec.walkApplier(recipe.applier, recipe.Root, wlStructPtrs, values)
}

// walkApplier is the internal implementation of the apply walk.
//
// Applies results of operations to the walked structs using the provided applier.
//
// wlPtrs: slice of unsafe.Pointer to the current struct level (child of root) being walked.
func (exec *Executor) walkApplier(aplr Applier, etree *ExecTree, wlPtrs []unsafe.Pointer, vals []any) error {

	if etree.hasChild() {
		for _, ctree := range etree.Children {
			wcPtrs := exec.extractChildPointers(ctree, wlPtrs)

			err := exec.walkApplier(aplr, ctree, wcPtrs, vals)
			if err != nil {
				return fmt.Errorf("executing struct child %s: %w", ctree.Name, err)
			}
		}
		return nil
	}

	if etree.hasOperation() {
		for _, operation := range etree.Operations {
			res, err := operation.Op.Execute(operation.Opts, vals...)
			if err != nil {
				// Handle opts here
				if false /*opts placehold*/ {

				} else {
					return fmt.Errorf("executing operation %s: %w", operation.Name, err)
				}
			}

			for _, wlPtr := range wlPtrs {
				err := aplr.Apply(wlPtr, etree.fieldOffset, etree.fieldType, res)
				if err != nil {
					return fmt.Errorf("applying result to field %s: %w", etree.Name, err)
				}
			}

			switch etree.OpStrategy {
			case FirstSuccess:
				return nil
			case AllOrNothing:
				continue
			default:
				return fmt.Errorf("unknown multi-op strategy %d", etree.OpStrategy)
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
