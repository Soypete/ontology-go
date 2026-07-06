package gorr

import (
	"testing"
)

func TestExprKindString(t *testing.T) {
	tests := []struct {
		kind     ExprKind
		expected string
	}{
		{ExprClass, "Class"},
		{ExprObjectIntersectionOf, "ObjectIntersectionOf"},
		{ExprObjectSomeValuesFrom, "ObjectSomeValuesFrom"},
		{ExprObjectProperty, "ObjectProperty"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.expected {
				t.Errorf("ExprKind.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExprKindIsClassExpr(t *testing.T) {
	tests := []struct {
		kind     ExprKind
		expected bool
	}{
		{ExprClass, true},
		{ExprObjectIntersectionOf, true},
		{ExprObjectSomeValuesFrom, true},
		{ExprObjectProperty, true},
		{ExprDataProperty, true},
		{100, false},
	}

	for _, tt := range tests {
		name := tt.kind.String()
		if tt.kind >= 100 {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			if got := tt.kind.IsClassExpr(); got != tt.expected {
				t.Errorf("ExprKind.IsClassExpr() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConclusionKindString(t *testing.T) {
	tests := []struct {
		kind     ConclusionKind
		expected string
	}{
		{ConclusionSubsumerC, "SubsumerC"},
		{ConclusionSubsumerD, "SubsumerD"},
		{ConclusionForwardLink, "ForwardLink"},
		{ConclusionInconsistency, "Inconsistency"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.expected {
				t.Errorf("ConclusionKind.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConclusionIsSubsumption(t *testing.T) {
	tests := []struct {
		kind     ConclusionKind
		expected bool
	}{
		{ConclusionSubsumerC, true},
		{ConclusionSubsumerD, true},
		{ConclusionForwardLink, false},
		{ConclusionInconsistency, false},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			if got := tt.kind.IsSubsumption(); got != tt.expected {
				t.Errorf("ConclusionKind.IsSubsumption() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConclusionString(t *testing.T) {
	tests := []struct {
		conclusion Conclusion
		expected   string
	}{
		{
			Conclusion{Kind: ConclusionSubsumerC, Root: Handle(1), Target: Handle(2)},
			"SubsumerC(1 ⊑ 2)",
		},
		{
			Conclusion{Kind: ConclusionSubsumerD, Root: Handle(3), Target: Handle(4)},
			"SubsumerD(3 ⊑ 4)",
		},
		{
			Conclusion{Kind: ConclusionForwardLink, Root: Handle(1), Target: Handle(2)},
			"ForwardLink(1 ->2 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.conclusion.String(); got != tt.expected {
				t.Errorf("Conclusion.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRuleNames(t *testing.T) {
	names := RuleNames()
	if len(names) == 0 {
		t.Error("RuleNames() returned empty map")
	}

	expected := map[int]string{
		RuleInit:                     "Init",
		RuleToldExpansion:            "ToldExpansion",
		RuleConjunctionComposition:   "ConjunctionComposition",
		RuleExistentialDecomposition: "ExistentialDecomposition",
	}

	for num, name := range expected {
		if names[num] != name {
			t.Errorf("RuleNames()[%d] = %v, want %v", num, names[num], name)
		}
	}
}
