package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/protocol"
)

// Scripted peers let tests stop at exact frame boundaries instead of relying on
// scheduler timing or TCP packet layouts to trigger cancellation and reconnects.
func scriptedPeer(t *testing.T, handle func(context.Context, int, *remoteWire) error) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var mu sync.Mutex
	connections := make(map[net.Conn]struct{})
	var wg sync.WaitGroup
	wg.Go(func() {
		for number := 1; ; number++ {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			connections[conn] = struct{}{}
			mu.Unlock()
			wg.Go(func() {
				defer func() { _ = conn.Close(); mu.Lock(); delete(connections, conn); mu.Unlock() }()
				_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
				if err := handle(ctx, number, newRemoteWire(conn)); err != nil && ctx.Err() == nil {
					t.Errorf("peer connection %d: %v", number, err)
				}
			})
		}
	})
	t.Cleanup(func() {
		cancel()
		_ = listener.Close()
		mu.Lock()
		for conn := range connections {
			_ = conn.Close()
		}
		mu.Unlock()
		wg.Wait()
	})
	return listener.Addr().String()
}

func peerAuthenticate(wire *remoteWire) error {
	request, err := wire.reader.Read()
	if err != nil {
		return err
	}
	if request.RequestID != 1 || request.Op != protocol.OpAuthenticate {
		return protocol.ErrInvalidCommand
	}
	return wire.writer.Write(protocol.Command{RequestID: request.RequestID, Op: request.Op})
}

func peerSubscribe(wire *remoteWire) error {
	request, err := wire.reader.Read()
	if err != nil {
		return err
	}
	if request.RequestID != 2 || request.Op != protocol.OpSubscribe {
		return protocol.ErrInvalidCommand
	}
	return wire.writer.Write(protocol.Command{RequestID: request.RequestID, Op: request.Op})
}

func peerEvent(wire *remoteWire, event nodis.Event) error {
	return wire.writer.Write(protocol.Command{RequestID: 2, Op: protocol.OpEvent, Payload: protocol.AppendEvent(nil, event)})
}

func awaitSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(3 * time.Second):
		t.Fatal("peer did not reach expected frame boundary")
	}
}

func awaitStatus(t *testing.T, status func() SyncStatus, ready func(SyncStatus) bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		current := status()
		if ready(current) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("sync status did not converge: %+v", current)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestRemoteCancellationQueueAndFreshCommandConnection(t *testing.T) {
	started := make(chan struct{})
	var creates atomic.Int32
	var accepts atomic.Int32
	address := scriptedPeer(t, func(ctx context.Context, number int, wire *remoteWire) error {
		accepts.Add(1)
		if err := peerAuthenticate(wire); err != nil {
			return err
		}
		request, err := wire.reader.Read()
		if err != nil {
			return err
		}
		if number == 1 {
			if request.Op != protocol.OpCreate {
				return fmt.Errorf("unexpected op %d", request.Op)
			}
			creates.Add(1)
			close(started)
			// No response: cancellation must close this stream, never reuse it.
			_, _ = wire.reader.Read()
			return nil
		}
		if request.Op == protocol.OpCreate {
			creates.Add(1)
			return errors.New("mutation was replayed")
		}
		if request.Op != protocol.OpGetDocument {
			return protocol.ErrInvalidCommand
		}
		payload := binary.LittleEndian.AppendUint64(nil, 42)
		return wire.writer.Write(protocol.Command{RequestID: request.RequestID, Op: request.Op, Payload: payload})
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	remote, err := DialWithOptions(ctx, address, []byte("token"), DialOptions{DisableSync: true})
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	defer cancelRequest()
	future := remote.CreateAsync(requestCtx, 1, map[string]any{"name": "Ada"})
	awaitSignal(t, started)
	queuedCtx, cancelQueued := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancelQueued()
	queued := make(chan error, 1)
	go func() { _, err := remote.Document(queuedCtx, "queued"); queued <- err }()
	select {
	case err := <-queued:
		if !errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrOutcomeUnknown) {
			t.Fatalf("queued request = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("queued request ignored cancellation")
	}
	cancelRequest()
	_, err = future.Wait(ctx)
	if !errors.Is(err, context.Canceled) || !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("interrupted write = %v", err)
	}
	var uncertain *OutcomeUnknownError
	if !errors.As(err, &uncertain) || uncertain.Op != protocol.OpCreate {
		t.Fatal("write uncertainty lost operation")
	}
	id, err := remote.Document(ctx, "after-cancel")
	if err != nil || id != 42 {
		t.Fatalf("fresh request = %d, %v", id, err)
	}
	if creates.Load() != 1 || accepts.Load() != 2 {
		t.Fatal("request replayed or connection was reused")
	}
}

func TestRemoteConcurrentCloseInterruptsBlockedRequest(t *testing.T) {
	started := make(chan struct{})
	address := scriptedPeer(t, func(ctx context.Context, _ int, wire *remoteWire) error {
		if err := peerAuthenticate(wire); err != nil {
			return err
		}
		if _, err := wire.reader.Read(); err != nil {
			return err
		}
		close(started)
		_, _ = wire.reader.Read()
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	remote, err := DialWithOptions(ctx, address, nil, DialOptions{DisableSync: true})
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	result := make(chan error, 1)
	go func() { _, err := remote.Get(context.Background(), 1, 1); result <- err }()
	awaitSignal(t, started)
	closed := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(remote.Close)
		}
		wg.Wait()
		close(closed)
	}()
	awaitSignal(t, closed)
	select {
	case err := <-result:
		if !errors.Is(err, nodis.ErrClosed) {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("closed connection left request blocked")
	}
	if _, err := remote.Document(ctx, "closed"); !errors.Is(err, nodis.ErrClosed) {
		t.Fatal(err)
	}
}

func TestRemoteReconnectPublishesOnlyCompleteReplacement(t *testing.T) {
	drop := make(chan struct{})
	partial := make(chan struct{})
	finish := make(chan struct{})
	var connections atomic.Int32
	address := scriptedPeer(t, func(ctx context.Context, number int, wire *remoteWire) error {
		connections.Add(1)
		if err := peerAuthenticate(wire); err != nil {
			return err
		}
		if number == 1 {
			<-ctx.Done()
			return nil
		}
		if err := peerSubscribe(wire); err != nil {
			return err
		}
		switch number {
		case 2:
			if err := peerEvent(wire, objectEvent(1, 1, 5, 1, map[string]string{"name": "Ada", "secret": "private"})); err != nil {
				return err
			}
			if err := peerEvent(wire, objectEvent(2, 1, 5, 1, map[string]string{"name": "deleted-while-offline"})); err != nil {
				return err
			}
			if err := peerEvent(wire, snapshotDone(5, 1)); err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return nil
			case <-drop:
				return nil
			}
		case 3:
			// An interrupted snapshot must not leak even one of its records.
			return peerEvent(wire, objectEvent(3, 1, 7, 2, map[string]string{"name": "incomplete"}))
		case 4:
			if err := peerEvent(wire, objectEvent(1, 2, 9, 3, map[string]string{"name": "Grace"})); err != nil {
				return err
			}
			close(partial)
			select {
			case <-ctx.Done():
				return nil
			case <-finish:
			}
			if err := peerEvent(wire, snapshotDone(9, 3)); err != nil {
				return err
			}
			// This next frame overwrites the reusable reader buffer.
			if err := peerEvent(wire, nodis.Event{Type: nodis.EventDocumentRead, Document: 1, Sequence: 9, PermissionRevision: 3}); err != nil {
				return err
			}
			<-ctx.Done()
			return nil
		default:
			return errors.New("unexpected extra reconnect")
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	remote, err := Dial(ctx, address, nil, nodis.SyncScope{})
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	old, ok := remote.Cached(1, 1)
	if !ok {
		t.Fatal("initial cache missing")
	}
	close(drop)
	awaitSignal(t, partial)
	if remote.SyncStatus().Ready {
		t.Fatal("partial replacement marked ready")
	}
	for _, id := range []nodis.ObjectID{1, 2, 3} {
		if _, ok := remote.Cached(1, id); ok {
			t.Fatalf("object %d visible during resnapshot", id)
		}
	}
	close(finish)
	awaitStatus(t, remote.SyncStatus, func(status SyncStatus) bool {
		return status.Ready && status.Sequence == 9 && status.PermissionRevision == 3
	})
	current, ok := remote.Cached(1, 1)
	if !ok {
		t.Fatal("complete replacement missing")
	}
	if name, _ := current.String("name"); name != "Grace" {
		t.Fatal("reader reuse corrupted cache")
	}
	if _, exists := current.String("secret"); exists {
		t.Fatal("revoked field survived reconnect")
	}
	for _, id := range []nodis.ObjectID{2, 3} {
		if _, ok := remote.Cached(1, id); ok {
			t.Fatal("offline deletion or partial snapshot survived")
		}
	}
	if name, _ := old.String("name"); name != "Ada" {
		t.Fatal("previous result changed after reconnect")
	}
	if connections.Load() != 4 {
		t.Fatal("unexpected connection lifecycle")
	}
}

func TestRemoteCloseCancelsReconnectAuthentication(t *testing.T) {
	drop := make(chan struct{})
	authStarted := make(chan struct{})
	address := scriptedPeer(t, func(ctx context.Context, number int, wire *remoteWire) error {
		if number == 3 {
			if _, err := wire.reader.Read(); err != nil {
				return err
			}
			close(authStarted)
			// Never answer authentication; Close must cancel this unpublished socket.
			_, _ = wire.reader.Read()
			return nil
		}
		if err := peerAuthenticate(wire); err != nil {
			return err
		}
		if number == 1 {
			<-ctx.Done()
			return nil
		}
		if err := peerSubscribe(wire); err != nil {
			return err
		}
		if err := peerEvent(wire, snapshotDone(0, 0)); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
		case <-drop:
		}
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	remote, err := Dial(ctx, address, nil, nodis.SyncScope{})
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	close(drop)
	awaitSignal(t, authStarted)
	closed := make(chan struct{})
	go func() { remote.Close(); close(closed) }()
	awaitSignal(t, closed)
	if !errors.Is(remote.SyncStatus().Err, nodis.ErrClosed) {
		t.Fatal("closed status lost")
	}
}

func TestRemoteRejectsUncorrelatedAuthenticationResponse(t *testing.T) {
	address := scriptedPeer(t, func(_ context.Context, _ int, wire *remoteWire) error {
		request, err := wire.reader.Read()
		if err != nil {
			return err
		}
		return wire.writer.Write(protocol.Command{RequestID: request.RequestID + 1, Op: request.Op})
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if remote, err := DialWithOptions(ctx, address, nil, DialOptions{DisableSync: true}); !errors.Is(err, protocol.ErrInvalidCommand) {
		if remote != nil {
			remote.Close()
		}
		t.Fatal("accepted uncorrelated authentication", err)
	}
}

func TestRemoteInvalidEventHidesCacheBeforeReconnect(t *testing.T) {
	for _, damage := range []string{"request-id", "frame-checksum", "object-id"} {
		t.Run(damage, func(t *testing.T) {
			trigger := make(chan struct{})
			reconnecting := make(chan struct{})
			address := scriptedPeer(t, func(ctx context.Context, number int, wire *remoteWire) error {
				if number == 3 {
					if _, err := wire.reader.Read(); err != nil {
						return err
					}
					close(reconnecting)
					_, _ = wire.reader.Read()
					return nil
				}
				if err := peerAuthenticate(wire); err != nil {
					return err
				}
				if number == 1 {
					<-ctx.Done()
					return nil
				}
				if err := peerSubscribe(wire); err != nil {
					return err
				}
				if err := peerEvent(wire, objectEvent(1, 1, 1, 1, map[string]string{"name": "Ada"})); err != nil {
					return err
				}
				if err := peerEvent(wire, snapshotDone(1, 1)); err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return nil
				case <-trigger:
				}
				event := objectEvent(1, 2, 2, 1, map[string]string{"name": "changed"})
				event.Type = nodis.EventObjectUpdated
				if damage == "object-id" {
					event.Object = 2
				}
				message := protocol.Command{RequestID: 2, Op: protocol.OpEvent, Payload: protocol.AppendEvent(nil, event)}
				if damage == "request-id" {
					message.RequestID++
				}
				frame := protocol.AppendCommand(nil, message)
				if damage == "frame-checksum" {
					frame[len(frame)-1] ^= 1
				}
				if _, err := wire.conn.Write(frame); err != nil {
					return err
				}
				<-ctx.Done()
				return nil
			})
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			remote, err := Dial(ctx, address, nil, nodis.SyncScope{})
			if err != nil {
				t.Fatal(err)
			}
			defer remote.Close()
			close(trigger)
			awaitSignal(t, reconnecting)
			status := remote.SyncStatus()
			if status.Ready || status.Err == nil {
				t.Fatalf("invalid event left a ready cache: %+v", status)
			}
			if _, ok := remote.Cached(1, 1); ok {
				t.Fatal("invalid stream retained visible data")
			}
		})
	}
}
