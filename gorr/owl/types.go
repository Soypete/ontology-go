package owl

import "fmt"

type ClassExpression interface {
	isClassExpression()
	String() string
	FreeVars() []string
}

type Class struct {
	URI URI
}

func (c *Class) isClassExpression() {}
func (c *Class) String() string     { return string(c.URI) }
func (c *Class) FreeVars() []string { return nil }

type ObjectIntersectionOf struct {
	A ClassExpression
	B ClassExpression
}

func (o *ObjectIntersectionOf) isClassExpression() {}
func (o *ObjectIntersectionOf) String() string {
	return fmt.Sprintf("(%s ⊓ %s)", o.A, o.B)
}
func (o *ObjectIntersectionOf) FreeVars() []string {
	return unionFreeVars(o.A.FreeVars(), o.B.FreeVars())
}

type ObjectUnionOf struct {
	A ClassExpression
	B ClassExpression
}

func (o *ObjectUnionOf) isClassExpression() {}
func (o *ObjectUnionOf) String() string {
	return fmt.Sprintf("(%s ⊔ %s)", o.A, o.B)
}
func (o *ObjectUnionOf) FreeVars() []string {
	return unionFreeVars(o.A.FreeVars(), o.B.FreeVars())
}

type ObjectSomeValuesFrom struct {
	Property URI
	Filler   ClassExpression
}

func (o *ObjectSomeValuesFrom) isClassExpression() {}
func (o *ObjectSomeValuesFrom) String() string {
	return fmt.Sprintf("∃%s.%s", o.Property, o.Filler)
}
func (o *ObjectSomeValuesFrom) FreeVars() []string {
	return unionFreeVars([]string{string(o.Property)}, o.Filler.FreeVars())
}

type ObjectHasSelf struct {
	Property URI
}

func (o *ObjectHasSelf) isClassExpression() {}
func (o *ObjectHasSelf) String() string {
	return fmt.Sprintf("∃%s.Self", o.Property)
}
func (o *ObjectHasSelf) FreeVars() []string { return []string{string(o.Property)} }

type ObjectOneOf struct {
	Individuals []URI
}

func (o *ObjectOneOf) isClassExpression() {}
func (o *ObjectOneOf) String() string {
	if len(o.Individuals) == 0 {
		return "{}"
	}
	result := "{"
	for i, ind := range o.Individuals {
		if i > 0 {
			result += " "
		}
		result += string(ind)
	}
	return result + "}"
}
func (o *ObjectOneOf) FreeVars() []string { return nil }

type URI string

func (i URI) String() string  { return string(i) }
func (i URI) IsClass() bool   { return string(i) == "http://www.w3.org/2002/07/owl#Class" }
func (i URI) IsThing() bool   { return string(i) == "http://www.w3.org/2002/07/owl#Thing" }
func (i URI) IsNothing() bool { return string(i) == "http://www.w3.org/2002/07/owl#Nothing" }

func unionFreeVars(a, b []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, v := range a {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	for _, v := range b {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
