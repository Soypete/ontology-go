package gorr

// Context represents a saturation context for a root class expression.
// Each root class has its own context where conclusions are derived.
type Context struct {
	Root          Handle
	SubsumersC    *HandleSet
	SubsumersD    *HandleSet
	ForwardLinks  map[Handle]*HandleSet
	BackwardLinks map[Handle]*HandleSet
	Disjoint      map[Handle]uint8
	Todo          []Conclusion
	Activated     bool
	Saturated     bool
}

// NewContext creates a new context for the given root.
func NewContext(root Handle) *Context {
	return &Context{
		Root:          root,
		SubsumersC:    NewHandleSet(),
		SubsumersD:    NewHandleSet(),
		ForwardLinks:  make(map[Handle]*HandleSet),
		BackwardLinks: make(map[Handle]*HandleSet),
		Disjoint:      make(map[Handle]uint8),
		Todo:          nil,
		Activated:     false,
		Saturated:     false,
	}
}

// AddSubsumerC adds a composed subsumer (forward link derived).
func (c *Context) AddSubsumerC(h Handle) bool {
	return c.SubsumersC.Add(h)
}

// AddSubsumerD adds a decomposed subsumer (direct link).
func (c *Context) AddSubsumerD(h Handle) bool {
	return c.SubsumersD.Add(h)
}

// AddForwardLink adds a forward link to the context.
func (c *Context) AddForwardLink(prop, target Handle) bool {
	if c.ForwardLinks[prop] == nil {
		c.ForwardLinks[prop] = NewHandleSet()
	}
	return c.ForwardLinks[prop].Add(target)
}

// AddBackwardLink adds a backward link to the context.
func (c *Context) AddBackwardLink(prop, source Handle) bool {
	if c.BackwardLinks[prop] == nil {
		c.BackwardLinks[prop] = NewHandleSet()
	}
	return c.BackwardLinks[prop].Add(source)
}

// HasSubsumerC returns true if the context has the given composed subsumer.
func (c *Context) HasSubsumerC(h Handle) bool {
	return c.SubsumersC.Contains(h)
}

// HasSubsumerD returns true if the context has the given decomposed subsumer.
func (c *Context) HasSubsumerD(h Handle) bool {
	return c.SubsumersD.Contains(h)
}

// HasForwardLink returns true if the context has a forward link.
func (c *Context) HasForwardLink(prop, target Handle) bool {
	if c.ForwardLinks[prop] == nil {
		return false
	}
	return c.ForwardLinks[prop].Contains(target)
}

// HasBackwardLink returns true if the context has a backward link.
func (c *Context) HasBackwardLink(prop, source Handle) bool {
	if c.BackwardLinks[prop] == nil {
		return false
	}
	return c.BackwardLinks[prop].Contains(source)
}

// PushTodo adds a conclusion to the ToDo queue.
func (c *Context) PushTodo(conclusion Conclusion) {
	c.Todo = append(c.Todo, conclusion)
}

// PopTodo removes and returns the next conclusion from the ToDo queue.
func (c *Context) PopTodo() Conclusion {
	if len(c.Todo) == 0 {
		return Conclusion{}
	}
	conclusion := c.Todo[0]
	c.Todo = c.Todo[1:]
	return conclusion
}

// TodoEmpty returns true if the ToDo queue is empty.
func (c *Context) TodoEmpty() bool {
	return len(c.Todo) == 0
}
