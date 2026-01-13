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

type executor struct {
	reg     *OpRegistry
	builder *Builder
}

func newExecutor(registry *OpRegistry, builder *Builder) *executor {
	return &executor{
		reg:     registry,
		builder: builder,
	}
}

func (exec *executor) prepareExecute(ctx *ExecContext, wt WalkType, walked []any) (*Recipe, error) {
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

func (exec *executor) applyContext(ctx *ExecContext, r *Recipe) error {
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

func (exec *executor) resolveRecipe(wt reflect.Type) (*Recipe, error) {
	recipe, err := exec.builder.GetOrBuild(wt)
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

	err = exec.builder.Set(wt, recipe)
	if err != nil {
		return nil, err
	}

	return recipe, nil
}

func (exec *executor) resolveTree(etree *ExecTree, arity OpArity) error {
	for _, lazyOp := range etree.LazyOps {
		rOp, err := exec.reg.resolveOperation(lazyOp)
		if err != nil {
			return fmt.Errorf("field %s, resolving operation %s: %w", etree.Field, lazyOp.Name, err)
		}

		if rOp.Op.Arity() != arity {
			return fmt.Errorf("field %s, operation %s arity %d does not match recipe arity %d", etree.Field, lazyOp.Name, rOp.Op.Arity(), arity)
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

func (exec *executor) validKind(t reflect.Type) error {
	if t.Kind() != reflect.Pointer {
		return ErrNotPointerKind
	}

	if t.Elem().Kind() != reflect.Struct {
		return ErrNotStructElem
	}

	return nil
}

func (exec *executor) elemType(t reflect.Type) (reflect.Type, error) {
	if err := exec.validKind(t); err != nil {
		return nil, err
	}

	return t.Elem(), nil
}

type CombineExecutor struct {
	executor
}

func (exec *CombineExecutor) Execute(ctx *ExecContext, walked []any) (any, error) {

	recipe, err := exec.prepareExecute(ctx, CombineWalk, walked)
	if err != nil {
		return nil, fmt.Errorf("preparing combine execute: %w", err)
	}

	wStructPtrs := make([]unsafe.Pointer, len(walked))
	for i, w := range walked {
		wStructPtrs[i] = unsafe.Pointer(reflect.ValueOf(w).Pointer())
	}

	acc, err := exec.executeWalk(recipe.combiner, recipe.Root, wStructPtrs)
	if err != nil {
		return nil, fmt.Errorf("executing combine walk: %w", err)
	}

	zero := recipe.combiner.Zero()
	acc = recipe.combiner.Combine(zero, acc)

	return acc, nil
}

func (exec *CombineExecutor) executeWalk(combiner Combiner, etree *ExecTree, wStructPtrs []unsafe.Pointer) (any, error) {
	acc := combiner.Zero()

	if etree.isStruct() {
		for _, child := range etree.Children {

			childStructPtrs := make([]unsafe.Pointer, len(wStructPtrs))
			for i, structPtr := range wStructPtrs {
				childStructPtrs[i] = child.structAddressor(structPtr)
			}

			res, err := exec.executeWalk(combiner, child, childStructPtrs)
			if err != nil {
				return nil, fmt.Errorf("executing struct child %s: %w", child.Field, err)
			}
			acc = combiner.Combine(acc, res)
		}
		return acc, nil
	}

	wvFields := make([]any, len(wStructPtrs))
	for i, structPtr := range wStructPtrs {
		wvFields[i] = etree.fieldExtractor(structPtr)
	}

	if etree.hasOperations() {
		for _, operation := range etree.Operations {
			res, err := operation.Op.Execute(operation.Opts, wvFields...) // Must unpack slice
			if err != nil {
				// Handle opts here
				if false /*opts placehold*/ {

				} else {
					return nil, fmt.Errorf("executing operation %s on field %s: %w", operation.Name, etree.Field, err)
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

type ApplyExecutor struct {
	executor
}

func (exec *ApplyExecutor) Execute(ctx *ExecContext, walked []any, faceValues []any) error {
	recipe, err := exec.prepareExecute(ctx, ApplyWalk, walked)
	if err != nil {
		return fmt.Errorf("preparing apply execute: %w", err)
	}

	wPointers := make([]unsafe.Pointer, len(walked))
	for i, w := range walked {
		wPointers[i] = unsafe.Pointer(reflect.ValueOf(w).Pointer())
	}

	return exec.executeWalk(recipe.applier, recipe.Root, wPointers, faceValues)
}

func (exec *ApplyExecutor) executeWalk(applier Applier, etree *ExecTree, wStructPtrs []unsafe.Pointer, values []any) error {

	if etree.isStruct() {
		for _, child := range etree.Children {

			childStructPtrs := make([]unsafe.Pointer, len(wStructPtrs))
			for i, structPtr := range wStructPtrs {
				childStructPtrs[i] = child.structAddressor(structPtr)
			}

			err := exec.executeWalk(applier, child, childStructPtrs, values)
			if err != nil {
				return fmt.Errorf("executing struct child %s: %w", child.Field, err)
			}
		}
		return nil
	}

	if etree.hasOperations() {
		for _, operation := range etree.Operations {
			res, err := operation.Op.Execute(operation.Opts, values...) // Must unpack slice
			if err != nil {
				// Handle opts here
				if false /*opts placehold*/ {

				} else {
					return fmt.Errorf("executing operation %s: %w", operation.Name, err)
				}
			}

			for _, structPtr := range wStructPtrs {
				err := applier.Apply(structPtr, etree.fieldOffset, etree.fieldType, res)
				if err != nil {
					return fmt.Errorf("applying result to field %s: %w", etree.Field, err)
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
