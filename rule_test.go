package gook

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTestRule(t *testing.T) {
	ctx := context.Background()

	// Test passing rule
	passRule := Test("pass", func(ctx context.Context, s string) error {
		return nil
	})
	result, ok := passRule.Validate(ctx, "test")
	if !ok {
		t.Error("Expected test rule to pass")
	}
	if result.Status != StatusPass {
		t.Error("Expected StatusPass")
	}
	if result.Kind != KindTest {
		t.Error("Expected KindTest")
	}

	// Test failing rule
	failRule := Test("fail", func(ctx context.Context, s string) error {
		return errors.New("validation failed")
	})
	result, ok = failRule.Validate(ctx, "test")
	if ok {
		t.Error("Expected test rule to fail")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
	if result.Message != "validation failed" {
		t.Errorf("Expected message 'validation failed', got '%s'", result.Message)
	}
}

func TestAllRule(t *testing.T) {
	ctx := context.Background()

	// Test all passing
	pass1 := Test("pass1", func(ctx context.Context, n int) error { return nil })
	pass2 := Test("pass2", func(ctx context.Context, n int) error { return nil })
	allRule := All(pass1, pass2)
	result, ok := allRule.Validate(ctx, 42)
	if !ok {
		t.Error("Expected All rule to pass when all children pass")
	}
	if result.Status != StatusPass {
		t.Error("Expected StatusPass")
	}
	if len(result.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(result.Children))
	}
	if result.Children[0].Status != StatusPass || result.Children[1].Status != StatusPass {
		t.Error("Expected all children to pass")
	}

	// Test first failure short-circuits
	fail1 := Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("first fails")
	})
	pass3 := Test("pass3", func(ctx context.Context, n int) error { return nil })
	allRule = All(fail1, pass3)
	result, ok = allRule.Validate(ctx, 42)
	if ok {
		t.Error("Expected All rule to fail")
	}
	if result.Children[0].Status != StatusFail {
		t.Error("Expected first child to fail")
	}
	if result.Children[1].Status != StatusSkip {
		t.Error("Expected second child to be skipped")
	}

	// Test empty All rule
	emptyAll := All[int]()
	result, ok = emptyAll.Validate(ctx, 42)
	if !ok {
		t.Error("Expected empty All rule to pass")
	}
	if len(result.Children) != 0 {
		t.Error("Expected no children")
	}
}

func TestAnyRule(t *testing.T) {
	ctx := context.Background()

	// Test first passes (short-circuit)
	pass1 := Test("pass1", func(ctx context.Context, n int) error { return nil })
	fail1 := Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("fails")
	})
	anyRule := Any(pass1, fail1)
	result, ok := anyRule.Validate(ctx, 42)
	if !ok {
		t.Error("Expected Any rule to pass when first child passes")
	}
	if result.Status != StatusPass {
		t.Error("Expected StatusPass")
	}
	if result.Children[0].Status != StatusPass {
		t.Error("Expected first child to pass")
	}
	if result.Children[1].Status != StatusSkip {
		t.Error("Expected second child to be skipped")
	}

	// Test all fail
	fail2 := Test("fail2", func(ctx context.Context, n int) error {
		return errors.New("second fails")
	})
	anyRule = Any(fail1, fail2)
	result, ok = anyRule.Validate(ctx, 42)
	if ok {
		t.Error("Expected Any rule to fail when all children fail")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
	if result.Children[0].Status != StatusFail || result.Children[1].Status != StatusFail {
		t.Error("Expected all children to fail")
	}

	// Test empty Any rule
	emptyAny := Any[int]()
	result, ok = emptyAny.Validate(ctx, 42)
	if ok {
		t.Error("Expected empty Any rule to fail")
	}
	if len(result.Children) != 0 {
		t.Error("Expected no children")
	}
}

func TestNotRule(t *testing.T) {
	ctx := context.Background()

	// Test Not with failing child (should pass)
	failRule := Test("fail", func(ctx context.Context, s string) error {
		return errors.New("fails")
	})
	notRule := Not(failRule)
	result, ok := notRule.Validate(ctx, "test")
	if !ok {
		t.Error("Expected Not rule to pass when child fails")
	}
	if result.Status != StatusPass {
		t.Error("Expected StatusPass")
	}
	if result.Children[0].Status != StatusFail {
		t.Error("Expected child to fail")
	}

	// Test Not with passing child (should fail)
	passRule := Test("pass", func(ctx context.Context, s string) error {
		return nil
	})
	notRule = Not(passRule)
	result, ok = notRule.Validate(ctx, "test")
	if ok {
		t.Error("Expected Not rule to fail when child passes")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
	if result.Message != "not rule failed (child passed)" {
		t.Errorf("Expected message 'not rule failed (child passed)', got '%s'", result.Message)
	}
}

func TestOneOf(t *testing.T) {
	ctx := context.Background()

	rule1 := Test("rule1", func(ctx context.Context, n int) error {
		if n == 1 {
			return nil
		}
		return errors.New("not 1")
	})
	rule2 := Test("rule2", func(ctx context.Context, n int) error {
		if n == 2 {
			return nil
		}
		return errors.New("not 2")
	})
	rule3 := Test("rule3", func(ctx context.Context, n int) error {
		if n == 3 {
			return nil
		}
		return errors.New("not 3")
	})

	oneOfRule := OneOf(rule1, rule2, rule3)

	// Test exactly one passes
	_, ok := oneOfRule.Validate(ctx, 1)
	if !ok {
		t.Error("Expected OneOf to pass when exactly one rule passes")
	}

	// Test none pass
	result, ok := oneOfRule.Validate(ctx, 0)
	if ok {
		t.Error("Expected OneOf to fail when no rules pass")
	}
	if !strings.Contains(result.Message, "none of the rules passed") {
		t.Errorf("Expected message about none passing, got: %s", result.Message)
	}

	// Test multiple pass (should fail)
	rule4 := Test("rule4", func(ctx context.Context, n int) error { return nil })
	rule5 := Test("rule5", func(ctx context.Context, n int) error { return nil })
	oneOfRule = OneOf(rule4, rule5)
	result, ok = oneOfRule.Validate(ctx, 42)
	if ok {
		t.Error("Expected OneOf to fail when multiple rules pass")
	}
	if !strings.Contains(result.Message, "multiple rules passed") {
		t.Errorf("Expected message about multiple passing, got: %s", result.Message)
	}
}

func TestNewRule(t *testing.T) {
	ctx := context.Background()

	// Create rules for NewRule
	rule1 := Test("rule1", func(ctx context.Context, val any) error {
		if val == nil {
			return errors.New("nil value")
		}
		return nil
	})
	rule2 := Test("rule2", func(ctx context.Context, val any) error {
		s, ok := val.(string)
		if !ok {
			return errors.New("not a string")
		}
		if len(s) < 3 {
			return errors.New("too short")
		}
		return nil
	})

	// Test NewRule with multiple rules
	newRule := NewRule("test-rule", rule1, rule2)
	result, ok := newRule.Validate(ctx, "hello")
	if !ok {
		t.Errorf("Expected NewRule to pass, got: %s", result.Format())
	}
	if result.Label != "test-rule" {
		t.Errorf("Expected label 'test-rule', got '%s'", result.Label)
	}
	if result.Kind != KindAll {
		t.Errorf("Expected KindAll, got %v", result.Kind)
	}
	if len(result.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(result.Children))
	}

	// Test NewRule failure
	_, ok = newRule.Validate(ctx, "hi")
	if ok {
		t.Error("Expected NewRule to fail for short string")
	}

	// Test NewRule with nil
	_, ok = newRule.Validate(ctx, nil)
	if ok {
		t.Error("Expected NewRule to fail for nil")
	}
}

func TestAs(t *testing.T) {
	ctx := context.Background()

	// Test As with type assertion (any -> string)
	assertString := func(val any) (string, error) {
		s, ok := val.(string)
		if !ok {
			return "", fmt.Errorf("not a string")
		}
		return s, nil
	}

	stringRule := Test("min-length", func(ctx context.Context, s string) error {
		if len(s) < 3 {
			return errors.New("too short")
		}
		return nil
	})

	asRule := As(assertString, stringRule)

	// Test successful transformation and validation
	result, ok := asRule.Validate(ctx, "hello")
	if !ok {
		t.Errorf("Expected As rule to pass, got: %s", result.Format())
	}

	// Test transformation failure
	result, ok = asRule.Validate(ctx, 42)
	if ok {
		t.Error("Expected As rule to fail for non-string")
	}
	if !strings.Contains(result.Message, "not a string") {
		t.Errorf("Expected transform error, got: %s", result.Message)
	}

	// Test validation failure after successful transformation
	_, ok = asRule.Validate(ctx, "hi")
	if ok {
		t.Error("Expected As rule to fail for short string")
	}
}

func TestAsWithTransformation(t *testing.T) {
	ctx := context.Background()

	// Test As with transformation (string -> int)
	transformToInt := func(val any) (int, error) {
		s, ok := val.(string)
		if !ok {
			return 0, fmt.Errorf("not a string")
		}
		var n int
		_, err := fmt.Sscanf(s, "%d", &n)
		if err != nil {
			return 0, fmt.Errorf("not a number: %v", err)
		}
		return n, nil
	}

	intRule := Test("range", func(ctx context.Context, n int) error {
		if n < 10 || n > 100 {
			return errors.New("out of range")
		}
		return nil
	})

	asRule := As(transformToInt, intRule)

	// Test successful transformation and validation
	result, ok := asRule.Validate(ctx, "50")
	if !ok {
		t.Errorf("Expected As rule to pass, got: %s", result.Format())
	}

	// Test transformation failure (not a string)
	_, ok = asRule.Validate(ctx, 42)
	if ok {
		t.Error("Expected As rule to fail for non-string")
	}

	// Test transformation failure (not a number)
	_, ok = asRule.Validate(ctx, "abc")
	if ok {
		t.Error("Expected As rule to fail for non-numeric string")
	}

	// Test validation failure after successful transformation
	_, ok = asRule.Validate(ctx, "5")
	if ok {
		t.Error("Expected As rule to fail for out of range value")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	rule := Test("test", func(ctx context.Context, s string) error {
		return nil
	})

	result, ok := rule.Validate(ctx, "test")
	if ok {
		t.Error("Expected validation to fail on cancelled context")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
	if result.Message != "context cancelled" {
		t.Errorf("Expected 'context cancelled', got '%s'", result.Message)
	}
}

func TestContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	time.Sleep(time.Millisecond) // Ensure timeout

	rule := Test("test", func(ctx context.Context, s string) error {
		return nil
	})

	result, ok := rule.Validate(ctx, "test")
	if ok {
		t.Error("Expected validation to fail on timed out context")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
}

func TestNestedRules(t *testing.T) {
	ctx := context.Background()

	inner1 := Test("inner1", func(ctx context.Context, s string) error { return nil })
	inner2 := Test("inner2", func(ctx context.Context, s string) error { return nil })
	innerAll := All(inner1, inner2)

	outer1 := Test("outer1", func(ctx context.Context, s string) error { return nil })
	outerAll := All(outer1, innerAll)

	result, ok := outerAll.Validate(ctx, "test")
	if !ok {
		t.Error("Expected nested All rules to pass")
	}
	if len(result.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(result.Children))
	}
	if result.Children[1].Kind != KindAll {
		t.Error("Expected second child to be All")
	}
	if len(result.Children[1].Children) != 2 {
		t.Errorf("Expected nested All to have 2 children, got %d", len(result.Children[1].Children))
	}
}

func TestResultOK(t *testing.T) {
	passResult := &Result{Status: StatusPass}
	if !passResult.OK() {
		t.Error("Expected OK() to return true for StatusPass")
	}

	failResult := &Result{Status: StatusFail}
	if failResult.OK() {
		t.Error("Expected OK() to return false for StatusFail")
	}

	skipResult := &Result{Status: StatusSkip}
	if skipResult.OK() {
		t.Error("Expected OK() to return false for StatusSkip")
	}
}

func TestResultFormat(t *testing.T) {
	result := &Result{
		Status:  StatusFail,
		Label:   "test",
		Kind:    KindTest,
		Message: "error message",
	}

	formatted := result.Format()
	if !strings.Contains(formatted, "FAIL") {
		t.Error("Expected Format() to contain status")
	}
	if !strings.Contains(formatted, "test") {
		t.Error("Expected Format() to contain label")
	}
	if !strings.Contains(formatted, "error message") {
		t.Error("Expected Format() to contain message")
	}

	// Test with children
	result = &Result{
		Status: StatusPass,
		Label:  "all",
		Kind:   KindAll,
		Children: []*Result{
			{Status: StatusPass, Label: "child1", Kind: KindTest},
			{Status: StatusFail, Label: "child2", Kind: KindTest, Message: "failed"},
		},
	}

	formatted = result.Format()
	if !strings.Contains(formatted, "child1") {
		t.Error("Expected Format() to contain child labels")
	}
	if !strings.Contains(formatted, "child2") {
		t.Error("Expected Format() to contain child labels")
	}
}

func TestNotRuleInvalidChildren(t *testing.T) {
	ctx := context.Background()

	// Create a Not rule manually with wrong number of children
	notRule := &Rule[string]{
		Label:    "not",
		Kind:     KindNot,
		Children: []*Rule[string]{}, // Empty, should be exactly 1
	}

	result, ok := notRule.Validate(ctx, "test")
	if ok {
		t.Error("Expected Not rule with invalid children to fail")
	}
	if !strings.Contains(result.Message, "exactly one child") {
		t.Errorf("Expected message about invalid children, got: %s", result.Message)
	}
}

func TestUnknownRuleKind(t *testing.T) {
	ctx := context.Background()

	rule := &Rule[string]{
		Label: "unknown",
		Kind:  RuleKind(999), // Invalid kind
	}

	result, ok := rule.Validate(ctx, "test")
	if ok {
		t.Error("Expected unknown rule kind to fail")
	}
	if !strings.Contains(result.Message, "unknown rule kind") {
		t.Errorf("Expected message about unknown kind, got: %s", result.Message)
	}
}
