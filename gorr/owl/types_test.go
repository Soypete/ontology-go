package owl

import (
	"testing"
)

func TestClassString(t *testing.T) {
	c := Class{URI: "http://example.org/A"}
	s := c.String()
	if s != "http://example.org/A" {
		t.Errorf("Expected http://example.org/A, got %s", s)
	}
}

func TestClassFreeVars(t *testing.T) {
	c := Class{URI: "http://example.org/A"}
	vars := c.FreeVars()
	if len(vars) != 0 {
		t.Errorf("Expected 0 free vars, got %d", len(vars))
	}
}

func TestObjectPropertyString(t *testing.T) {
	p := ObjectProperty{URI: "http://example.org/hasParent"}
	s := p.String()
	if s != "http://example.org/hasParent" {
		t.Errorf("Expected http://example.org/hasParent, got %s", s)
	}
}

func TestObjectPropertyFreeVars(t *testing.T) {
	p := ObjectProperty{URI: "http://example.org/hasParent"}
	s := p.String()
	if s != "http://example.org/hasParent" {
		t.Errorf("Expected URI string, got %s", s)
	}
}

func TestURI_IsThing(t *testing.T) {
	uri := URI("http://www.w3.org/2002/07/owl#Thing")
	if !uri.IsThing() {
		t.Error("Expected IsThing() = true for Thing URI")
	}

	uri2 := URI("http://example.org/A")
	if uri2.IsThing() {
		t.Error("Expected IsThing() = false for non-Thing URI")
	}
}

func TestURI_Nothing(t *testing.T) {
	uri := URI("http://www.w3.org/2002/07/owl#Nothing")
	if !uri.IsNothing() {
		t.Error("Expected IsNothing() = true for Nothing URI")
	}

	uri2 := URI("http://example.org/A")
	if uri2.IsNothing() {
		t.Error("Expected IsNothing() = false for non-Nothing URI")
	}
}

func TestURI_Class(t *testing.T) {
	uri := URI("http://www.w3.org/2002/07/owl#Class")
	if !uri.IsClass() {
		t.Error("Expected IsClass() = true for Class URI")
	}

	uri2 := URI("http://example.org/A")
	if uri2.IsClass() {
		t.Error("Expected IsClass() = false for non-Class URI")
	}
}

func TestObjectIntersectionOfString(t *testing.T) {
	o := ObjectIntersectionOf{
		A: &Class{URI: "http://example.org/A"},
		B: &Class{URI: "http://example.org/B"},
	}
	s := o.String()
	if s != "(http://example.org/A ⊓ http://example.org/B)" {
		t.Errorf("Expected intersection string, got %s", s)
	}
}

func TestObjectUnionOfString(t *testing.T) {
	o := ObjectUnionOf{
		A: &Class{URI: "http://example.org/A"},
		B: &Class{URI: "http://example.org/B"},
	}
	s := o.String()
	if s != "(http://example.org/A ⊔ http://example.org/B)" {
		t.Errorf("Expected union string, got %s", s)
	}
}

func TestObjectSomeValuesFromString(t *testing.T) {
	o := ObjectSomeValuesFrom{
		Property: "http://example.org/hasChild",
		Filler:   &Class{URI: "http://example.org/Person"},
	}
	s := o.String()
	if s != "∃http://example.org/hasChild.http://example.org/Person" {
		t.Errorf("Expected some values from string, got %s", s)
	}
}

func TestObjectHasSelfString(t *testing.T) {
	o := ObjectHasSelf{Property: "http://example.org/hasRelative"}
	s := o.String()
	if s != "∃http://example.org/hasRelative.Self" {
		t.Errorf("Expected has self string, got %s", s)
	}
}

func TestObjectOneOfString(t *testing.T) {
	o := ObjectOneOf{
		Individuals: []URI{"http://example.org/Alice", "http://example.org/Bob"},
	}
	s := o.String()
	if s != "{http://example.org/Alice http://example.org/Bob}" {
		t.Errorf("Expected one of string, got %s", s)
	}
}

func TestObjectOneOfEmpty(t *testing.T) {
	o := ObjectOneOf{Individuals: []URI{}}
	s := o.String()
	if s != "{}" {
		t.Errorf("Expected empty set, got %s", s)
	}
}

func TestObjectIntersectionOfFreeVars(t *testing.T) {
	o := ObjectIntersectionOf{
		A: &Class{URI: "http://example.org/A"},
		B: &Class{URI: "http://example.org/B"},
	}
	vars := o.FreeVars()
	if len(vars) != 0 {
		t.Errorf("Expected 0 free vars, got %d", len(vars))
	}
}

func TestObjectSomeValuesFromFreeVars(t *testing.T) {
	o := ObjectSomeValuesFrom{
		Property: "http://example.org/hasChild",
		Filler:   &Class{URI: "http://example.org/Person"},
	}
	vars := o.FreeVars()
	if len(vars) != 1 {
		t.Errorf("Expected 1 free var, got %d", len(vars))
	}
	if vars[0] != "http://example.org/hasChild" {
		t.Errorf("Expected http://example.org/hasChild, got %s", vars[0])
	}
}

func TestObjectHasSelfFreeVars(t *testing.T) {
	o := ObjectHasSelf{Property: "http://example.org/hasRelative"}
	vars := o.FreeVars()
	if len(vars) != 1 {
		t.Errorf("Expected 1 free var, got %d", len(vars))
	}
	if vars[0] != "http://example.org/hasRelative" {
		t.Errorf("Expected http://example.org/hasRelative, got %s", vars[0])
	}
}

func TestObjectOneOfFreeVars(t *testing.T) {
	o := ObjectOneOf{
		Individuals: []URI{"http://example.org/Alice"},
	}
	vars := o.FreeVars()
	if len(vars) != 0 {
		t.Errorf("Expected 0 free vars, got %d", len(vars))
	}
}
