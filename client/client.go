// Package client exposes a permission-filtered local materialized view while
// executing every mutation against the authoritative server database.
package client

import (
	"context"

	"github.com/diiyw/nodis"
)

type key struct {
	document nodis.DocumentID
	object   nodis.ObjectID
}

func (c *Client) Get(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID) (*nodis.ObjectView, error) {
	if err := c.operationError(ctx); err != nil {
		return nil, err
	}
	return c.db.Get(c.principal, document, object)
}

func (c *Client) Create(ctx context.Context, document nodis.DocumentID, fields map[string]any) (nodis.ObjectID, error) {
	if err := c.operationError(ctx); err != nil {
		return 0, err
	}
	tx := c.db.Begin(c.principal).WithContext(ctx)
	draft, err := tx.Create(document)
	if err != nil {
		return 0, err
	}
	for name, value := range fields {
		if err := draft.Set(name, value); err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	id, err := draft.Save()
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (c *Client) Update(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID, revision uint64, fields map[string]any) error {
	if err := c.operationError(ctx); err != nil {
		return err
	}
	tx := c.db.Begin(c.principal).WithContext(ctx)
	draft, err := tx.Update(document, object, revision)
	if err != nil {
		return err
	}
	for name, value := range fields {
		if err := draft.Set(name, value); err != nil {
			tx.Rollback()
			return err
		}
	}
	if _, err := draft.Save(); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (c *Client) Delete(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID, revision uint64) error {
	if err := c.operationError(ctx); err != nil {
		return err
	}
	tx := c.db.Begin(c.principal).WithContext(ctx)
	if err := tx.Delete(document, object, revision); err != nil {
		return err
	}
	return tx.Commit()
}

type Result[T any] struct {
	Value T
	Err   error
}
type Future[T any] struct {
	done   <-chan struct{}
	result *Result[T]
}

func (f Future[T]) Wait(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case <-f.done:
		return f.result.Value, f.result.Err
	}
}

func (c *Client) CreateAsync(ctx context.Context, document nodis.DocumentID, fields map[string]any) Future[nodis.ObjectID] {
	fields = cloneFields(fields)
	done := make(chan struct{})
	result := &Result[nodis.ObjectID]{}
	go func() {
		result.Value, result.Err = c.Create(ctx, document, fields)
		close(done)
	}()
	return Future[nodis.ObjectID]{done: done, result: result}
}

func (c *Client) GetAsync(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID) Future[*nodis.ObjectView] {
	done := make(chan struct{})
	result := &Result[*nodis.ObjectView]{}
	go func() { result.Value, result.Err = c.Get(ctx, document, object); close(done) }()
	return Future[*nodis.ObjectView]{done: done, result: result}
}

func (c *Client) UpdateAsync(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID, revision uint64, fields map[string]any) Future[struct{}] {
	fields = cloneFields(fields)
	done := make(chan struct{})
	result := &Result[struct{}]{}
	go func() {
		result.Err = c.Update(ctx, document, object, revision, fields)
		close(done)
	}()
	return Future[struct{}]{done: done, result: result}
}

func (c *Client) DeleteAsync(ctx context.Context, document nodis.DocumentID, object nodis.ObjectID, revision uint64) Future[struct{}] {
	done := make(chan struct{})
	result := &Result[struct{}]{}
	go func() { result.Err = c.Delete(ctx, document, object, revision); close(done) }()
	return Future[struct{}]{done: done, result: result}
}

func cloneFields(fields map[string]any) map[string]any {
	clone := make(map[string]any, len(fields))
	for name, value := range fields {
		inferred, err := nodis.ValueOf(value)
		if err != nil {
			clone[name] = value
			continue
		}
		clone[name] = inferred.Clone()
	}
	return clone
}
