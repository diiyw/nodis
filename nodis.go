package nodis

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/diiyw/nodis/storage"

	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/internal/listener"
	"github.com/diiyw/nodis/patch"
	"github.com/diiyw/nodis/redis"
	"github.com/tidwall/btree"
)

var (
	ErrUnknownOperation = errors.New("unknown operation")
)

type Nodis struct {
	store             *store
	listeners         []*listener.Listener
	blockingKeysMutex sync.RWMutex
	blockingKeys      btree.Map[string, *list.LinkedListG[chan string]] // blocking keys
	options           *Options
}

func Open(opt *Options) *Nodis {
	if opt.Storage == nil {
		opt.Storage = &storage.Memory{}
	}
	n := &Nodis{
		options: opt,
	}
	n.store = newStore(opt.Storage)
	go func() {
		if opt.GCDuration != 0 {
			for {
				time.Sleep(opt.GCDuration)
				n.store.gc()
			}
		}
	}()
	go func() {
		if opt.SnapshotDuration != 0 {
			for {
				time.Sleep(opt.SnapshotDuration)
				if err := n.Snapshot(); err != nil {
					log.Println("Snapshot Err: ", err)
					continue
				}
				log.Println("Snapshot at", time.Now().Format("2006-01-02 15:04:05"))
			}
		}
	}()
	return n
}

// Snapshot saves the data to disk
func (n *Nodis) Snapshot() error {
	return n.store.sg.Snapshot()
}

// Close the store
func (n *Nodis) Close() error {
	return n.store.close()
}

// Clear removes all keys from the store
func (n *Nodis) Clear() {
	err := n.store.clear()
	if err != nil {
		log.Println("Clear: ", err)
	}
}

func (n *Nodis) notify(f func() []patch.Op) {
	if len(n.listeners) == 0 {
		return
	}
	go func() {
		for _, w := range n.listeners {
			for _, op := range f() {
				if w.Matched(op.Data.GetKey()) {
					w.Push(op)
				}
			}
		}
	}()
}

func (n *Nodis) WatchKey(pattern []string, fn func(op patch.Op)) int {
	w := listener.New(pattern, fn)
	n.listeners = append(n.listeners, w)
	return len(n.listeners) - 1
}

func (n *Nodis) UnWatchKey(id int) {
	n.listeners = append(n.listeners[:id], n.listeners[id+1:]...)
}

func (n *Nodis) ApplyPatch(ops ...patch.Op) error {
	for _, op := range ops {
		err := n.appylyPatch(op)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Nodis) appylyPatch(p patch.Op) error {
	switch op := p.Data.(type) {
	case *patch.OpClear:
		n.Clear()
	case *patch.OpDel:
		n.Del(op.Key)
	case *patch.OpExpire:
		n.Expire(op.Key, op.Expiration)
	case *patch.OpExpireAt:
		n.ExpireAt(op.Key, time.Unix(op.Expiration, 0))
	case *patch.OpHClear:
		n.HClear(op.Key)
	case *patch.OpHDel:
		n.HDel(op.Key, op.Fields...)
	case *patch.OpHIncrBy:
		_, err := n.HIncrBy(op.Key, op.Field, op.IncrInt)
		return err
	case *patch.OpHIncrByFloat:
		_, err := n.HIncrByFloat(op.Key, op.Field, op.IncrFloat)
		return err
	case *patch.OpHSet:
		n.HSet(op.Key, op.Field, op.Value)
	case *patch.OpLInsert:
		n.LInsert(op.Key, op.Pivot, op.Value, op.Before)
	case *patch.OpLPop:
		n.LPop(op.Key, op.Count)
	case *patch.OpLPopRPush:
		n.LPopRPush(op.Key, op.DstKey)
	case *patch.OpLPush:
		n.LPush(op.Key, op.Values...)
	case *patch.OpLPushX:
		n.LPushX(op.Key, op.Value)
	case *patch.OpLRem:
		n.LRem(op.Key, op.Value, op.Count)
	case *patch.OpLSet:
		n.LSet(op.Key, op.Index, op.Value)
	case *patch.OpLTrim:
		n.LTrim(op.Key, op.Start, op.Stop)
	case *patch.OpRPop:
		n.RPop(op.Key, op.Count)
	case *patch.OpRPopLPush:
		n.RPopLPush(op.Key, op.DstKey)
	case *patch.OpRPush:
		n.RPush(op.Key, op.Values...)
	case *patch.OpRPushX:
		n.RPushX(op.Key, op.Value)
	case *patch.OpSAdd:
		n.SAdd(op.Key, op.Members...)
	case *patch.OpSRem:
		n.SRem(op.Key, op.Members...)
	case *patch.OpSet:
		n.Set(op.Key, op.Value, op.KeepTTL)
	case *patch.OpZAdd:
		n.ZAdd(op.Key, op.Member, op.Score)
	case *patch.OpZClear:
		n.ZClear(op.Key)
	case *patch.OpZIncrBy:
		n.ZIncrBy(op.Key, op.Member, op.Score)
	case *patch.OpZRem:
		n.ZRem(op.Key, op.Member)
	case *patch.OpZRemRangeByRank:
		n.ZRemRangeByRank(op.Key, op.Start, op.Stop)
	case *patch.OpZRemRangeByScore:
		n.ZRemRangeByScore(op.Key, op.Min, op.Max, int(op.Mode))
	case *patch.OpRename:
		return n.Rename(op.Key, op.DstKey)
	default:
		return ErrUnknownOperation
	}
	return nil
}

func (n *Nodis) Broadcast(addr string, pattern []string) error {
	return n.options.Synchronizer.Publish(addr, func(s SyncConn) {
		id := n.WatchKey(pattern, func(op patch.Op) {
			err := s.Send(op)
			if err != nil {
				log.Println("Publish: ", err)
			}
		})
		s.Wait()
		n.UnWatchKey(id)
	})
}

func (n *Nodis) Subscribe(addr string) error {
	return n.options.Synchronizer.Subscribe(addr, func(o patch.Op) {
		n.ApplyPatch(o)
	})
}

func (n *Nodis) Serve(addr string) error {
	log.Println("Nodis listen on", addr)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		<-c
		log.Printf("Nodis closed %v \n", n.Close())
		os.Exit(0)
	}()
	return redis.Serve(addr, func(conn *redis.Conn, cmd redis.Command) {
		c := getCommand(cmd.Name)
		c(n, conn, cmd)
		if conn.HasError() && conn.State != 0 {
			conn.State |= redis.MultiError
		}
	})
}

func (n *Nodis) exec(fn func(tx *Tx) error) error {
	tx := &Tx{
		store:       n.store,
		lockedMetas: make([]*metadata, 0),
	}
	defer tx.commit()
	return fn(tx)
}
