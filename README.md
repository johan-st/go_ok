# Unified Rule Tree Validation Framework

A compile-time typed validation framework for Go with full failure tracing, short-circuit evaluation, and type-safe pipelines.

## Features

- **Compile-time type safety**: Full type checking for validation pipelines
- **Short-circuit evaluation**: `All` stops at first failure, `Any` stops at first success
- **Full trace reporting**: Detailed failure trees with `Skip` status for unvisited nodes
- **Context support**: Cancellation and timeouts in predicates
- **Type-narrowing pipelines**: Chain validations with `Then` operations
- **Rich combinators**: `Required`, `Optional`, `OneOf`, `AtLeastN`, etc.
- **Thread-safe**: Immutable rules shareable across goroutines

## Quick Start

```go
package main

import (
    "context"
    "errors"
    "fmt"
)

func main() {
    ctx := context.Background()
    
    // Basic validation
    rule := All(
        NotEmpty("email"),
        StringLength(5, 100),
        Test("email-format", func(ctx context.Context, s string) error {
            if !strings.Contains(s, "@") {
                return errors.New("invalid email format")
            }
            return nil
        }),
    )
    
    result, ok := rule.Validate(ctx, "test@example.com")
    if !ok {
        fmt.Println(result.Format())
    }
    
    // Type-safe pipeline (improved ergonomics)
    // Option 1: Most ergonomic - NotNilThenString helper
    pipeline := NotNilThenString("value",
        NotEmpty("string"),
        StringLength(5, 100),
    )
    
    result, ok = pipeline.Validate(ctx, "hello")
    fmt.Printf("Pipeline result: %v\n", ok)
    
    // Option 2: Using builder pattern for type narrowing (most fluent)
    pipeline2 := ThenFrom(NotNilBuilder("value"), AsString).
        All(NotEmpty("string"), StringLength(5, 100))
    
    // Option 3: Manual Then with All combinator
    pipeline3 := Then(
        NotNil("value"),
        AsString,
        All(NotEmpty("string"), StringLength(5, 100)),
    )
}
```

## Architecture

### Core Types

- `Rule[T]`: Base validation rule with `Children []*Rule[T]` (no type erasure)
- `ThenRule[T,U]`: Type-narrowing pipeline wrapper
- `Result`: Evaluation result with `Status` (Pass/Fail/Skip) and full trace tree

### Rule Kinds

- `KindTest`: Leaf validation node
- `KindAll`: AND combinator (short-circuits on first failure)
- `KindAny`: OR combinator (short-circuits on first success)  
- `KindNot`: Negation

### Performance Features

- **Short-circuit with Skip**: Unvisited nodes marked `StatusSkip`
- **No reflection**: Pure generics and type assertions
- **Immutable rules**: Thread-safe, shareable across goroutines
- **Context-aware**: Cancellation and timeouts in predicates

## API Reference

### Core Constructors

```go
Test[T](label, fn)                    // Leaf test rule
All[T](...rules)                      // AND combinator
Any[T](...rules)                      // OR combinator  
Not[T](rule)                          // Negation
Then[T,U](rule, transform, nextRule)  // Type-narrowing pipeline
ThenFrom[T,U](builder, transform)     // Builder-based type narrowing
ThenString(rule, ...stringRules)       // Convenience: any -> string pipeline
NotNilThenString(label, ...stringRules) // Most ergonomic: NotNil -> string rules
ChainThen[T,U,V](thenRule, transform, nextRule) // Chain pipelines
```

### Helper Functions

```go
Required[T](label)                    // Non-zero value check
Optional[T](rule)                     // Skip validation for zero values
OneOf[T](...rules)                    // Exactly one rule must pass
AtLeastN[T](n, ...rules)              // At least N rules must pass
NotNil(label)                         // Non-nil check for any type
NotNilThenString(label, ...rules)     // Convenience: NotNil -> string validation
IsString()                            // Type check: any -> string
AsString(v)                           // Transform: any -> string
StringLength(min, max)                // String length validation
NotEmpty(label)                       // Non-empty string check
NumericRange[T](min, max)             // Numeric range validation
```

### Evaluation

```go
rule.Validate(ctx, value) -> (*Result, bool)
thenRule.Validate(ctx, value) -> (*Result, bool)
result.OK() -> bool
result.Format() -> string  // Pretty-printed trace tree
```

## Examples

See `example_test.go` for comprehensive examples including:
- Basic string validation
- Type-safe pipelines with `Then`
- User validation with combinators
- Short-circuit behavior
- Context cancellation
- Numeric validation
- Required/Optional patterns

## Running Examples and Tests

```bash
# Run examples
go run cmd/main.go

# Run tests
go test -v
```

## Trade-offs

- **Separate ThenRule type**: Slight API complexity, but enables full compile-time type safety
- **Error returns in predicates**: Slightly more verbose than bool, but enables rich messages
- **No runtime plugins**: Rules are compile-time defined for type safety
