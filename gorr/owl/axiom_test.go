package owl

import (
	"testing"
)

func TestSubClassOfValidate(t *testing.T) {
	tests := []struct {
		name     string
		axiom    *SubClassOf
		wantOk   bool
		wantWarn int
	}{
		{
			name: "valid",
			axiom: &SubClassOf{
				SubClass:   &Class{URI: "http://example.org/A"},
				SuperClass: &Class{URI: "http://example.org/B"},
			},
			wantOk:   true,
			wantWarn: 0,
		},
		{
			name: "nil subclass",
			axiom: &SubClassOf{
				SubClass:   nil,
				SuperClass: &Class{URI: "http://example.org/B"},
			},
			wantOk:   false,
			wantWarn: 1,
		},
		{
			name: "nil superclass",
			axiom: &SubClassOf{
				SubClass:   &Class{URI: "http://example.org/A"},
				SuperClass: nil,
			},
			wantOk:   false,
			wantWarn: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, ok := tt.axiom.Validate()
			if ok != tt.wantOk {
				t.Errorf("Validate() ok = %v, want %v", ok, tt.wantOk)
			}
			if len(warnings) != tt.wantWarn {
				t.Errorf("Validate() warnings = %v, want %v", len(warnings), tt.wantWarn)
			}
		})
	}
}

func TestEquivalentClassesValidate(t *testing.T) {
	tests := []struct {
		name     string
		axiom    *EquivalentClasses
		wantOk   bool
		wantWarn int
	}{
		{
			name: "valid two classes",
			axiom: &EquivalentClasses{
				Classes: []ClassExpression{
					&Class{URI: "http://example.org/A"},
					&Class{URI: "http://example.org/B"},
				},
			},
			wantOk:   true,
			wantWarn: 0,
		},
		{
			name: "valid three classes",
			axiom: &EquivalentClasses{
				Classes: []ClassExpression{
					&Class{URI: "http://example.org/A"},
					&Class{URI: "http://example.org/B"},
					&Class{URI: "http://example.org/C"},
				},
			},
			wantOk:   true,
			wantWarn: 0,
		},
		{
			name: "single class",
			axiom: &EquivalentClasses{
				Classes: []ClassExpression{
					&Class{URI: "http://example.org/A"},
				},
			},
			wantOk:   false,
			wantWarn: 1,
		},
		{
			name: "nil class",
			axiom: &EquivalentClasses{
				Classes: []ClassExpression{
					&Class{URI: "http://example.org/A"},
					nil,
				},
			},
			wantOk:   false,
			wantWarn: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, ok := tt.axiom.Validate()
			if ok != tt.wantOk {
				t.Errorf("Validate() ok = %v, want %v", ok, tt.wantOk)
			}
			if len(warnings) != tt.wantWarn {
				t.Errorf("Validate() warnings = %v, want %v", len(warnings), tt.wantWarn)
			}
		})
	}
}

func TestDisjointClassesValidate(t *testing.T) {
	tests := []struct {
		name     string
		axiom    *DisjointClasses
		wantOk   bool
		wantWarn int
	}{
		{
			name: "valid two classes",
			axiom: &DisjointClasses{
				Classes: []ClassExpression{
					&Class{URI: "http://example.org/A"},
					&Class{URI: "http://example.org/B"},
				},
			},
			wantOk:   true,
			wantWarn: 0,
		},
		{
			name: "odd number warning",
			axiom: &DisjointClasses{
				Classes: []ClassExpression{
					&Class{URI: "http://example.org/A"},
					&Class{URI: "http://example.org/B"},
					&Class{URI: "http://example.org/C"},
				},
			},
			wantOk:   true,
			wantWarn: 1,
		},
		{
			name: "single class",
			axiom: &DisjointClasses{
				Classes: []ClassExpression{
					&Class{URI: "http://example.org/A"},
				},
			},
			wantOk:   false,
			wantWarn: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, ok := tt.axiom.Validate()
			if ok != tt.wantOk {
				t.Errorf("Validate() ok = %v, want %v", ok, tt.wantOk)
			}
			if len(warnings) != tt.wantWarn {
				t.Errorf("Validate() warnings = %v, want %v", len(warnings), tt.wantWarn)
			}
		})
	}
}

func TestSubObjectPropertyOfValidate(t *testing.T) {
	tests := []struct {
		name   string
		axiom  *SubObjectPropertyOf
		wantOk bool
	}{
		{
			name: "valid",
			axiom: &SubObjectPropertyOf{
				SubProperty:   &ObjectProperty{URI: "http://example.org/p"},
				SuperProperty: &ObjectProperty{URI: "http://example.org/q"},
			},
			wantOk: true,
		},
		{
			name: "nil subproperty",
			axiom: &SubObjectPropertyOf{
				SubProperty:   nil,
				SuperProperty: &ObjectProperty{URI: "http://example.org/q"},
			},
			wantOk: false,
		},
		{
			name: "nil superproperty",
			axiom: &SubObjectPropertyOf{
				SubProperty:   &ObjectProperty{URI: "http://example.org/p"},
				SuperProperty: nil,
			},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := tt.axiom.Validate()
			if ok != tt.wantOk {
				t.Errorf("Validate() ok = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}

func TestPropertyChainValidate(t *testing.T) {
	tests := []struct {
		name   string
		axiom  *PropertyChain
		wantOk bool
	}{
		{
			name: "valid chain",
			axiom: &PropertyChain{
				Chain: []PropertyExpression{
					&ObjectProperty{URI: "http://example.org/p"},
					&ObjectProperty{URI: "http://example.org/q"},
				},
				SuperProperty: &ObjectProperty{URI: "http://example.org/r"},
			},
			wantOk: true,
		},
		{
			name: "single element chain",
			axiom: &PropertyChain{
				Chain: []PropertyExpression{
					&ObjectProperty{URI: "http://example.org/p"},
				},
				SuperProperty: &ObjectProperty{URI: "http://example.org/r"},
			},
			wantOk: false,
		},
		{
			name: "nil in chain",
			axiom: &PropertyChain{
				Chain: []PropertyExpression{
					&ObjectProperty{URI: "http://example.org/p"},
					nil,
				},
				SuperProperty: &ObjectProperty{URI: "http://example.org/r"},
			},
			wantOk: false,
		},
		{
			name: "nil superproperty",
			axiom: &PropertyChain{
				Chain: []PropertyExpression{
					&ObjectProperty{URI: "http://example.org/p"},
					&ObjectProperty{URI: "http://example.org/q"},
				},
				SuperProperty: nil,
			},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := tt.axiom.Validate()
			if ok != tt.wantOk {
				t.Errorf("Validate() ok = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}

func TestObjectPropertyRangeValidate(t *testing.T) {
	tests := []struct {
		name   string
		axiom  *ObjectPropertyRange
		wantOk bool
	}{
		{
			name: "valid",
			axiom: &ObjectPropertyRange{
				Property: &ObjectProperty{URI: "http://example.org/p"},
				Range:    &Class{URI: "http://example.org/Range"},
			},
			wantOk: true,
		},
		{
			name: "nil property",
			axiom: &ObjectPropertyRange{
				Property: nil,
				Range:    &Class{URI: "http://example.org/Range"},
			},
			wantOk: false,
		},
		{
			name: "nil range",
			axiom: &ObjectPropertyRange{
				Property: &ObjectProperty{URI: "http://example.org/p"},
				Range:    nil,
			},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := tt.axiom.Validate()
			if ok != tt.wantOk {
				t.Errorf("Validate() ok = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}

func TestObjectPropertyDomainValidate(t *testing.T) {
	tests := []struct {
		name   string
		axiom  *ObjectPropertyDomain
		wantOk bool
	}{
		{
			name: "valid",
			axiom: &ObjectPropertyDomain{
				Property: &ObjectProperty{URI: "http://example.org/p"},
				Domain:   &Class{URI: "http://example.org/Domain"},
			},
			wantOk: true,
		},
		{
			name: "nil property",
			axiom: &ObjectPropertyDomain{
				Property: nil,
				Domain:   &Class{URI: "http://example.org/Domain"},
			},
			wantOk: false,
		},
		{
			name: "nil domain",
			axiom: &ObjectPropertyDomain{
				Property: &ObjectProperty{URI: "http://example.org/p"},
				Domain:   nil,
			},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := tt.axiom.Validate()
			if ok != tt.wantOk {
				t.Errorf("Validate() ok = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}

func TestValidateAxiom(t *testing.T) {
	_, ok := ValidateAxiom(nil)
	if ok {
		t.Error("Expected ValidateAxiom(nil) to return false")
	}

	ax := &SubClassOf{
		SubClass:   &Class{URI: "http://example.org/A"},
		SuperClass: &Class{URI: "http://example.org/B"},
	}
	_, ok = ValidateAxiom(ax)
	if !ok {
		t.Error("Expected valid axiom to pass")
	}
}

func TestUnsupportedAxiom(t *testing.T) {
	ua := &UnsupportedAxiom{
		OriginalType: "owl:AllValuesFrom",
		Content:      "∀p.C",
	}

	warnings, ok := ua.Validate()
	if !ok {
		t.Error("UnsupportedAxiom should return ok=true with warning")
	}
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}
	if ua.String() != "∀p.C" {
		t.Errorf("String() = %v, want ∀p.C", ua.String())
	}
}
