package owl

import "fmt"

// Axiom represents an OWL axiom in the EL profile.
// Each axiom type has a Validate method that checks structural correctness.
type Axiom interface {
	isAxiom()
	Validate() (warnings []string, ok bool)
	String() string
}

// ============================================================================
// Class Expression Axioms
// ============================================================================

// SubClassOf represents rdfs:subClassOf axioms.
// Validates: both SubClass and SuperClass must be non-nil.
type SubClassOf struct {
	SubClass   ClassExpression
	SuperClass ClassExpression
}

func (a *SubClassOf) isAxiom() {}
func (a *SubClassOf) Validate() (warnings []string, ok bool) {
	if a.SubClass == nil {
		return []string{"SubClassOf: SubClass is nil"}, false
	}
	if a.SuperClass == nil {
		return []string{"SubClassOf: SuperClass is nil"}, false
	}
	return nil, true
}
func (a *SubClassOf) String() string {
	return fmt.Sprintf("%s rdfs:subClassOf %s", a.SubClass, a.SuperClass)
}

// EquivalentClasses represents owl:EquivalentClasses axioms.
// Validates: at least 2 class expressions, no nil values.
type EquivalentClasses struct {
	Classes []ClassExpression
}

func (a *EquivalentClasses) isAxiom() {}
func (a *EquivalentClasses) Validate() (warnings []string, ok bool) {
	if len(a.Classes) < 2 {
		return []string{"EquivalentClasses: requires at least 2 classes"}, false
	}
	for i, c := range a.Classes {
		if c == nil {
			return []string{fmt.Sprintf("EquivalentClasses: class[%d] is nil", i)}, false
		}
	}
	return nil, true
}
func (a *EquivalentClasses) String() string {
	return fmt.Sprintf("EquivalentClasses(%v)", a.Classes)
}

// DisjointClasses represents owl:DisjointClasses axioms.
// Validates: at least 2 class expressions, all non-nil.
type DisjointClasses struct {
	Classes []ClassExpression
}

func (a *DisjointClasses) isAxiom() {}
func (a *DisjointClasses) Validate() (warnings []string, ok bool) {
	if len(a.Classes) < 2 {
		return []string{"DisjointClasses: requires at least 2 classes"}, false
	}
	for i, c := range a.Classes {
		if c == nil {
			return []string{fmt.Sprintf("DisjointClasses: class[%d] is nil", i)}, false
		}
	}
	// Warning for odd number - can't be satisfiable
	if len(a.Classes)%2 == 1 {
		warnings = append(warnings, "DisjointClasses: odd number of classes may indicate unsatisfiability")
	}
	return warnings, true
}
func (a *DisjointClasses) String() string {
	return fmt.Sprintf("DisjointClasses(%v)", a.Classes)
}

// ============================================================================
// Property Axioms
// ============================================================================

// SubObjectPropertyOf represents rdfs:subObjectPropertyOf axioms.
// Validates: SubProperty and SuperProperty must be non-nil.
type SubObjectPropertyOf struct {
	SubProperty   PropertyExpression
	SuperProperty PropertyExpression
}

func (a *SubObjectPropertyOf) isAxiom() {}
func (a *SubObjectPropertyOf) Validate() (warnings []string, ok bool) {
	if a.SubProperty == nil {
		return []string{"SubObjectPropertyOf: SubProperty is nil"}, false
	}
	if a.SuperProperty == nil {
		return []string{"SubObjectPropertyOf: SuperProperty is nil"}, false
	}
	return nil, true
}
func (a *SubObjectPropertyOf) String() string {
	return fmt.Sprintf("%s rdfs:subObjectPropertyOf %s", a.SubProperty, a.SuperProperty)
}

// PropertyChain represents owl:propertyChainAxiom (SubPropertyChain).
// Validates: chain has at least 2 properties, all non-nil.
type PropertyChain struct {
	Chain         []PropertyExpression
	SuperProperty PropertyExpression
}

func (a *PropertyChain) isAxiom() {}
func (a *PropertyChain) Validate() (warnings []string, ok bool) {
	if len(a.Chain) < 2 {
		return []string{"PropertyChain: chain requires at least 2 properties"}, false
	}
	for i, p := range a.Chain {
		if p == nil {
			return []string{fmt.Sprintf("PropertyChain: chain[%d] is nil", i)}, false
		}
	}
	if a.SuperProperty == nil {
		return []string{"PropertyChain: SuperProperty is nil"}, false
	}
	return nil, true
}
func (a *PropertyChain) String() string {
	return fmt.Sprintf("%v rdfs:subPropertyOf %s", a.Chain, a.SuperProperty)
}

// FunctionalObjectProperty represents owl:FunctionalObjectProperty.
// No validation needed - just presence.
type FunctionalObjectProperty struct {
	Property PropertyExpression
}

func (a *FunctionalObjectProperty) isAxiom() {}
func (a *FunctionalObjectProperty) Validate() (warnings []string, ok bool) {
	if a.Property == nil {
		return []string{"FunctionalObjectProperty: Property is nil"}, false
	}
	return nil, true
}
func (a *FunctionalObjectProperty) String() string {
	return fmt.Sprintf("FunctionalObjectProperty(%s)", a.Property)
}

// ============================================================================
// Property Expression
// ============================================================================

// PropertyExpression represents an object property or its inverse.
type PropertyExpression interface {
	isPropertyExpression()
	String() string
}

type ObjectProperty struct {
	URI URI
}

func (p *ObjectProperty) isPropertyExpression() {}
func (p *ObjectProperty) String() string        { return string(p.URI) }

type InverseObjectProperty struct {
	URI URI
}

func (p *InverseObjectProperty) isPropertyExpression() {}
func (p *InverseObjectProperty) String() string        { return fmt.Sprintf("inverse(%s)", p.URI) }

// ============================================================================
// Domain and Range
// ============================================================================

// ObjectPropertyDomain represents rdfs:domain axioms for object properties.
// Validates: property and domain class are non-nil.
type ObjectPropertyDomain struct {
	Property PropertyExpression
	Domain   ClassExpression
}

func (a *ObjectPropertyDomain) isAxiom() {}
func (a *ObjectPropertyDomain) Validate() (warnings []string, ok bool) {
	if a.Property == nil {
		return []string{"ObjectPropertyDomain: Property is nil"}, false
	}
	if a.Domain == nil {
		return []string{"ObjectPropertyDomain: Domain is nil"}, false
	}
	return nil, true
}
func (a *ObjectPropertyDomain) String() string {
	return fmt.Sprintf("%s rdfs:domain %s", a.Property, a.Domain)
}

// ObjectPropertyRange represents rdfs:range axioms for object properties.
// Validates: property and range class are non-nil.
type ObjectPropertyRange struct {
	Property PropertyExpression
	Range    ClassExpression
}

func (a *ObjectPropertyRange) isAxiom() {}
func (a *ObjectPropertyRange) Validate() (warnings []string, ok bool) {
	if a.Property == nil {
		return []string{"ObjectPropertyRange: Property is nil"}, false
	}
	if a.Range == nil {
		return []string{"ObjectPropertyRange: Range is nil"}, false
	}
	return nil, true
}
func (a *ObjectPropertyRange) String() string {
	return fmt.Sprintf("%s rdfs:range %s", a.Property, a.Range)
}

// ============================================================================
// Validation Helpers
// ============================================================================

// ValidateAxiom validates any axiom and returns warnings and ok status.
func ValidateAxiom(a Axiom) (warnings []string, ok bool) {
	if a == nil {
		return []string{"Axiom is nil"}, false
	}
	return a.Validate()
}

// ValidateAxioms validates a slice of axioms, collecting all warnings.
func ValidateAxioms(ax []Axiom) (allWarnings []string, ok bool) {
	ok = true
	for i, a := range ax {
		warns, valid := a.Validate()
		if !valid {
			ok = false
			for _, w := range warns {
				allWarnings = append(allWarnings, fmt.Sprintf("[%d] %s", i, w))
			}
		}
		allWarnings = append(allWarnings, warns...)
	}
	return allWarnings, ok
}

// UnsupportedAxiom records an axiom type that is not supported in the EL profile.
type UnsupportedAxiom struct {
	OriginalType string
	Content      string
}

func (a *UnsupportedAxiom) isAxiom() {}
func (a *UnsupportedAxiom) Validate() (warnings []string, ok bool) {
	warnings = append(warnings, fmt.Sprintf("Unsupported axiom type: %s", a.OriginalType))
	return warnings, true
}
func (a *UnsupportedAxiom) String() string {
	if a.Content != "" {
		return a.Content
	}
	return fmt.Sprintf("Unsupported(%s)", a.OriginalType)
}
