package gorr

import (
	"context"
	"log/slog"

	"github.com/soypete/ontology-go/gorr/owl"
)

type SaturationEngine struct {
	index  *OntologyIndex
	ctxs   map[Handle]*Context
	logger *slog.Logger
}

type SaturationOption func(*SaturationEngine)

func WithLogger(logger *slog.Logger) SaturationOption {
	return func(e *SaturationEngine) {
		if logger != nil {
			e.logger = logger
		}
	}
}

func NewSaturationEngine(index *OntologyIndex, opts ...SaturationOption) *SaturationEngine {
	e := &SaturationEngine{
		index:  index,
		ctxs:   make(map[Handle]*Context),
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *SaturationEngine) Saturation(ctx context.Context) error {
	e.initialize()

	e.logger.Info("starting saturation", "classes", e.index.ClassCount())

	e.saturateAll()

	totalConcs := 0
	for _, ctx := range e.ctxs {
		totalConcs += ctx.SubsumersC.Size() + ctx.SubsumersD.Size()
	}

	e.logger.Info("saturation complete", "conclusions", totalConcs)
	return nil
}

func (e *SaturationEngine) initialize() {
	e.ctxs = make(map[Handle]*Context)

	axioms := e.index.Axioms()
	for _, ax := range axioms {
		switch a := ax.(type) {
		case *owl.SubClassOf:
			subHandle := e.index.internClass(a.SubClass)
			superHandle := e.index.internClass(a.SuperClass)

			e.getOrCreateContext(subHandle).AddSubsumerD(superHandle)
		}
	}
}

func (e *SaturationEngine) getOrCreateContext(root Handle) *Context {
	if ctx, ok := e.ctxs[root]; ok {
		return ctx
	}
	ctx := NewContext(root)
	e.ctxs[root] = ctx
	return ctx
}

func (e *SaturationEngine) saturateAll() {
	changed := true
	for changed {
		changed = false
		for root := range e.ctxs {
			ctx := e.ctxs[root]
			if e.saturateContext(ctx, root) {
				changed = true
			}
		}
	}
}

func (e *SaturationEngine) saturateContext(ctx *Context, root Handle) bool {
	changed := false

	allSubsumers := make([]Handle, 0, 32)
	ctx.SubsumersD.Iterate(func(h Handle) {
		allSubsumers = append(allSubsumers, h)
	})

	for _, target := range allSubsumers {
		subsumers := e.index.GetSubsumers(target)
		subsumers.Iterate(func(super Handle) {
			if ctx.AddSubsumerD(super) {
				changed = true
			}
			e.getOrCreateContext(super).AddSubsumerC(root)
		})

		if domain, ok := e.index.GetPropertyDomain(target); ok {
			if ctx.AddSubsumerC(domain) {
				changed = true
			}
		}

		if rng, ok := e.index.GetPropertyRange(target); ok {
			if ctx.AddSubsumerC(rng) {
				changed = true
			}
		}
	}

	return changed
}

func (e *SaturationEngine) GetContext(root Handle) *Context {
	return e.ctxs[root]
}

func (e *SaturationEngine) IsEntailed(subj, pred, obj Handle) bool {
	ctx, ok := e.ctxs[subj]
	if !ok {
		return false
	}

	switch pred {
	case Handle(1):
		return ctx.HasSubsumerD(obj) || ctx.HasSubsumerC(obj)
	}
	return false
}

func (e *SaturationEngine) GetSubsumers(handle Handle) HandleSet {
	ctx, ok := e.ctxs[handle]
	if !ok {
		return HandleSet{}
	}
	result := NewHandleSet()
	ctx.SubsumersC.Iterate(func(h Handle) {
		result.Add(h)
	})
	ctx.SubsumersD.Iterate(func(h Handle) {
		result.Add(h)
	})
	return *result
}
