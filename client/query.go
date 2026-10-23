package client

import (
	"context"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/protocol"
)

func (c *Client) Find(ctx context.Context, query nodis.Query) (nodis.QueryPage, error) {
	if err := c.operationError(ctx); err != nil {
		return nodis.QueryPage{}, err
	}
	return c.db.Find(ctx, c.principal, query)
}

func (c *Client) FindInto(ctx context.Context, query nodis.Query, buffer *nodis.QueryBuffer) (nodis.QueryPage, error) {
	if err := c.operationError(ctx); err != nil {
		return nodis.QueryPage{}, err
	}
	return c.db.FindInto(ctx, c.principal, query, buffer)
}

func (r *Remote) Find(ctx context.Context, query nodis.Query) (nodis.QueryPage, error) {
	payload, err := protocol.AppendQuery(nil, query)
	if err != nil {
		return nodis.QueryPage{}, err
	}
	response, err := r.request(ctx, protocol.OpFind, payload)
	if err != nil {
		return nodis.QueryPage{}, err
	}
	var decoder protocol.QueryPageDecoder
	page, err := protocol.DecodeQueryPageInto(response, &decoder)
	if err != nil {
		return nodis.QueryPage{}, err
	}
	limit := query.Limit
	if limit == 0 {
		limit = nodis.DefaultQueryLimit
	}
	if len(page.Objects) > limit || len(page.Objects) > 0 && page.Objects[0].ID() <= query.After {
		return nodis.QueryPage{}, protocol.ErrInvalidCommand
	}
	return page, nil
}

func (c *Client) FindAsync(ctx context.Context, query nodis.Query) Future[nodis.QueryPage] {
	if err := query.Validate(); err != nil {
		return completedFuture(nodis.QueryPage{}, err)
	}
	query = query.Clone()
	done := make(chan struct{})
	result := &Result[nodis.QueryPage]{}
	go func() { result.Value, result.Err = c.Find(ctx, query); close(done) }()
	return Future[nodis.QueryPage]{done: done, result: result}
}

func (r *Remote) FindAsync(ctx context.Context, query nodis.Query) Future[nodis.QueryPage] {
	if err := query.Validate(); err != nil {
		return completedFuture(nodis.QueryPage{}, err)
	}
	query = query.Clone()
	done := make(chan struct{})
	result := &Result[nodis.QueryPage]{}
	go func() { result.Value, result.Err = r.Find(ctx, query); close(done) }()
	return Future[nodis.QueryPage]{done: done, result: result}
}

func completedFuture[T any](value T, err error) Future[T] {
	done := make(chan struct{})
	close(done)
	return Future[T]{done: done, result: &Result[T]{Value: value, Err: err}}
}
