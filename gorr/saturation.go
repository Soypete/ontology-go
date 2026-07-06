package gorr

import (
	"context"
	"log/slog"

	"github.com/soypete/ontology-go/gorr/owl"
)

type SaturationEngine struct {
	index   *OntologyIndex
	ctxs    map[Handle]*Context
	logger  *slog.Logger
	workers int
}

type SaturationOption func(*SaturationEngine)

func WithLogger(logger *slog.Logger) SaturationOption {
	return func(e *SaturationEngine) {
		e.logger = logger
	}
}

func WithWorkers(workers int) SaturationOption {
	return func(e *SaturationEngine) {
		e.workers = workers
	}
}

func NewSaturationEngine(index *OntologyIndex, opts ...SaturationOption) *SaturationEngine {
	e := &SaturationEngine{
		index:   index,
		ctxs:    make(map[Handle]*Context),
		workers: 4,
		logger:  slog.Default(),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *SaturationEngine) Saturation(ctx context.Context) error {
	e.initialize()

	e.logger.Info("starting saturation", "classes", e.index.ClassCount())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		anyProgress := false

		for root, ctx := range e.ctxs {
			if !ctx.Saturated {
				progress := e.processContext(ctx, root)
				if progress {
					anyProgress = true
				} else {
					ctx.Saturated = true
				}
			}
		}

		if !anyProgress {
			break
		}
	}

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

			ctx := e.getOrCreateContext(subHandle)
			ctx.AddSubsumerD(superHandle)
			ctx.PushTodo(Conclusion{
				Kind:   ConclusionSubsumerD,
				Root:   subHandle,
				Target: superHandle,
			})

			if subHandle != superHandle {
				ctx2 := e.getOrCreateContext(superHandle)
				ctx2.AddSubsumerC(subHandle)
			}
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

func (e *SaturationEngine) processContext(ctx *Context, root Handle) bool {
	progress := false

	for !ctx.TodoEmpty() {
		conclusion := ctx.PopTodo()

		switch conclusion.Kind {
		case ConclusionSubsumerD:
			if e.applySubsumerDRules(ctx, root, conclusion) {
				progress = true
			}
		case ConclusionSubsumerC:
			if e.applySubsumerCRules(ctx, root, conclusion) {
				progress = true
			}
		}
	}

	ctx.Saturated = true
	return progress
}

func (e *SaturationEngine) applySubsumerCRules(ctx *Context, root Handle, concl Conclusion) bool {
	progress := false

	subsumers := e.index.GetSubsumers(concl.Target)
	subsumers.Iterate(func(super Handle) {
		if ctx.AddSubsumerD(super) {
			ctx.PushTodo(Conclusion{
				Kind:   ConclusionSubsumerD,
				Root:   root,
				Target: super,
			})
			progress = true

			ctx2 := e.getOrCreateContext(super)
			if ctx2.AddSubsumerC(root) {
				ctx2.PushTodo(Conclusion{
					Kind:   ConclusionSubsumerC,
					Root:   super,
					Target: root,
				})
			}
		}
	})

	domain, ok := e.index.GetPropertyDomain(concl.Target)
	if ok {
		if ctx.AddSubsumerC(domain) {
			ctx.PushTodo(Conclusion{
				Kind:   ConclusionSubsumerC,
				Root:   root,
				Target: domain,
			})
			progress = true
		}
	}

	return progress
}

func (e *SaturationEngine) applySubsumerDRules(ctx *Context, root Handle, concl Conclusion) bool {
	progress := false

	domain, ok := e.index.GetPropertyDomain(concl.Target)
	if ok {
		if ctx.AddSubsumerC(domain) {
			ctx.PushTodo(Conclusion{
				Kind:   ConclusionSubsumerC,
				Root:   root,
				Target: domain,
			})
			progress = true
		}
	}

	rng, ok := e.index.GetPropertyRange(concl.Target)
	if ok {
		if ctx.AddSubsumerC(rng) {
			ctx.PushTodo(Conclusion{
				Kind:   ConclusionSubsumerC,
				Root:   root,
				Target: rng,
			})
			progress = true
		}
	}

	propSubsumers := e.index.GetPropertySubsumers(concl.Target)
	propSubsumers.Iterate(func(prop Handle) {
		if ctx.AddForwardLink(concl.Target, prop) {
			ctx.PushTodo(Conclusion{
				Kind:   ConclusionForwardLink,
				Root:   root,
				Target: prop,
			})
			progress = true
		}
	})

	return progress
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
	case Handle(1): // rdfs:subClassOf
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
