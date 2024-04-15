package nodis

import (
	"errors"
	"hash/crc32"
	"log"
	"os"
	"strings"
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
	"github.com/diiyw/nodis/redis"
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
	locks       []*sync.RWMutex
}

func Open(opt *Options) *Nodis {
	if opt.FileSize == 0 {
		opt.FileSize = FileSizeGB
	}
	if opt.Filesystem == nil {
		opt.Filesystem = &fs.Memory{}
	}
	if opt.LockPoolSize == 0 {
		opt.LockPoolSize = 10240
	}
	n := &Nodis{
		options: opt,
		locks:   make([]*sync.RWMutex, opt.LockPoolSize),
	}
	for i := 0; i < opt.LockPoolSize; i++ {
		n.locks[i] = &sync.RWMutex{}
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

func fnv32(key string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		h *= 16777619
		h ^= uint32(key[i])
	}
	return h
}

func (n *Nodis) spread(hashCode uint32) *sync.RWMutex {
	tableSize := uint32(n.options.LockPoolSize)
	return n.locks[(tableSize-1)&hashCode]
}

func (n *Nodis) getLocker(key string) *sync.RWMutex {
	return n.spread(fnv32(key))
}

func (n *Nodis) writeKey(key string, newFn func() ds.DataStruct) *metadata {
	n.Lock()
	defer n.Unlock()
	locker := n.getLocker(key)
	locker.Lock()
	k, ok := n.keys.Get(key)
	if ok {
		if k.expired(time.Now().UnixMilli()) {
			if newFn == nil {
				return newEmptyMetadata(locker, true)
			}
			k.expiration = 0
		}
		d, ok := n.dataStructs.Get(key)
		if !ok {
			if newFn == nil {
				return newEmptyMetadata(locker, true)
			}
			d = newFn()
			n.dataStructs.Set(key, d)
		}
		k.changed = true
		return newMetadata(k, d, true, locker)
	}
	meta := n.fromStore(key, true, locker)
	if !meta.isOk() {
		if newFn != nil {
			d := newFn()
			if d == nil {
				return newEmptyMetadata(locker, true)
			}
			k = newKey()
			n.keys.Set(key, k)
			n.dataStructs.Set(key, d)
			n.store.put(newEntry(key, d, k.expiration))
			meta = newMetadata(k, d, true, locker)
			meta.markChanged()
		}
	}
	return meta
}

func (n *Nodis) readKey(key string) *metadata {
	locker := n.getLocker(key)
	locker.RLock()
	k, ok := n.keys.Get(key)
	if ok {
		if k.expired(time.Now().UnixMilli()) {
			return newEmptyMetadata(locker, false)
		}
		d, ok := n.dataStructs.Get(key)
		if !ok {
			return newEmptyMetadata(locker, false)
		}
		return newMetadata(k, d, false, locker)
	}
	return n.fromStore(key, false, locker)
}

func (n *Nodis) delKey(key string) {
	n.keys.Delete(key)
	n.dataStructs.Delete(key)
	n.store.remove(key)
	n.notify(pb.NewOp(pb.OpType_Del, key))
}

func (n *Nodis) fromStore(key string, writable bool, locker *sync.RWMutex) *metadata {
	// try get from store
	v, err := n.store.get(key)
	if err == nil && len(v) > 0 {
		key, d, expiration, err := n.parseDs(v)
		if err != nil {
			log.Println("Parse DataStruct:", err)
			return newEmptyMetadata(locker, writable)
		}
		if d != nil {
			n.dataStructs.Set(key, d)
			k := newKey()
			k.expiration = expiration
			n.keys.Set(key, k)
			return newMetadata(k, d, writable, locker)
		}
	}
	return newEmptyMetadata(locker, writable)
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
	now := time.Now().UnixMilli()
	recycleTime := now - n.options.RecycleDuration.Milliseconds()
	n.keys.Scan(func(key string, k *Key) bool {
		locker := n.getLocker(key)
		locker.Lock()
		defer locker.Unlock()
		if k.expired(time.Now().UnixMilli()) {
			n.dataStructs.Delete(key)
			n.keys.Delete(key)
			n.store.remove(key)
			return true
		}
		if k.lastUse != 0 && k.lastUse <= recycleTime {
			d, ok := n.dataStructs.Get(key)
			n.dataStructs.Delete(key)
			n.keys.Delete(key)
			if ok {
				k.changed = false
				// save to disk
				err := n.store.put(newEntry(key, d, k.expiration))
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
		if !k.changed || k.expired(time.Now().UnixMilli()) {
			return true
		}
		d, ok := n.dataStructs.Get(key)
		if !ok {
			return true
		}
		entries = append(entries, newEntry(key, d, k.expiration))
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

// Clear removes all keys from the store
func (n *Nodis) Clear() {
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
	meta := n.readKey(key)
	if !meta.isOk() {
		meta.commit()
		return nil
	}
	var entity = newEntry(key, meta.ds, meta.key.expiration)
	meta.commit()
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
		n.LRem(op.Key, op.Operation.Count, op.Operation.Value)
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
	return redis.Serve(addr, func(cmd redis.Value, args []redis.Value) redis.Value {
		c, ok := redisCommands[strings.ToUpper(cmd.Bulk)]
		if !ok {
			return redis.ErrorValue("Unsupported command")
		}
		return c(n, cmd, args)
	})
}
