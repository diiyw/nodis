package nodis

import (
	"errors"
	"hash/crc32"
	"log"
	"os"
	"sync"
	"time"

	"encoding/binary"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/hash"
	"github.com/diiyw/nodis/ds/list"
	"github.com/diiyw/nodis/ds/set"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/fs"
	"github.com/diiyw/nodis/pb"
	nSync "github.com/diiyw/nodis/sync"
	"github.com/diiyw/nodis/watch"
	"github.com/tidwall/btree"
	"google.golang.org/protobuf/proto"
)

var (
	ErrUnknownOperation = errors.New("unknown operation")
)

type Nodis struct {
	sync.RWMutex
	dataStructs btree.Map[string, ds.DataStruct]
	keys        btree.Map[string, *Key]
	options     *Options
	store       *store
	closed      bool
	watchers    []*watch.Watcher
}

func Open(opt *Options) *Nodis {
	if opt.FileSize == 0 {
		opt.FileSize = FileSizeGB
	}
	if opt.Filesystem == nil {
		opt.Filesystem = &fs.Memory{}
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
	n.store = newStore(opt.Path, opt.FileSize, opt.Filesystem)
	go func() {
		if opt.RecycleDuration != 0 {
			for {
				time.Sleep(opt.RecycleDuration)
				n.Recycle()
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
	n.Recycle()
	n.store.snapshot(path, n.getChangedEntries())
}

// Recycle removes expired and unused keys
func (n *Nodis) Recycle() {
	n.Lock()
	defer n.Unlock()
	if n.closed {
		return
	}
	now := uint32(time.Now().Unix())
	recycleTime := now - uint32(n.options.RecycleDuration.Seconds())
	n.keys.Scan(func(key string, k *Key) bool {
		if k.expired() {
			n.dataStructs.Delete(key)
			n.keys.Delete(key)
			n.store.remove(key)
			return true
		}
		lastUse := k.lastUse.Load()
		if lastUse != 0 && lastUse <= recycleTime {
			d, ok := n.dataStructs.Get(key)
			n.dataStructs.Delete(key)
			n.keys.Delete(key)
			if ok {
				k.changed.Store(false)
				// save to disk
				err := n.store.put(newEntry(key, d, k.Expiration))
				if err != nil {
					log.Println("Recycle: ", err)
				}
			}
		}
		return true
	})
}

// getChangedEntries returns all keys that have been getChangedEntries
func (n *Nodis) getChangedEntries() []*pb.Entry {
	entries := make([]*pb.Entry, 0)
	n.keys.Scan(func(key string, k *Key) bool {
		if !k.changed.Load() || k.expired() {
			return true
		}
		d, ok := n.dataStructs.Get(key)
		if !ok {
			return true
		}
		entries = append(entries, newEntry(key, d, k.Expiration))
		return true
	})
	return entries
}

// Close the store
func (n *Nodis) Close() error {
	n.Lock()
	defer n.Unlock()
	// save values to disk
	entries := n.getChangedEntries()
	for _, entry := range entries {
		err := n.store.put(entry)
		if err != nil {
			return err
		}
	}
	n.closed = true
	return n.store.close()
}

func (n *Nodis) getDs(key string, newFn func() ds.DataStruct, ttl int64) (*Key, ds.DataStruct) {
	n.Lock()
	defer n.Unlock()
	k, ok := n.exists(key)
	if !ok && newFn == nil {
		return nil, nil
	}
	d, ok := n.dataStructs.Get(key)
	if !ok {
		if newFn != nil {
			d = newFn()
			n.dataStructs.Set(key, d)
			k = newKey(d.Type(), ttl)
			k.lastUse.Store(uint32(time.Now().Unix()))
			n.keys.Set(key, k)
			return k, d
		}
		return nil, nil
	}
	k.lastUse.Store(uint32(time.Now().Unix()))
	return k, d
}

// Clear removes all keys from the store
func (n *Nodis) Clear() {
	n.Lock()
	defer n.Unlock()
	n.dataStructs.Clear()
	n.keys.Clear()
	err := n.store.clear()
	if err != nil {
		log.Println("Clear: ", err)
	}
}

// SetEntry sets an entity
func (n *Nodis) SetEntry(data []byte) error {
	entity, err := n.parseEntry(data)
	if err != nil {
		return err
	}
	return n.store.put(entity)
}

func (n *Nodis) parseEntry(data []byte) (*pb.Entry, error) {
	c32 := binary.LittleEndian.Uint32(data)
	if c32 != crc32.ChecksumIEEE(data[4:]) {
		return nil, ErrCorruptedData
	}
	var entity pb.Entry
	if err := proto.Unmarshal(data[4:], &entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetEntry gets an entity
func (n *Nodis) GetEntry(key string) []byte {
	k, d := n.getDs(key, nil, 0)
	if d == nil {
		return nil
	}
	var entity = newEntry(key, d, k.Expiration)
	data, _ := entity.Marshal()
	return data
}

// parseDs the data
func (n *Nodis) parseDs(data []byte) (string, ds.DataStruct, int64, error) {
	var entity, err = n.parseEntry(data)
	if err != nil {
		return "", nil, 0, err
	}
	var dataStruct ds.DataStruct
	switch ds.DataType(entity.Type) {
	case ds.String:
		s := str.NewString()
		s.SetValue(entity.GetStringValue().Value)
		dataStruct = s
	case ds.ZSet:
		z := zset.NewSortedSet()
		z.SetValue(entity.GetZSetValue().Values)
		dataStruct = z
	case ds.List:
		l := list.NewDoublyLinkedList()
		l.SetValue(entity.GetListValue().Values)
		dataStruct = l
	case ds.Hash:
		h := hash.NewHashMap()
		h.SetValue(entity.GetHashValue().Values)
		dataStruct = h
	case ds.Set:
		s := set.NewSet()
		s.SetValue(entity.GetSetValue().Values)
		dataStruct = s
	}
	return entity.Key, dataStruct, entity.Expiration, nil
}

func (n *Nodis) notify(ops ...*pb.Op) {
	if len(n.watchers) == 0 {
		return
	}
	go func() {
		for _, w := range n.watchers {
			for _, op := range ops {
				if w.Matched(op.Key) {
					w.Push(op.Operation)
				}
			}
		}
	}()
}

func (n *Nodis) Watch(pattern []string, fn func(op *pb.Operation)) int {
	n.Lock()
	w := watch.NewWatcher(pattern, fn)
	n.watchers = append(n.watchers, w)
	n.Unlock()
	return len(n.watchers) - 1
}

func (n *Nodis) UnWatch(id int) {
	n.Lock()
	n.watchers = append(n.watchers[:id], n.watchers[id+1:]...)
	n.Unlock()
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
		n.HDel(op.Key, op.Operation.Field)
	case pb.OpType_HIncrBy:
		n.HIncrBy(op.Key, op.Operation.Field, op.Operation.IncrInt)
	case pb.OpType_HIncrByFloat:
		n.HIncrByFloat(op.Key, op.Operation.Field, op.Operation.IncrFloat)
	case pb.OpType_HSet:
		n.HSet(op.Key, op.Operation.Field, op.Operation.Value)
	case pb.OpType_HSetNX:
		n.HSetNX(op.Key, op.Operation.Field, op.Operation.Value)
	case pb.OpType_LInsert:
		n.LInsert(op.Key, op.Operation.Pivot, op.Operation.Value, op.Operation.Before)
	case pb.OpType_LPop:
		n.LPop(op.Key)
	case pb.OpType_LPopRPush:
		n.LPopRPush(op.Key, op.Operation.DstKey)
	case pb.OpType_LPush:
		n.LPush(op.Key, op.Operation.Values...)
	case pb.OpType_LPushX:
		n.LPushX(op.Key, op.Operation.Value)
	case pb.OpType_LRem:
		n.LRem(op.Key, op.Operation.Count, op.Operation.Value)
	case pb.OpType_LSet:
		n.LSet(op.Key, op.Operation.Index, op.Operation.Value)
	case pb.OpType_LTrim:
		n.LTrim(op.Key, op.Operation.Start, op.Operation.Stop)
	case pb.OpType_RPop:
		n.RPop(op.Key)
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
		n.Set(op.Key, op.Operation.Value, op.Operation.Expiration)
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
