package nodis

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diiyw/nodis/fs"
	"github.com/diiyw/nodis/pb"
	"github.com/diiyw/nodis/redis"
	nSync "github.com/diiyw/nodis/sync"
	"github.com/diiyw/nodis/watch"
)

var (
	ErrUnknownOperation = errors.New("unknown operation")
)

type Nodis struct {
	options  *Options
	store    *store
	watchers []*watch.Watcher
}

func Open(opt *Options) *Nodis {
	if opt.FileSize == 0 {
		opt.FileSize = FileSizeGB
	}
	if opt.Filesystem == nil {
		opt.Filesystem = &fs.Memory{}
	}
	if opt.MetaPoolSize == 0 {
		opt.MetaPoolSize = 10240
	}
	n := &Nodis{
		options: opt,
	}
	isDir, err := opt.Filesystem.IsDir(opt.Path)
	if err != nil {
		if os.IsNotExist(err) {
			err = opt.Filesystem.MkdirAll(opt.Path)
			if err != nil {
				panic(err)
			}
		}
	} else if !isDir {
		panic("Path is not a directory")
	}
	n.store = newStore(opt.Path, opt.FileSize, opt.MetaPoolSize, opt.Filesystem)
	go func() {
		if opt.TidyDuration != 0 {
			for {
				time.Sleep(opt.TidyDuration)
				n.store.tidy(opt.TidyDuration.Milliseconds())
			}
		}
	}()
	go func() {
		if opt.SnapshotDuration != 0 {
			for {
				time.Sleep(opt.SnapshotDuration)
				n.Snapshot(n.store.path)
				log.Println("Snapshot at", time.Now().Format("2006-01-02 15:04:05"))
			}
		}
	}()
	return n
}

// Snapshot saves the data to disk
func (n *Nodis) Snapshot(path string) {
	n.store.snapshot(path)
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

// SetEntry sets an entity
func (n *Nodis) SetEntry(data []byte) error {
	entity, err := n.store.parseEntry(data)
	if err != nil {
		return err
	}
	return n.store.putEntry(entity)
}

// GetEntry gets an entity
func (n *Nodis) GetEntry(key string) (data []byte) {
	_ = n.Update(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		var entity = newEntry(key, meta.ds, meta.key.expiration)
		data, _ = entity.Marshal()
		return nil
	})
	return
}

// parseDs the data
func (n *Nodis) notify(ops ...*pb.Op) {
	if len(n.watchers) == 0 {
		return
	}
	go func() {
		for _, w := range n.watchers {
			for _, op := range ops {
				if w.Matched(op.Key) {
					w.Push(op.Operation)
					op.Reset()
				}
			}
		}
	}()
}

func (n *Nodis) Watch(pattern []string, fn func(op *pb.Operation)) int {
	w := watch.NewWatcher(pattern, fn)
	n.watchers = append(n.watchers, w)
	return len(n.watchers) - 1
}

func (n *Nodis) UnWatch(id int) {
	n.watchers = append(n.watchers[:id], n.watchers[id+1:]...)
}

func (n *Nodis) Patch(ops ...*pb.Op) error {
	for _, op := range ops {
		err := n.patch(op)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Nodis) patch(op *pb.Op) error {
	switch op.Operation.Type {
	case pb.OpType_Clear:
		n.Clear()
	case pb.OpType_Del:
		n.Del(op.Key)
	case pb.OpType_Expire:
		n.Expire(op.Key, op.Operation.Expiration)
	case pb.OpType_ExpireAt:
		n.ExpireAt(op.Key, time.Unix(op.Operation.Expiration, 0))
	case pb.OpType_HClear:
		n.HClear(op.Key)
	case pb.OpType_HDel:
		n.HDel(op.Key, op.Operation.Fields...)
	case pb.OpType_HIncrBy:
		n.HIncrBy(op.Key, op.Operation.Field, op.Operation.IncrInt)
	case pb.OpType_HIncrByFloat:
		n.HIncrByFloat(op.Key, op.Operation.Field, op.Operation.IncrFloat)
	case pb.OpType_HSet:
		n.HSet(op.Key, op.Operation.Field, op.Operation.Value)
	case pb.OpType_LInsert:
		n.LInsert(op.Key, op.Operation.Pivot, op.Operation.Value, op.Operation.Before)
	case pb.OpType_LPop:
		n.LPop(op.Key, op.Operation.Count)
	case pb.OpType_LPopRPush:
		n.LPopRPush(op.Key, op.Operation.DstKey)
	case pb.OpType_LPush:
		n.LPush(op.Key, op.Operation.Values...)
	case pb.OpType_LPushX:
		n.LPushX(op.Key, op.Operation.Value)
	case pb.OpType_LRem:
		n.LRem(op.Key, op.Operation.Value, op.Operation.Count)
	case pb.OpType_LSet:
		n.LSet(op.Key, op.Operation.Index, op.Operation.Value)
	case pb.OpType_LTrim:
		n.LTrim(op.Key, op.Operation.Start, op.Operation.Stop)
	case pb.OpType_RPop:
		n.RPop(op.Key, op.Operation.Count)
	case pb.OpType_RPopLPush:
		n.RPopLPush(op.Key, op.Operation.DstKey)
	case pb.OpType_RPush:
		n.RPush(op.Key, op.Operation.Values...)
	case pb.OpType_RPushX:
		n.RPushX(op.Key, op.Operation.Value)
	case pb.OpType_SAdd:
		n.SAdd(op.Key, op.Operation.Members...)
	case pb.OpType_SRem:
		n.SRem(op.Key, op.Operation.Members...)
	case pb.OpType_Set:
		n.Set(op.Key, op.Operation.Value)
	case pb.OpType_ZAdd:
		n.ZAdd(op.Key, op.Operation.Member, op.Operation.Score)
	case pb.OpType_ZClear:
		n.ZClear(op.Key)
	case pb.OpType_ZIncrBy:
		n.ZIncrBy(op.Key, op.Operation.Member, op.Operation.Score)
	case pb.OpType_ZRem:
		n.ZRem(op.Key, op.Operation.Member)
	case pb.OpType_ZRemRangeByRank:
		n.ZRemRangeByRank(op.Key, op.Operation.Start, op.Operation.Stop)
	case pb.OpType_ZRemRangeByScore:
		n.ZRemRangeByScore(op.Key, op.Operation.Min, op.Operation.Max)
	case pb.OpType_Rename:
		_ = n.Rename(op.Key, op.Operation.DstKey)
	default:
		return ErrUnknownOperation
	}
	return nil
}

func (n *Nodis) Publish(addr string, pattern []string) error {
	return n.options.Synchronizer.Publish(addr, func(s nSync.Conn) {
		id := n.Watch(pattern, func(op *pb.Operation) {
			err := s.Send(&pb.Op{Operation: op})
			if err != nil {
				log.Println("Publish: ", err)
			}
		})
		s.Wait()
		n.UnWatch(id)
	})
}

func (n *Nodis) Subscribe(addr string) error {
	return n.options.Synchronizer.Subscribe(addr, func(o *pb.Op) {
		n.Patch(o)
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
	return redis.Serve(addr, func(conn *redis.Conn, cmd *redis.Command) {
		c := getCommand(cmd.Name)
		if c == nil {
			conn.WriteError("ERR unknown command: " + cmd.Name)
			return
		}
		if conn.Multi {
			if cmd.Name == "MULTI" {
				conn.WriteError("ERR MULTI calls can not be nested")
				return
			}
			if cmd.Name != "MULTI" && cmd.Name != "EXEC" {
				newCmd := &redis.Command{
					Name:    cmd.Name,
					Options: cmd.Options,
				}
				newCmd.Args = cmd.Args
				conn.Commands = append(conn.Commands, newCmd)
				conn.WriteString("QUEUED")
				return
			}
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Recovered from ", r)
					conn.WriteError("ERR " + cmd.Name + " error" + r.(error).Error())
					return
				}
			}()
			c(n, conn, cmd)
		}()
	})
}

func (n *Nodis) Update(fn func(tx *Tx) error) error {
	tx := &Tx{
		store:       n.store,
		lockedMetas: make([]*metadata, 0),
	}
	defer tx.commit()
	return fn(tx)
}
