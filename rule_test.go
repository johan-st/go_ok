package go_ok

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestBasicValidation(t *testing.T) {
	ctx := context.Background()

	// Test basic string validation
	stringRule := All(
		Not(StringEmpty()),
		StringLength(5, 100),
		Test("email-format", func(ctx context.Context, s string) error {
			if s == "" {
				return errors.New("empty")
			}
			return nil
		}),
	)

	result, ok := stringRule.Validate(ctx, "test@example.com")
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Errorf("Expected validation to pass, got: %s", result.Format())
	}

	result, ok = stringRule.Validate(ctx, "")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected validation to fail for empty string")
	}
}

func TestShortCircuit(t *testing.T) {
	ctx := context.Background()

	failRule1 := Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("first failure")
	})
	failRule2 := Test("fail2", func(ctx context.Context, n int) error {
		return errors.New("second failure")
	})
	passRule := Test("pass", func(ctx context.Context, n int) error {
		return nil
	})

	allRule := All(failRule1, failRule2, passRule)
	result, ok := allRule.Validate(ctx, 42)
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected All rule to fail")
	}

	// Check that only first rule failed, others are skipped
	if len(result.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(result.Children))
	}
	if result.Children[0].Status != StatusFail {
		t.Error("Expected first child to fail")
	}
	if result.Children[1].Status != StatusSkip {
		t.Error("Expected second child to be skipped")
	}
	if result.Children[2].Status != StatusSkip {
		t.Error("Expected third child to be skipped")
	}
}

// Test Test rule kind
func TestTestRule(t *testing.T) {
	ctx := context.Background()

	// Test passing rule
	passRule := Test("pass", func(ctx context.Context, s string) error {
		return nil
	})
	result, ok := passRule.Validate(ctx, "test")
	t.Logf("ok: %v, result: %+v", ok, result)
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
	t.Logf("ok: %v, result: %+v", ok, result)
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

// Test All rule kind
func TestAllRule(t *testing.T) {
	ctx := context.Background()

	// Test all passing
	pass1 := Test("pass1", func(ctx context.Context, n int) error { return nil })
	pass2 := Test("pass2", func(ctx context.Context, n int) error { return nil })
	allRule := All(pass1, pass2)
	result, ok := allRule.Validate(ctx, 42)
	t.Logf("ok: %v, result: %+v", ok, result)
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
	t.Logf("ok: %v, result: %+v", ok, result)
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
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected empty All rule to pass")
	}
	if len(result.Children) != 0 {
		t.Error("Expected no children")
	}
}

// Test Any rule kind
func TestAnyRule(t *testing.T) {
	ctx := context.Background()

	// Test first passes (short-circuit)
	pass1 := Test("pass1", func(ctx context.Context, n int) error { return nil })
	fail1 := Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("fails")
	})
	anyRule := Any(pass1, fail1)
	result, ok := anyRule.Validate(ctx, 42)
	t.Logf("ok: %v, result: %+v", ok, result)
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
	t.Logf("ok: %v, result: %+v", ok, result)
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
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected empty Any rule to fail")
	}
	if len(result.Children) != 0 {
		t.Error("Expected no children")
	}
}

// Test Not rule kind
func TestNotRule(t *testing.T) {
	ctx := context.Background()

	// Test Not with failing child (should pass)
	failRule := Test("fail", func(ctx context.Context, s string) error {
		return errors.New("fails")
	})
	notRule := Not(failRule)
	result, ok := notRule.Validate(ctx, "test")
	t.Logf("ok: %v, result: %+v", ok, result)
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
	t.Logf("ok: %v, result: %+v", ok, result)
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

// Test Required helper
func TestRequired(t *testing.T) {
	ctx := context.Background()

	// Test string - empty should fail
	requiredStr := Required[string]("name")
	result, ok := requiredStr.Validate(ctx, "")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected Required to fail for empty string")
	}

	// Test string - non-empty should pass
	result, ok = requiredStr.Validate(ctx, "John")
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected Required to pass for non-empty string")
	}

	// Test int - zero should fail
	requiredInt := Required[int]("age")
	result, ok = requiredInt.Validate(ctx, 0)
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected Required to fail for zero int")
	}

	// Test int - non-zero should pass
	result, ok = requiredInt.Validate(ctx, 25)
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected Required to pass for non-zero int")
	}
}

// Test Optional helper
func TestOptional(t *testing.T) {
	ctx := context.Background()

	// Test empty value (should pass)
	optionalRule := Optional(Not(StringEmpty()))
	result, ok := optionalRule.Validate(ctx, "")
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected Optional to pass for empty value")
	}

	// Test non-empty valid value (should pass)
	result, ok = optionalRule.Validate(ctx, "test@example.com")
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected Optional to pass for valid non-empty value")
	}

	// Test non-empty invalid value (should fail)
	result, ok = optionalRule.Validate(ctx, "   ")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected Optional to fail for invalid non-empty value")
	}
}

// Test OneOf helper
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
	result, ok := oneOfRule.Validate(ctx, 1)
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected OneOf to pass when exactly one rule passes")
	}

	// Test none pass
	result, ok = oneOfRule.Validate(ctx, 0)
	t.Logf("ok: %v, result: %+v", ok, result)
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
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected OneOf to fail when multiple rules pass")
	}
	if !strings.Contains(result.Message, "multiple rules passed") {
		t.Errorf("Expected message about multiple passing, got: %s", result.Message)
	}
}

// Test AtLeastN helper
func TestAtLeastN(t *testing.T) {
	ctx := context.Background()

	rule1 := Test("rule1", func(ctx context.Context, n int) error {
		if n > 0 {
			return nil
		}
		return errors.New("not positive")
	})
	rule2 := Test("rule2", func(ctx context.Context, n int) error {
		if n < 100 {
			return nil
		}
		return errors.New("not less than 100")
	})
	rule3 := Test("rule3", func(ctx context.Context, n int) error {
		if n%2 == 0 {
			return nil
		}
		return errors.New("not even")
	})

	// Test at least 2 pass
	atLeast2 := AtLeastN(2, rule1, rule2, rule3)
	result, ok := atLeast2.Validate(ctx, 50) // positive, < 100, even
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected AtLeastN(2) to pass when 3 rules pass")
	}

	// Test exactly 2 pass
	result, ok = atLeast2.Validate(ctx, 51) // positive, < 100, not even
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected AtLeastN(2) to pass when exactly 2 rules pass")
	}

	// Test only 1 passes
	result, ok = atLeast2.Validate(ctx, 101) // positive, not < 100, not even
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected AtLeastN(2) to fail when only 1 rule passes")
	}
	if !strings.Contains(result.Message, "only 1 rules passed") {
		t.Errorf("Expected message about insufficient passes, got: %s", result.Message)
	}
}

// Test StringLength helper
func TestStringLength(t *testing.T) {
	ctx := context.Background()

	lengthRule := StringLength(5, 10)

	// Test valid length
	result, ok := lengthRule.Validate(ctx, "hello")
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected StringLength to pass for valid length")
	}

	result, ok = lengthRule.Validate(ctx, "helloworld")
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected StringLength to pass for max length")
	}

	// Test too short
	result, ok = lengthRule.Validate(ctx, "hi")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected StringLength to fail for too short")
	}
	if !strings.Contains(result.Message, "too short") {
		t.Errorf("Expected message about too short, got: %s", result.Message)
	}

	// Test too long
	result, ok = lengthRule.Validate(ctx, "this is too long")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected StringLength to fail for too long")
	}
	if !strings.Contains(result.Message, "too long") {
		t.Errorf("Expected message about too long, got: %s", result.Message)
	}
}

// Test NotEmpty helper
func TestNotEmpty(t *testing.T) {
	ctx := context.Background()

	notEmptyRule := Not(StringEmpty())

	// Test non-empty
	result, ok := notEmptyRule.Validate(ctx, "hello")
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected NotEmpty to pass for non-empty string")
	}

	// Test empty
	result, ok = notEmptyRule.Validate(ctx, "")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected NotEmpty to fail for empty string")
	}

	// Test whitespace only
	result, ok = notEmptyRule.Validate(ctx, "   ")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected NotEmpty to fail for whitespace-only string")
	}

	result, ok = notEmptyRule.Validate(ctx, "\t\n")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected NotEmpty to fail for whitespace-only string")
	}
}

// Test NumericRange helper
func TestNumericRange(t *testing.T) {
	ctx := context.Background()

	// Test int
	intRule := NumericRange(10, 100)
	result, ok := intRule.Validate(ctx, 50)
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected NumericRange to pass for valid int")
	}

	result, ok = intRule.Validate(ctx, 5)
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected NumericRange to fail for too small int")
	}

	result, ok = intRule.Validate(ctx, 150)
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected NumericRange to fail for too large int")
	}

	// Test int64
	int64Rule := NumericRange(int64(10), int64(100))
	result, ok = int64Rule.Validate(ctx, int64(50))
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected NumericRange to pass for valid int64")
	}

	// Test float64
	floatRule := NumericRange(10.0, 100.0)
	result, ok = floatRule.Validate(ctx, 50.5)
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected NumericRange to pass for valid float64")
	}

	result, ok = floatRule.Validate(ctx, 9.9)
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected NumericRange to fail for too small float64")
	}

	result, ok = floatRule.Validate(ctx, 100.1)
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected NumericRange to fail for too large float64")
	}

	// Test boundary values
	result, ok = intRule.Validate(ctx, 10)
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected NumericRange to pass for min boundary")
	}

	result, ok = intRule.Validate(ctx, 100)
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected NumericRange to pass for max boundary")
	}
}

// Test context cancellation
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	rule := Test("test", func(ctx context.Context, s string) error {
		return nil
	})

	result, ok := rule.Validate(ctx, "test")
	t.Logf("ok: %v, result: %+v", ok, result)
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

// Test context timeout
func TestContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	time.Sleep(time.Millisecond) // Ensure timeout

	rule := Test("test", func(ctx context.Context, s string) error {
		return nil
	})

	result, ok := rule.Validate(ctx, "test")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected validation to fail on timed out context")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
}

// Test nested rules
func TestNestedRules(t *testing.T) {
	ctx := context.Background()

	inner1 := Test("inner1", func(ctx context.Context, s string) error { return nil })
	inner2 := Test("inner2", func(ctx context.Context, s string) error { return nil })
	innerAll := All(inner1, inner2)

	outer1 := Test("outer1", func(ctx context.Context, s string) error { return nil })
	outerAll := All(outer1, innerAll)

	result, ok := outerAll.Validate(ctx, "test")
	t.Logf("ok: %v, result: %+v", ok, result)
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

// Test complex nested Any/All combination
func TestComplexNestedRules(t *testing.T) {
	ctx := context.Background()

	pass1 := Test("pass1", func(ctx context.Context, n int) error { return nil })
	fail1 := Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("fails")
	})

	// (pass1 OR fail1) AND pass1
	anyRule := Any(pass1, fail1)
	allRule := All(anyRule, pass1)

	result, ok := allRule.Validate(ctx, 42)
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected complex nested rule to pass")
	}

	// (fail1 OR fail1) AND pass1
	anyRule = Any(fail1, fail1)
	allRule = All(anyRule, pass1)
	result, ok = allRule.Validate(ctx, 42)
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected complex nested rule to fail when Any fails")
	}
}

// Test Not with nested rules
func TestNotWithNestedRules(t *testing.T) {
	ctx := context.Background()

	pass1 := Test("pass1", func(ctx context.Context, s string) error { return nil })
	pass2 := Test("pass2", func(ctx context.Context, s string) error { return nil })
	allPass := All(pass1, pass2)

	notAll := Not(allPass)
	result, ok := notAll.Validate(ctx, "test")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected Not(All(pass, pass)) to fail")
	}
	if result.Children[0].Status != StatusPass {
		t.Error("Expected inner All to pass")
	}
}

// Test Result.OK() method
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

// Test Result.Format() method
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

// Test Result.String() method
func TestResultString(t *testing.T) {
	result := &Result{
		Status: StatusPass,
		Label:  "test",
	}

	str := result.String()
	if !strings.Contains(str, "PASS") {
		t.Error("Expected String() to contain status")
	}
	if !strings.Contains(str, "test") {
		t.Error("Expected String() to contain label")
	}
}

// Test RuleKind.String()
func TestRuleKindString(t *testing.T) {
	if KindTest.String() != "test" {
		t.Error("Expected KindTest.String() to return 'test'")
	}
	if KindAll.String() != "all" {
		t.Error("Expected KindAll.String() to return 'all'")
	}
	if KindAny.String() != "any" {
		t.Error("Expected KindAny.String() to return 'any'")
	}
	if KindNot.String() != "not" {
		t.Error("Expected KindNot.String() to return 'not'")
	}
}

// Test ResultStatus.String()
func TestResultStatusString(t *testing.T) {
	if StatusPass.String() != "PASS" {
		t.Error("Expected StatusPass.String() to return 'PASS'")
	}
	if StatusFail.String() != "FAIL" {
		t.Error("Expected StatusFail.String() to return 'FAIL'")
	}
	if StatusSkip.String() != "SKIP" {
		t.Error("Expected StatusSkip.String() to return 'SKIP'")
	}
}

// Test Rule.String() method
func TestRuleString(t *testing.T) {
	testRule := Test("test", func(ctx context.Context, s string) error { return nil })
	str := testRule.String()
	if !strings.Contains(str, "Test") {
		t.Error("Expected Rule.String() to contain 'Test'")
	}
	if !strings.Contains(str, "test") {
		t.Error("Expected Rule.String() to contain label")
	}

	allRule := All(testRule)
	str = allRule.String()
	if !strings.Contains(str, "all") {
		t.Error("Expected All rule String() to contain 'all'")
	}
}

// Test edge case: Not rule with invalid children count
func TestNotRuleInvalidChildren(t *testing.T) {
	ctx := context.Background()

	// Create a Not rule manually with wrong number of children
	notRule := &Rule[string]{
		Label:    "not",
		Kind:     KindNot,
		Children: []*Rule[string]{}, // Empty, should be exactly 1
	}

	result, ok := notRule.Validate(ctx, "test")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected Not rule with invalid children to fail")
	}
	if !strings.Contains(result.Message, "exactly one child") {
		t.Errorf("Expected message about invalid children, got: %s", result.Message)
	}
}

// Test edge case: unknown rule kind
func TestUnknownRuleKind(t *testing.T) {
	ctx := context.Background()

	rule := &Rule[string]{
		Label: "unknown",
		Kind:  RuleKind(999), // Invalid kind
	}

	result, ok := rule.Validate(ctx, "test")
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected unknown rule kind to fail")
	}
	if !strings.Contains(result.Message, "unknown rule kind") {
		t.Errorf("Expected message about unknown kind, got: %s", result.Message)
	}
}

// Test different types
func TestDifferentTypes(t *testing.T) {
	ctx := context.Background()

	// Test with custom struct
	type User struct {
		Name string
		Age  int
	}

	userRule := Test("user", func(ctx context.Context, u User) error {
		if u.Name == "" {
			return errors.New("name required")
		}
		return nil
	})

	result, ok := userRule.Validate(ctx, User{Name: "John", Age: 30})
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected user rule to pass")
	}

	result, ok = userRule.Validate(ctx, User{})
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected user rule to fail for empty name")
	}
}

// Test deep nesting
func TestDeepNesting(t *testing.T) {
	ctx := context.Background()

	leaf := Test("leaf", func(ctx context.Context, n int) error { return nil })
	level1 := All(leaf)
	level2 := All(level1)
	level3 := All(level2)
	level4 := All(level3)

	result, ok := level4.Validate(ctx, 42)
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected deeply nested rule to pass")
	}

	// Verify nesting depth
	depth := 0
	current := result
	for len(current.Children) > 0 {
		depth++
		current = current.Children[0]
	}
	if depth != 4 {
		t.Errorf("Expected depth 4, got %d", depth)
	}
}

// Test All with all passing then one failing
func TestAllAllPassThenFail(t *testing.T) {
	ctx := context.Background()

	pass1 := Test("pass1", func(ctx context.Context, n int) error { return nil })
	pass2 := Test("pass2", func(ctx context.Context, n int) error { return nil })
	fail1 := Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("fails")
	})

	allRule := All(pass1, pass2, fail1)
	result, ok := allRule.Validate(ctx, 42)
	t.Logf("ok: %v, result: %+v", ok, result)
	if ok {
		t.Error("Expected All rule to fail")
	}
	if result.Children[0].Status != StatusPass {
		t.Error("Expected first child to pass")
	}
	if result.Children[1].Status != StatusPass {
		t.Error("Expected second child to pass")
	}
	if result.Children[2].Status != StatusFail {
		t.Error("Expected third child to fail")
	}
}

// Test Any with all failing then one passing
func TestAnyAllFailThenPass(t *testing.T) {
	ctx := context.Background()

	fail1 := Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("fails")
	})
	fail2 := Test("fail2", func(ctx context.Context, n int) error {
		return errors.New("fails")
	})
	pass1 := Test("pass1", func(ctx context.Context, n int) error { return nil })

	anyRule := Any(fail1, fail2, pass1)
	result, ok := anyRule.Validate(ctx, 42)
	t.Logf("ok: %v, result: %+v", ok, result)
	if !ok {
		t.Error("Expected Any rule to pass")
	}
	if result.Children[0].Status != StatusFail {
		t.Error("Expected first child to fail")
	}
	if result.Children[1].Status != StatusFail {
		t.Error("Expected second child to fail")
	}
	if result.Children[2].Status != StatusPass {
		t.Error("Expected third child to pass")
	}
}
