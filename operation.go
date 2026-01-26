package recipe

// What the hell is an `Operation`?
// Do they:
// 1. Execute on a struct, or
// 2. Execute on a field
// ????
//
// Need to constrain them. If on a struct, should prefer in-place setting (requires passing dest to it).
// If on a field, then need to pass source v

import (
	"fmt"
	"sync"
)

var (
	ErrOpMismatch     = fmt.Errorf("operation type mismatch")
	ErrOpInvalid      = fmt.Errorf("invalid operation")
	ErrOpStratInvalid = fmt.Errorf("invalid operation strategy")
)

type OpMod uint8

// OpArity specifies the arity of an operation.
//
// All Operations within a [Recipe] must have the same arity.
// or will fail out during execution if they do not match.
type OpArity uint8

const (
	// OpUnary takes exactly one source argument
	OpUnary OpArity = iota + 1
	// OpVariadic takes a variable number of source arguments
	OpVariadic = 255
)

type MultiOpStrategy uint8

const (
	// Execute operations in order, stop at first success
	FirstSuccess MultiOpStrategy = iota
	// All operations must succeed
	AllOrNothing
)

type Operation interface {
	Arity() OpArity

	// Execute performs the operation on the provided sources
	//
	// sources must match the arity of the operation.
	//
	// Returns the result of the operation, or an error if execution failed.
	//
	// WARNING: You MUST unpack the sources if a slice when passing to Execute,
	// e.g., op.Execute(opts, sources...) NOT op.Execute(opts, sources)
	Execute(opts *OpOpts, sources ...any) (any, error)
}

type OpOpts struct {
}

// LazyOperation is a reference to an operation with its execution metadata
//
// They are used for lazy resolution to an operation at execution time.
// Prior to the first operation actually executing in a recipe, the Executor
// may be configured to pre-resolve all operation names to ensure they exist.
type LazyOperation struct {
	Name string // e.g., "bind=header", "mask=email", "validate=uuid"
	Opts *OpOpts
}

type ResolvedOperation struct {
	LazyOperation
	Op Operation
}

var (
	ErrOpNotFound = fmt.Errorf("operation not found")
)

type OpRegistry struct {
	mu         sync.RWMutex
	operations map[string]Operation
}

func NewOpRegistry() *OpRegistry {
	return &OpRegistry{
		operations: make(map[string]Operation),
	}
}

func (reg *OpRegistry) RegisterOperation(name string, op Operation) {
	reg.mu.Lock()
	reg.operations[name] = op
	reg.mu.Unlock()
}

func (reg *OpRegistry) resolveOperation(lazyOp LazyOperation) (*ResolvedOperation, error) {
	op, err := reg.getOperation(lazyOp.Name)
	if err != nil {
		return nil, err
	}

	return &ResolvedOperation{
		LazyOperation: lazyOp,
		Op:            op,
	}, nil
}

func (reg *OpRegistry) getOperation(name string) (Operation, error) {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	op, ok := reg.operations[name]
	if !ok {
		return nil, ErrOpNotFound
	}
	return op, nil
}
