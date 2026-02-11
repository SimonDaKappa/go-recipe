# Status
Trying (and failing) to (performantly) defeat the Go Type System via type-erasure closures of Generics. 

Idea is:
1. User code is strongly typed
2. Prefer type-erasure closure that preserves type information through assertions when translating between User and Library execution paths
3. Prefer `Unsafe.Pointer` code with recipe builder pre-"Compiled" type conversion methods where `reflect` is unnavoidable.
  - Will at least prevent swithching over all `reflect.Kind`s and handles "dynamic dispatch" for conversion in `Operation`s of `Walks`
  - `Walks` should be entirely DFS'd in the type-context of a unsafe pointer code. actual typing should only occur in external code

## Markings
**@external:** This code MUST be strongly typed, acceptable by some corresponding generic closure for erasure
  - Translation occurs through construction of the generic erased closure, and its registration with the execution/registry layers

**@internal:** This code must be un-typed but *safe*. Most of this code is in the translation layers. *No* reflect hot-path metadata should be necessary during these translations.
  - Translation occurs through the calling of the generic erased closure that wraps an `@external` function.
 
**@externalreflect:** This code is NOT strongly typed, and must handle at minimum:
  - All `reflect.Kind`s.
  - All interface dispatch/conversion for the desired types (e.g, if parsing from an http request, taking a `gjson.Result` and converting to any elementary value or a simple struct like `time.Time` via `encoding.TextUnmarshaller`

**@internalreflect:** This code must be un-typed and *unsafe*. Most of this code is in the execution layer. Reflect hot-path metadata and pre-"Compiled" utilities are used here
  - Translation occurs primarily through direct `unsafe.Pointer` accesses to fields, and may be re-typed through one of two ways:
    - Reflect based handlers. Directly convert ptr to field to `reflect.Value`, and use Kinds/User provided. These must be pre-compiled at recipe build/resolution time.
    - If possible to implement (I dont think it is): Type erasure closures. **Most of my attempts have failed here**
