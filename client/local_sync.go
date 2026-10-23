package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/diiyw/nodis"
)

type Client struct {
	db        *nodis.DB
	principal nodis.Principal
	scope     nodis.SyncScope
	cache     viewCache
	sub       *nodis.Subscription
	subMu     sync.Mutex
	closed    chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

func Connect(ctx context.Context, db *nodis.DB, principal nodis.Principal, scope nodis.SyncScope) (*Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	principal.Roles = append([]string(nil), principal.Roles...)
	principal.Groups = append([]string(nil), principal.Groups...)
	scope.Documents = append([]nodis.DocumentID(nil), scope.Documents...)
	sub, err := db.Subscribe(principal, scope)
	if err != nil {
		return nil, err
	}
	c := &Client{db: db, principal: principal, scope: scope, sub: sub, closed: make(chan struct{}), done: make(chan struct{})}
	c.cache.init(true)
	if err := c.consumeSnapshot(ctx, sub); err != nil {
		sub.Close()
		return nil, err
	}
	go c.run()
	return c, nil
}

func (c *Client) consumeSnapshot(ctx context.Context, sub *nodis.Subscription) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.closed:
			return nodis.ErrClosed
		case event, ok := <-sub.C:
			if !ok {
				return nodis.ErrInvalidState
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := c.cache.stage(event); err != nil {
				return err
			}
			if event.Type == nodis.EventSnapshotDone {
				return nil
			}
		}
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		c.cache.close()
		c.subMu.Lock()
		c.sub.Close()
		c.subMu.Unlock()
		<-c.done
	})
}

func (c *Client) run() {
	defer close(c.done)
	for {
		select {
		case <-c.closed:
			return
		case event, ok := <-c.sub.C:
			var err error
			if ok && event.Type != nodis.EventReset {
				err = c.cache.apply(event)
				if err == nil {
					continue
				}
			}
			if !c.resubscribe(err) {
				return
			}
		}
	}
}

func (c *Client) resubscribe(reason error) bool {
	c.cache.invalidate(reason)
	c.subMu.Lock()
	c.sub.Close()
	c.subMu.Unlock()
	delay := 100 * time.Millisecond
	for {
		select {
		case <-c.closed:
			return false
		default:
		}
		sub, err := c.db.Subscribe(c.principal, c.scope)
		if err == nil {
			c.subMu.Lock()
			c.sub = sub
			c.subMu.Unlock()
			err = c.consumeSnapshot(context.Background(), sub)
			if err == nil {
				return true
			}
			sub.Close()
		}
		c.cache.invalidate(err)
		if errors.Is(err, nodis.ErrClosed) {
			return false
		}
		timer := time.NewTimer(delay)
		select {
		case <-c.closed:
			timer.Stop()
			return false
		case <-timer.C:
		}
		delay = min(delay*2, 5*time.Second)
	}
}

func (c *Client) Cached(document nodis.DocumentID, object nodis.ObjectID) (*nodis.PortableObject, bool) {
	return c.cache.cached(document, object)
}

func (c *Client) SyncStatus() SyncStatus { return c.cache.syncStatus() }

func (c *Client) operationError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-c.closed:
		return nodis.ErrClosed
	default:
		return nil
	}
}
