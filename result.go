package gook

import (
	"fmt"
	"strings"
)

// ResultStatus represents the evaluation status of a rule
type ResultStatus int

const (
	StatusPass ResultStatus = iota
	StatusFail
	StatusSkip // for short-circuited branches
)

// String returns a human-readable representation of the result status
func (s ResultStatus) String() string {
	switch s {
	case StatusPass:
		return "PASS"
	case StatusFail:
		return "FAIL"
	case StatusSkip:
		return "SKIP"
	default:
		return "UNKNOWN"
	}
}

// Result represents the evaluation result of a rule
type Result struct {
	Status   ResultStatus
	Label    string
	Kind     RuleKind
	Message  string // formatted at end, not during eval
	Children []*Result
}

// OK returns true if the result represents a successful validation
func (r *Result) OK() bool {
	return r.Status == StatusPass
}

// Format returns a formatted string representation of the result tree
func (r *Result) Format() string {
	var sb strings.Builder
	r.formatRecursive(&sb, 0)
	return sb.String()
}

func (r *Result) formatRecursive(sb *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)

	// Format the current node
	status := r.Status.String()
	if r.Message != "" {
		sb.WriteString(fmt.Sprintf("%s[%s] %s (%s): %s\n",
			indent, status, r.Label, r.Kind.String(), r.Message))
	} else {
		sb.WriteString(fmt.Sprintf("%s[%s] %s (%s)\n",
			indent, status, r.Label, r.Kind.String()))
	}

	// Format children
	for _, child := range r.Children {
		child.formatRecursive(sb, depth+1)
	}
}

// String returns a simple string representation of the result
func (r *Result) String() string {
	return fmt.Sprintf("[%s] %s", r.Status.String(), r.Label)
}
