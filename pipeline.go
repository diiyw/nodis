package nodis

import "context"

type Pipeline interface {
	Execute(context.Context, *PipelineContext) error
}

type PipelineFunc func(context.Context, *PipelineContext) error

func (fn PipelineFunc) Execute(ctx context.Context, pipeline *PipelineContext) error {
	return fn(ctx, pipeline)
}

type PipelineContext struct {
	Principal Principal
	Document  DocumentID
	Object    ObjectID
	Previous  *ObjectView
	Draft     *Draft
}

func (db *DB) RegisterPipeline(id PipelineID, pipeline Pipeline) error {
	if id == 0 || pipeline == nil {
		return ErrInvalidValue
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return ErrClosed
	}
	if _, ok := db.pipelines[id]; ok {
		return ErrAlreadyExists
	}
	db.pipelines[id] = pipeline
	return nil
}
