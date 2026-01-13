package recipe

// TODO $$$SIMON
// 1. Differentiation between anon and explicit struct fields

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

var (
	ErrNoCachedRecipe   = fmt.Errorf("no cached recipe found")
	ErrNotStructKind    = fmt.Errorf("provided type is not a struct")
	ErrEmptyTag         = fmt.Errorf("empty tag")
	ErrExpectedStruct   = fmt.Errorf("expected struct type but got field type")
	ErrUnexpectedStruct = fmt.Errorf("unexpected struct type for non-struct field")
)

type Builder struct {
	grammar Grammar
	mu      sync.RWMutex
	cache   map[reflect.Type]*Recipe
}

func NewBuilder(grammar Grammar) *Builder {
	return &Builder{
		grammar: grammar,
		mu:      sync.RWMutex{},
		cache:   make(map[reflect.Type]*Recipe),
	}
}

// Set manually sets a recipe in the cache
//
// t must be the element type (not pointer) of the struct
// for which the recipe is to be set.
func (b *Builder) Set(t reflect.Type, recipe *Recipe) error {
	if t.Kind() != reflect.Struct {
		return ErrNotStructKind
	}

	b.mu.Lock()
	b.cache[t] = recipe
	b.mu.Unlock()

	return nil
}

// GetOrBuild retrieves a cached Recipe for the given struct type t,
// or builds it if not found.
//
// t must be a struct kind, not a pointer to a struct,
// for which the recipe is to be retrieved or built.
func (b *Builder) GetOrBuild(t reflect.Type) (*Recipe, error) {
	if t.Kind() != reflect.Struct {
		return nil, ErrNotStructKind
	}

	b.mu.RLock()
	recipe, ok := b.cache[t]
	b.mu.RUnlock()

	if ok {
		return recipe, nil
	}

	return b.Build(t, true)
}

// Build constructs a Recipe for the given struct type t.
//
// t is assumed to be a struct kind, not a pointer to a struct,
// for which the recipe is to be built.
//
// Caches built recipe if requested.
func (b *Builder) Build(t reflect.Type, cache bool) (*Recipe, error) {
	recipe, err := b.buildRecipe(t)
	if err != nil {
		return nil, err
	}

	if cache {
		b.mu.Lock()
		b.cache[t] = recipe
		b.mu.Unlock()
	}

	return recipe, nil
}

func (b *Builder) buildRecipe(t reflect.Type) (*Recipe, error) {
	etree, err := b.buildTree(t)
	if err != nil {
		return nil, fmt.Errorf("building struct recipe for type %s: %w", t.Name(), err)
	}

	recipe := &Recipe{
		Root:     etree,
		WalkType: b.grammar.WalkType(),
		resolved: false,
	}

	switch recipe.WalkType {
	case CombineWalk:
		combiner, err := b.grammar.Combiner()
		if err != nil {
			return nil, fmt.Errorf("getting combiner for walk type %d: %w", recipe.WalkType, err)
		}
		recipe.combiner = combiner
	case ApplyWalk:
		applier, err := b.grammar.Applier()
		if err != nil {
			return nil, fmt.Errorf("getting applier for walk type %d: %w", recipe.WalkType, err)
		}
		recipe.applier = applier
	case TransformWalk:
		transformer, err := b.grammar.Transformer()
		if err != nil {
			return nil, fmt.Errorf("getting transformer for walk type %d: %w", recipe.WalkType, err)
		}
		recipe.transformer = transformer
	}

	return recipe, nil
}

func (b *Builder) buildTree(t reflect.Type) (*ExecTree, error) {

	// Assume struct, iterate fields
	etree := &ExecTree{
		Name:     t.Name(),
		LazyOps:  []LazyOperation{},
		Children: []*ExecTree{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		if field.Type.Kind() == reflect.Struct {
			ctree, err := b.buildTree(field.Type)
			if err != nil {
				return nil, fmt.Errorf("field %s, child exec tree: %w", field.Name, err)
			}

			ctree.Name = field.Name
			ctree.fieldIdx = i
			ctree.fieldType = field.Type
			ctree.fieldOffset = field.Offset
			ctree.fieldKind = field.Type.Kind()

			// Pre-compile struct getter to actual ptr for child struct
			b.compileStructAddressor(ctree)

			etree.Children = append(etree.Children, ctree)
			continue
		}

		ftree, err := b.buildField(field)
		if err != nil {
			return nil, fmt.Errorf("field %s, exec tree: %w", field.Name, err)
		}

		if ftree == nil || errors.Is(err, ErrEmptyTag) {
			continue
		}

		// Execution hot-path metadata optimizations
		ftree.fieldIdx = i
		ftree.fieldType = field.Type
		ftree.fieldOffset = field.Offset
		ftree.fieldKind = field.Type.Kind()

		// Pre-compile field extractor for leaf field from parent ptr
		b.compileFieldExtractor(ftree)

		etree.Children = append(etree.Children, ftree)
	}

	return etree, nil
}

func (b *Builder) buildField(field reflect.StructField) (*ExecTree, error) {
	tag := field.Tag.Get(b.grammar.Key())

	if tag == "" || tag == "-" {
		return nil, ErrEmptyTag
	}

	opStrs, err := b.grammar.SplitByOperation(tag)
	if err != nil {
		return nil, fmt.Errorf("splitting operations for field %s: %w", field.Name, err)
	}

	var lazyOps []LazyOperation
	for _, opStr := range opStrs {
		lazyOp, err := b.grammar.ParseOperation(opStr)
		if err != nil {
			return nil, fmt.Errorf("parsing operation %s for field %s: %w", opStr, field.Name, err)
		}
		lazyOps = append(lazyOps, lazyOp)
	}

	orderedOps, err := b.grammar.OrderOperations(lazyOps)
	if err != nil {
		return nil, fmt.Errorf("ordering operations for field %s: %w", field.Name, err)
	}

	return &ExecTree{
		Name:       field.Name,
		LazyOps:    orderedOps,
		Operations: []ResolvedOperation{},
		Children:   []*ExecTree{},
	}, nil
}

// Struct node - Raw pointer extractor to parent struct to avoid reflect.NewAt.
// Used to get pointer to child struct from parent struct pointer
//
// Make sure etree.isStruct() == true before calling this
func (b *Builder) compileStructAddressor(etree *ExecTree) error {
	offset := etree.fieldOffset
	etree.structAddressor = func(structPtr unsafe.Pointer) unsafe.Pointer {
		return unsafe.Pointer(uintptr(structPtr) + offset)
	}

	return nil
}

// Leaf node - Able to avoid reflect for elementary kinds, fallback to
// reflect for complex kinds
//
// Make sure etree.isStruct() == false before calling this
func (b *Builder) compileFieldExtractor(etree *ExecTree) error {
	offset := etree.fieldOffset
	fieldType := etree.fieldType

	switch etree.fieldKind {
	case reflect.Bool:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*bool)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Int:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*int)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Int8:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*int8)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Int16:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*int16)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Int32:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*int32)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Int64:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*int64)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Uint:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*uint)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Uint8:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*uint8)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Uint16:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*uint16)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Uint32:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*uint32)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Uint64:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*uint64)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Float32:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*float32)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Float64:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*float64)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.String:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*string)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Ptr:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			return *(*unsafe.Pointer)(unsafe.Pointer(uintptr(structPtr) + offset))
		}
	case reflect.Slice:
		// Slice header is 3 words: pointer, len, cap
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			fieldPtr := unsafe.Pointer(uintptr(structPtr) + offset)
			// Use reflect as fallback for complex types
			return reflect.NewAt(fieldType, fieldPtr).Elem().Interface()
		}
	case reflect.Array:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			fieldPtr := unsafe.Pointer(uintptr(structPtr) + offset)
			return reflect.NewAt(fieldType, fieldPtr).Elem().Interface()
		}
	case reflect.Map:
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			// Map is a pointer internally
			fieldPtr := unsafe.Pointer(uintptr(structPtr) + offset)
			return reflect.NewAt(fieldType, fieldPtr).Elem().Interface()
		}
	case reflect.Interface:
		// Interface is two words: type and value
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			fieldPtr := unsafe.Pointer(uintptr(structPtr) + offset)
			return reflect.NewAt(fieldType, fieldPtr).Elem().Interface()
		}
	default:
		// Fallback for any unsupported types
		etree.fieldExtractor = func(structPtr unsafe.Pointer) any {
			fieldPtr := unsafe.Pointer(uintptr(structPtr) + offset)
			return reflect.NewAt(fieldType, fieldPtr).Elem().Interface()
		}
	}

	return nil
}
