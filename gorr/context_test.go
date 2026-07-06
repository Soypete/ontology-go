package gorr

import (
	"testing"
)

func TestHandleSetAdd(t *testing.T) {
	s := NewHandleSet()

	if !s.Add(Handle(1)) {
		t.Error("Add(1) should return true for first add")
	}
	if s.Add(Handle(1)) {
		t.Error("Add(1) should return false for duplicate")
	}
	if !s.Add(Handle(2)) {
		t.Error("Add(2) should return true")
	}
}

func TestHandleSetContains(t *testing.T) {
	s := NewHandleSet()

	s.Add(Handle(1))
	s.Add(Handle(3))

	if !s.Contains(Handle(1)) {
		t.Error("Contains(1) should return true")
	}
	if s.Contains(Handle(2)) {
		t.Error("Contains(2) should return false")
	}
	if !s.Contains(Handle(3)) {
		t.Error("Contains(3) should return true")
	}
}

func TestHandleSetSize(t *testing.T) {
	s := NewHandleSet()

	if s.Size() != 0 {
		t.Errorf("Size() = %v, want 0", s.Size())
	}

	s.Add(Handle(1))
	s.Add(Handle(2))
	s.Add(Handle(3))

	if s.Size() != 3 {
		t.Errorf("Size() = %v, want 3", s.Size())
	}

	s.Add(Handle(1)) // duplicate

	if s.Size() != 3 {
		t.Errorf("Size() after duplicate = %v, want 3", s.Size())
	}
}

func TestHandleSetToSlice(t *testing.T) {
	s := NewHandleSet()
	s.Add(Handle(3))
	s.Add(Handle(1))
	s.Add(Handle(2))

	slice := s.ToSlice()
	if len(slice) != 3 {
		t.Errorf("ToSlice() len = %v, want 3", len(slice))
	}

	// Verify all elements present
	found := make(map[Handle]bool)
	for _, h := range slice {
		found[h] = true
	}

	if !found[Handle(1)] || !found[Handle(2)] || !found[Handle(3)] {
		t.Error("ToSlice() missing elements")
	}
}

func TestNewContext(t *testing.T) {
	ctx := NewContext(Handle(42))

	if ctx.Root != Handle(42) {
		t.Errorf("Root = %v, want 42", ctx.Root)
	}
	if ctx.SubsumersC == nil {
		t.Error("SubsumersC should not be nil")
	}
	if ctx.SubsumersD == nil {
		t.Error("SubsumersD should not be nil")
	}
	if ctx.ForwardLinks == nil {
		t.Error("ForwardLinks should not be nil")
	}
	if ctx.BackwardLinks == nil {
		t.Error("BackwardLinks should not be nil")
	}
}

func TestContextSubsumers(t *testing.T) {
	ctx := NewContext(Handle(1))

	if !ctx.AddSubsumerC(Handle(2)) {
		t.Error("AddSubsumerC(2) should return true")
	}
	if ctx.AddSubsumerC(Handle(2)) {
		t.Error("AddSubsumerC(2) duplicate should return false")
	}
	if !ctx.AddSubsumerD(Handle(3)) {
		t.Error("AddSubsumerD(3) should return true")
	}

	if !ctx.HasSubsumerC(Handle(2)) {
		t.Error("HasSubsumerC(2) should return true")
	}
	if ctx.HasSubsumerC(Handle(3)) {
		t.Error("HasSubsumerC(3) should return false")
	}
	if !ctx.HasSubsumerD(Handle(3)) {
		t.Error("HasSubsumerD(3) should return true")
	}
}

func TestContextLinks(t *testing.T) {
	ctx := NewContext(Handle(1))

	if !ctx.AddForwardLink(Handle(10), Handle(20)) {
		t.Error("AddForwardLink should return true")
	}
	if !ctx.AddBackwardLink(Handle(10), Handle(30)) {
		t.Error("AddBackwardLink should return true")
	}

	if !ctx.HasForwardLink(Handle(10), Handle(20)) {
		t.Error("HasForwardLink should return true")
	}
	if ctx.HasForwardLink(Handle(10), Handle(21)) {
		t.Error("HasForwardLink should return false for missing target")
	}
	if !ctx.HasBackwardLink(Handle(10), Handle(30)) {
		t.Error("HasBackwardLink should return true")
	}
}

func TestContextTodo(t *testing.T) {
	ctx := NewContext(Handle(1))

	if !ctx.TodoEmpty() {
		t.Error("New context should have empty Todo")
	}

	ctx.PushTodo(Conclusion{Kind: ConclusionSubsumerC, Root: Handle(1), Target: Handle(2)})

	if ctx.TodoEmpty() {
		t.Error("Todo should not be empty after push")
	}

	popped := ctx.PopTodo()
	if popped.Kind != ConclusionSubsumerC {
		t.Errorf("PopTodo() = %v, want SubsumerC", popped.Kind)
	}

	if !ctx.TodoEmpty() {
		t.Error("Todo should be empty after pop")
	}

	// Pop from empty
	empty := ctx.PopTodo()
	if empty.Kind != 0 {
		t.Error("PopTodo from empty should return zero conclusion")
	}
}
