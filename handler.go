package nodis

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/internal/geohash"
	"github.com/diiyw/nodis/internal/strings"
	"github.com/diiyw/nodis/redis"
)

func execCommand(conn *redis.Conn, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered error: ", r)
			conn.WriteError("WRONGTYPE " + r.(error).Error())
			return
		}
	}()
	if conn.State == redis.MultiNone || conn.State == redis.MultiCommit {
		fn()
		return
	}
	if conn.State&redis.MultiPrepare == redis.MultiPrepare {
		conn.Commands = append(conn.Commands, fn)
	}
	conn.WriteString("QUEUED")
}

func GetCommand(name string) func(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	switch name {
	case "HELLO":
		return hello
	case "CLIENT":
		return client
	case "CONFIG":
		return config
	case "DBSIZE":
		return dbSize
	case "PING":
		return ping
	case "ECHO":
		return echo
	case "QUIT":
		return quit
	case "FLUSHDB":
		return flushDB
	case "WATCH":
		return watchKey
	case "UNWATCH":
		return unwatchKey
	case "MULTI":
		return multi
	case "DISCARD":
		return discard
	case "EXEC":
		return exec
	case "FLUSHALL":
		return flushDB
	case "SAVE":
		return save
	case "INFO":
		return info
	case "DEL":
		return del
	case "UNLINK":
		return unlink
	case "EXISTS":
		return exists
	case "EXPIRE":
		return expire
	case "EXPIREAT":
		return expireAt
	case "KEYS":
		return keys
	case "RANDOMKEY":
		return randomKey
	case "TTL":
		return ttl
	case "PTTL":
		return pTtl
	case "PERSIST":
		return persist
	case "RENAME":
		return rename
	case "RENAMENX":
		return renameNx
	case "TYPE":
		return typ
	case "SCAN":
		return scan
	case "SET":
		return setString
	case "MSET":
		return mSet
	case "APPEND":
		return appendString
	case "SETEX":
		return setex
	case "SETNX":
		return setnx
	case "GET":
		return getString
	case "GETSET":
		return getSet
	case "MGET":
		return mGet
	case "SETRANGE":
		return setRange
	case "GETRANGE":
		return getRange
	case "STRLEN":
		return strLen
	case "INCR":
		return incr
	case "INCRBY":
		return incrBy
	case "DECR":
		return decr
	case "DECRBY":
		return decrBy
	case "INCRBYFLOAT":
		return incrByFloat
	case "SETBIT":
		return setBit
	case "GETBIT":
		return getBit
	case "BITCOUNT":
		return bitCount
	case "SADD":
		return sAdd
	case "SMOVE":
		return sMove
	case "SSCAN":
		return sScan
	case "SCARD":
		return scard
	case "SPOP":
		return sPop
	case "SDIFF":
		return sDiff
	case "SDIFFSTORE":
		return sDiffStore
	case "SINTER":
		return sInter
	case "SINTERSTORE":
		return sInterStore
	case "SUNION":
		return sUnion
	case "SUNIONSTORE":
		return sUnionStore
	case "SISMEMBER":
		return sIsMember
	case "SMEMBERS":
		return sMembers
	case "SRANDMEMBER":
		return sRandMember
	case "SREM":
		return sRem
	case "HSET":
		return hSet
	case "HGET":
		return hGet
	case "HDEL":
		return hDel
	case "HLEN":
		return hLen
	case "HKEYS":
		return hKeys
	case "HEXISTS":
		return hExists
	case "HGETALL":
		return hGetAll
	case "HINCRBY":
		return hIncrBy
	case "HINCRBYFLOAT":
		return hIncrByFloat
	case "HSETNX":
		return hSetNX
	case "HMGET":
		return hMGet
	case "HMSET":
		return hMSet
	case "HCLEAR":
		return hClear
	case "HSTRLEN":
		return hStrLen
	case "HSCAN":
		return hScan
	case "HVALS":
		return hVals
	case "LPUSH":
		return lPush
	case "RPUSH":
		return rPush
	case "LPOP":
		return lPop
	case "RPOP":
		return rPop
	case "LLEN":
		return llen
	case "LINDEX":
		return lIndex
	case "LINSERT":
		return lInsert
	case "LPUSHX":
		return lPushx
	case "RPUSHX":
		return rPushx
	case "LREM":
		return lRem
	case "LTRIM":
		return lTrim
	case "LSET":
		return lSet
	case "LRANGE":
		return lRange
	case "LPOPRPUSH":
		return lPopRPush
	case "RPOPLPUSH":
		return rPopLPush
	case "BLPOP":
		return bLPop
	case "BRPOP":
		return bRPop
	case "ZADD":
		return zAdd
	case "ZCARD":
		return zCard
	case "ZRANK":
		return zRank
	case "ZREVRANK":
		return zRevRank
	case "ZSCORE":
		return zScore
	case "ZINCRBY":
		return zIncrBy
	case "ZRANGE":
		return zRange
	case "ZREVRANGE":
		return zRevRange
	case "ZRANGEBYSCORE":
		return zRangeByScore
	case "ZREVRANGEBYSCORE":
		return zRevRangeByScore
	case "ZREM":
		return zRem
	case "ZCOUNT":
		return zCount
	case "ZREMRANGEBYRANK":
		return zRemRangeByRank
	case "ZREMRANGEBYSCORE":
		return zRemRangeByScore
	case "ZCLEAR":
		return zClear
	case "ZUNIONSTORE":
		return zUnionStore
	case "ZINTERSTORE":
		return zInterStore
	case "ZEXISTS":
		return zExists
	case "ZSCAN":
		return zScan
	case "GEOADD":
		return geoAdd
	case "GEODIST":
		return geoDist
	case "GEOHASH":
		return geoHash
	case "GEOPOS":
		return geoPos
	case "GEORADIUS":
		return geoRadius
	case "GEORADIUSBYMEMBER":
		return geoRadiusByMember
	}
	return cmdNotFound
}

func cmdNotFound(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	conn.WriteError("ERR unknown command '" + cmd.Name + "'")
}

func hello(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		conn.WriteArray(13)
		conn.WriteBulk("server")
		conn.WriteBulk("redis")
		conn.WriteBulk("version")
		conn.WriteBulk("6.0.0")
		conn.WriteBulk("proto")
		conn.WriteBulk("2")
		conn.WriteBulk("id")
		conn.WriteBulk("1")
		conn.WriteBulk("mode")
		conn.WriteBulk("standalone")
		conn.WriteBulk("role")
		conn.WriteBulk("master")
		conn.WriteBulk("modules")
	})
}

func client(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("CLIENT subcommand must be provided")
		return
	}
	execCommand(conn, func() {
		switch strings.ToUpper(cmd.Args[0]) {
		case "LIST":
			conn.WriteString("id=1 addr=" + conn.Network.RemoteAddr().String() + " fd=5 name= age=0 idle=0 flags=N db=0 sub=0 psub=0 multi=-1 qbuf=0 qbuf-free=0 obl=0 oll=0 omem=0 events=r cmd=client")
		case "SETNAME":
			conn.WriteString("OK")
		default:
			conn.WriteError("CLIENT subcommand must be provided")
		}
	})
}

func config(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("CONFIG GET requires at least two argument")
		return
	}
	execCommand(conn, func() {
		if cmd.Args[0] == "GET" {
			if len(cmd.Args) < 2 {
				conn.WriteError("CONFIG GET requires at least one argument")
				return
			}
			if strings.ToUpper(cmd.Args[1]) == "DATABASES" {
				conn.WriteArray(2)
				conn.WriteBulk("databases")
				conn.WriteBulk("0")
				return
			}
		}
		conn.WriteBulkNull()
	})
}

func dbSize(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		conn.WriteInt64(int64(n.store.metadata.Len()))
	})
}

func info(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		memStats := runtime.MemStats{}
		runtime.ReadMemStats(&memStats)
		usedMemory := strconv.FormatUint(memStats.HeapInuse+memStats.StackInuse, 10)
		pid := strconv.Itoa(os.Getpid())
		keys, expires, avgTTL := n.Keyspace()
		var keyspace = ""
		if keys > 0 {
			keyspace = "db0:keys=" + strconv.FormatInt(keys, 10) + ",expires=" + strconv.FormatInt(expires, 10) + ",avg_ttl=" + strconv.FormatInt(avgTTL, 10) + "\r\n"
		}
		conn.WriteBulk(`# Server` + "\r\n" +
			`redis_version:6.0.0` + "\r\n" +
			`os:` + runtime.GOOS + "\r\n" +
			`process_id:` + pid + "\r\n" +
			`# Memory` + "\r\n" +
			`used_memory:` + usedMemory + "\r\n" +
			`used_memory_human:` + strconv.FormatUint(memStats.HeapInuse+memStats.StackInuse/1024, 10) + "KB" + "\r\n" +
			`maxmemory:0` + "\r\n" +
			`maxmemory_human:0B` + "\r\n" +
			`maxmemory_policy:noeviction` + "\r\n" +
			`# Client` + "\r\n" +
			`maxclients:10000` + "\r\n" +
			`connected_clients:` + strconv.FormatInt(redis.ClientNum.Load(), 10) + "\r\n" +
			`# Keyspace` + "\r\n" + keyspace +
			"\r\n")
	})
}

func ping(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		if len(cmd.Args) == 0 {
			conn.WriteBulk("PONG")
			return
		}
		conn.WriteBulk(cmd.Args[0])
	})
}

func echo(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		if len(cmd.Args) == 0 {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(cmd.Args[0])
	})
}

func quit(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		conn.WriteOK()
		conn.Network.Close()
	})
}

func flushDB(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		n.Clear()
		conn.WriteOK()
	})
}

func watchKey(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if conn.State&redis.MultiPrepare == redis.MultiPrepare {
		conn.WriteError("ERR WATCH inside MULTI is not allowed")
		return
	}
	if len(cmd.Args) == 0 {
		conn.WriteError("WATCH requires at least one argument")
		return
	}
	n.Watch(conn, cmd.Args...)
	conn.WriteOK()
}

func multi(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if conn.State&redis.MultiPrepare == redis.MultiPrepare {
		conn.WriteError("ERR MULTI calls can not be nested")
		return
	}
	conn.State |= redis.MultiPrepare
	conn.WriteOK()
}

func discard(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	conn.State = redis.MultiNone
	conn.Commands = nil
	conn.WatchKeys.Clear()
	conn.WriteOK()
}

func exec(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	defer func() {
		conn.State = redis.MultiNone
		conn.Commands = nil
		conn.WatchKeys.Clear()
	}()
	if conn.State&redis.MultiPrepare != redis.MultiPrepare {
		conn.WriteError("ERR EXEC without MULTI")
		return
	}
	if conn.State&redis.MultiError == redis.MultiError {
		conn.WriteError("EXECABORT Transaction discarded because of previous errors.")
		return
	}
	if len(conn.Commands) == 0 {
		conn.WriteArray(0)
		return
	}
	var watchKeysNoChanged = true
	tx := newTx(n.store)
	defer tx.commit()
	conn.WatchKeys.Scan(func(key string, modified bool) bool {
		if modified {
			watchKeysNoChanged = false
			return false
		}
		return true
	})
	if !watchKeysNoChanged {
		conn.WriteBulkNull()
		return
	}
	conn.State |= redis.MultiCommit
	conn.WriteArray(len(conn.Commands))
	for _, command := range conn.Commands {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Recovered error: ", r)
					conn.WriteError("WRONGTYPE " + r.(error).Error())
					return
				}
			}()
			command()
		}()
	}
}

func unwatchKey(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		n.UnWatch(conn)
		conn.WriteOK()
	})
}

func del(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("DEL requires at least one argument")
		return
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.Del(cmd.Args...))
	})
}

func unlink(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("UNLINK requires at least one argument")
		return
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.Unlink(cmd.Args...))
	})
}

func exists(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("EXISTS requires at least one argument")
		return
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.Exists(cmd.Args...))
	})
}

// EXPIRE key seconds [NX | XX | GT | LT]
func expire(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("EXPIRE requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		seconds, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
		if cmd.Options.NX > 1 {
			conn.WriteInt64(n.ExpireNX(cmd.Args[0], seconds))
			return
		}
		if cmd.Options.XX > 1 {
			conn.WriteInt64(n.ExpireXX(cmd.Args[0], seconds))
			return
		}
		if cmd.Options.LT > 1 {
			conn.WriteInt64(n.ExpireLT(cmd.Args[0], seconds))
			return
		}
		if cmd.Options.GT > 1 {
			conn.WriteInt64(n.ExpireGT(cmd.Args[0], seconds))
			return
		}
		conn.WriteInt64(n.Expire(cmd.Args[0], seconds))
	})
}

// EXPIREAT key unix-time-seconds [NX | XX | GT | LT]
func expireAt(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("EXPIREAT requires at least two arguments")
		return
	}
	timestamp, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		e := time.Unix(timestamp, 0)
		if cmd.Options.NX > 1 {
			conn.WriteInt64(n.ExpireAtNX(cmd.Args[0], e))
			return
		}
		if cmd.Options.XX > 1 {
			conn.WriteInt64(n.ExpireAtXX(cmd.Args[0], e))
			return
		}
		if cmd.Options.LT > 1 {
			conn.WriteInt64(n.ExpireAtLT(cmd.Args[0], e))
			return
		}
		if cmd.Options.GT > 1 {
			conn.WriteInt64(n.ExpireAtGT(cmd.Args[0], e))
			return
		}
		conn.WriteInt64(n.ExpireAt(cmd.Args[0], e))
	})
}

func keys(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("KEYS requires at least one argument")
		return
	}
	execCommand(conn, func() {
		keys := n.Keys(cmd.Args[0])
		conn.WriteArray(len(keys))
		for _, v := range keys {
			conn.WriteBulk(v)
		}
	})
}

func ttl(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("TTL requires at least one argument")
		return
	}
	execCommand(conn, func() {
		v := n.TTL(cmd.Args[0])
		if v == -1 {
			conn.WriteInt64(-1)
			return
		}
		if v == -2 {
			conn.WriteInt64(-2)
			return
		}
		conn.WriteInt64(int64(v.Seconds()))
	})
}

func pTtl(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("PTTL requires at least one argument")
		return
	}
	execCommand(conn, func() {
		v := n.PTTL(cmd.Args[0])
		if v == -1 {
			conn.WriteInt64(-1)
			return
		}
		if v == -2 {
			conn.WriteInt64(-2)
			return
		}
		conn.WriteInt64(v)
	})
}

func persist(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("PERSIST requires at least one argument")
		return
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.Persist(cmd.Args[0]))
	})
}

func randomKey(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		key := n.RandomKey()
		if key == "" {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(key)
	})
}

func rename(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RENAME requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		oldKey := cmd.Args[0]
		newKey := cmd.Args[1]
		_ = n.Rename(oldKey, newKey)
		conn.WriteOK()
	})
}

func renameNx(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RENAMENX requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		oldKey := cmd.Args[0]
		newKey := cmd.Args[1]
		v := n.RenameNX(oldKey, newKey)
		if v == nil {
			conn.WriteInt64(1)
			return
		}
		conn.WriteInt64(0)
	})
}

func typ(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("TYPE requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		conn.WriteString(n.Type(key))
	})
}

// SCAN cursor [MATCH pattern] [COUNT count] [TYPE type]
func scan(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("SCAN requires at least one argument")
		return
	}
	cursor, err := strconv.ParseInt(cmd.Args[0], 10, 64)
	if err != nil {
		conn.WriteError("ERR cursor value is not an integer or out of range")
		return
	}
	var match = "*"
	var count int64 = 10
	var typ ds.ValueType
	if cmd.Options.MATCH > 0 {
		match = cmd.Args[cmd.Options.MATCH]
	}
	if cmd.Options.COUNT > 0 {
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.COUNT], 10, 64)
		if err != nil {
			conn.WriteError("ERR count value is not an integer or out of range")
			return
		}
		if count == 0 {
			conn.WriteError("ERR syntax error")
			return
		}
	}
	if cmd.Options.TYPE > 0 {
		typ = ds.StringToDataType(strings.ToUpper(cmd.Args[cmd.Options.TYPE]))
	}
	execCommand(conn, func() {
		nextCursor, keys := n.Scan(cursor, match, count, typ)
		conn.WriteArray(2)
		conn.WriteBulk(strconv.FormatInt(nextCursor, 10))
		conn.WriteArray(len(keys))
		for _, v := range keys {
			conn.WriteBulk(v)
		}
	})
}

// SET key value [NX | XX] [GET] [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL]
func setString(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SET requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		value := []byte(cmd.Args[1])
		var get []byte
		if cmd.Options.GET > 1 {
			get = n.Get(key)
		}
		if cmd.Options.NX > 1 {
			if !n.SetNX(key, value, cmd.Options.KEEPTTL > 1) {
				conn.WriteBulkNull()
				return
			}
		} else if cmd.Options.XX > 1 {
			if !n.SetXX(key, value, cmd.Options.KEEPTTL > 1) {
				conn.WriteBulkNull()
				return
			}
		} else {
			n.Set(key, value, cmd.Options.KEEPTTL > 1)
		}
		if cmd.Options.EX > 1 {
			seconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.EX], 10, 64)
			n.Expire(key, seconds)
			conn.WriteOK()
			return
		}
		if cmd.Options.PX > 1 {
			milliseconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.PX], 10, 64)
			if n.ExpirePX(key, milliseconds) == 0 {
				conn.WriteBulkNull()
				return
			}
			conn.WriteOK()
			return
		}
		if cmd.Options.EXAT > 1 {
			seconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.EXAT], 10, 64)
			if n.ExpireAt(key, time.Unix(seconds, 0)) == 0 {
				conn.WriteBulkNull()
				return
			}
			conn.WriteOK()
			return
		}
		if cmd.Options.PXAT > 1 {
			milliseconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.PXAT], 10, 64)
			seconds := milliseconds / 1000
			ns := (milliseconds - seconds*1000) * 1000 * 1000
			n.ExpireAt(key, time.Unix(seconds, ns))
			conn.WriteOK()
			return
		}
		if get != nil {
			conn.WriteBulk(string(get))
			return
		}
		conn.WriteOK()
	})
}

func mSet(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 || len(cmd.Args)%2 != 0 {
		conn.WriteError("MSET requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		for i := 0; i < len(cmd.Args); i += 2 {
			n.Set(cmd.Args[i], []byte(cmd.Args[i+1]), false)
		}
		conn.WriteOK()
	})
}

func appendString(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("APPEND requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		value := []byte(cmd.Args[1])
		conn.WriteInt64(n.Append(key, value))
	})
}

func setex(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("SETEX requires at least three arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		seconds, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
		value := []byte(cmd.Args[2])
		n.SetEX(key, value, seconds)
		conn.WriteOK()
	})
}

func setnx(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SETNX requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		value := []byte(cmd.Args[1])
		if n.SetNX(key, value, false) {
			conn.WriteInt64(1)
			return
		}
		conn.WriteInt64(0)
	})
}

func incr(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("INCR requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		v, err := n.Incr(key)
		if err != nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteInt64(v)
	})
}

func incrBy(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("INCRBY requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		v, err := n.IncrBy(key, value)
		if err != nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteInt64(v)
	})
}

func decr(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("DECR requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		v, err := n.Decr(key)
		if err != nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteInt64(v)
	})
}

func decrBy(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("INCRBY requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		v, err := n.DecrBy(key, value)
		if err != nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteInt64(v)
	})
}

func incrByFloat(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("INCRBYFLOAT requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value, err := redis.FormatFloat64(cmd.Args[1])
	if err != nil {
		conn.WriteError("ERR value is not a valid float")
		return
	}
	execCommand(conn, func() {
		v, err := n.IncrByFloat(key, value)
		if err != nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(strconv.FormatFloat(v, 'f', -1, 64))
	})
}

func getString(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("GET requires at least one argument")
		return
	}
	execCommand(conn, func() {
		v := n.Get(cmd.Args[0])
		if v == nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(string(v))
	})
}

func getSet(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("GETSET requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		value := []byte(cmd.Args[1])
		v := n.GetSet(key, value)
		if v == nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(string(v))
	})

}

func mGet(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("MGET requires at least one argument")
		return
	}
	execCommand(conn, func() {
		conn.WriteArray(len(cmd.Args))
		for _, v := range cmd.Args {
			value := n.Get(v)
			if value == nil {
				conn.WriteBulkNull()
				continue
			}
			conn.WriteBulk(string(value))
		}
	})
}

func setRange(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("SETRANGE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	offset, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR offset value is not an integer or out of range")
		return
	}
	value := []byte(cmd.Args[2])
	execCommand(conn, func() {
		conn.WriteInt64(n.SetRange(key, offset, value))
	})
}

func getRange(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("GETRANGE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR start value is not an integer or out of range")
		return
	}
	end, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		conn.WriteError("ERR end value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		v := n.GetRange(key, start, end)
		conn.WriteBulk(string(v))
	})
}

func strLen(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("STRLEN requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		conn.WriteInt64(n.StrLen(key))
	})
}

func setBit(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("SETBIT requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	offset, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil || offset < 0 {
		conn.WriteError("ERR offset value is not an integer or out of range")
		return
	}
	value, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil || (value != 0 && value != 1) {
		conn.WriteError("ERR bit value is not a valid integer")
		return
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.SetBit(key, offset, value == 1))
	})
}

func getBit(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("GETBIT requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	offset, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil || offset < 0 {
		conn.WriteError("ERR offset value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.GetBit(key, offset))
	})
}

func bitCount(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("BITCOUNT requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	var start, end int64
	if len(cmd.Args) > 1 {
		start, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			conn.WriteError("ERR start value is not an integer or out of range")
			return
		}
	}
	if len(cmd.Args) > 2 {
		end, err = strconv.ParseInt(cmd.Args[2], 10, 64)
		if err != nil {
			conn.WriteError("ERR end value is not an integer or out of range")
			return
		}
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.BitCount(key, start, end, cmd.Options.BIT > 2))
	})
}

func sAdd(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SADD requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		conn.WriteInt64(n.SAdd(key, cmd.Args[1:]...))
	})
}

func sMove(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("SMOVE requires at least three arguments")
		return
	}
	execCommand(conn, func() {
		src := cmd.Args[0]
		dst := cmd.Args[1]
		member := cmd.Args[2]
		if n.SMove(src, dst, member) {
			conn.WriteInt64(1)
			return
		}
		conn.WriteInt64(0)
	})
}

// SSCAN key cursor [MATCH pattern] [COUNT count]
func sScan(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 1 {
		conn.WriteError("SSCAN requires at least one argument")
		return
	}
	key := cmd.Args[0]
	cursor, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	var match = "*"
	var count int64
	if cmd.Options.MATCH > 1 {
		match = cmd.Args[cmd.Options.MATCH]
	}
	if cmd.Options.COUNT > 1 {
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.COUNT], 10, 64)
		if err != nil {
			conn.WriteError("ERR count value is not an integer or out of range")
			return
		}
	}
	execCommand(conn, func() {
		cursor, keys := n.SScan(key, cursor, match, count)
		conn.WriteArray(2)
		conn.WriteBulk(strconv.FormatInt(cursor, 10))
		conn.WriteArray(len(keys))
		for _, v := range keys {
			conn.WriteBulk(v)
		}
	})
}

func sPop(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 1 {
		conn.WriteError("SPOP requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	var count int64 = 1
	if len(cmd.Args) > 1 {
		count, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			conn.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	execCommand(conn, func() {
		results := n.SPop(key, count)
		if results == nil {
			conn.WriteBulkNull()
			return
		}
		if len(results) == 0 {
			conn.WriteBulkNull()
			return
		}
		if len(cmd.Args) <= 1 {
			conn.WriteBulk(results[0])
			return
		}
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func scard(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("SCARD requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		conn.WriteInt64(n.SCard(key))
	})
}

func sDiff(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SDIFF requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		results := n.SDiff(cmd.Args...)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func sDiffStore(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SDIFFSTORE requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		dst := cmd.Args[0]
		keys := cmd.Args[1:]
		if n.Exists(keys...) != int64(len(keys)) {
			conn.WriteInt64(0)
			return
		}
		results := n.SDiffStore(dst, keys...)
		conn.WriteInt64(results)
	})
}

func sInter(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SINTER requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		results := n.SInter(cmd.Args...)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func sInterStore(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SINTERSTORE requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		dst := cmd.Args[0]
		keys := cmd.Args[1:]
		if n.Exists(keys...) != int64(len(keys)) {
			conn.WriteInt64(0)
			return
		}
		results := n.SInterStore(dst, keys...)
		conn.WriteInt64(results)
	})
}

func sUnion(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SUNION requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		results := n.SUnion(cmd.Args...)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func sUnionStore(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SUNIONSTORE requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		dst := cmd.Args[0]
		keys := cmd.Args[1:]
		if n.Exists(keys...) == 0 {
			conn.WriteInt64(0)
			return
		}
		results := n.SUnionStore(dst, keys...)
		conn.WriteInt64(results)
	})
}

func sIsMember(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SISMEMBER requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		member := cmd.Args[1]
		var r int64 = 0
		is := n.SIsMember(key, member)
		if is {
			r = 1
		}
		conn.WriteInt64(r)
	})
}

func sMembers(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("SMEMBERS requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		results := n.SMembers(key)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func sRandMember(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("SRANDMEMBER requires at least one argument")
		return
	}
	key := cmd.Args[0]
	var count int64 = 1
	var hasCount bool
	if len(cmd.Args) > 1 {
		var err error
		count, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			conn.WriteError("ERR value is not an integer or out of range")
			return
		}
		hasCount = true
	}
	execCommand(conn, func() {
		results := n.SRandMember(key, count)
		if len(results) == 0 {
			if hasCount {
				conn.WriteArray(0)
				return
			}
			conn.WriteBulkNull()
			return
		}
		if hasCount {
			conn.WriteArray(len(results))
			for _, v := range results {
				conn.WriteBulk(v)
			}
			return
		}
		conn.WriteBulk(results[0])
	})
}

func sRem(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SREM requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		conn.WriteInt64(n.SRem(key, cmd.Args[1:]...))
	})
}

func hSet(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("HSET requires at least three arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		field := cmd.Args[1]
		value := cmd.Args[2]
		var i int64 = n.HSet(key, field, []byte(value))
		if len(cmd.Args) > 3 {
			var fields = make(map[string][]byte, len(cmd.Args)-3)
			for i := 3; i < len(cmd.Args); i += 2 {
				if i+1 >= len(cmd.Args) {
					break
				}
				fields[cmd.Args[i]] = []byte(cmd.Args[i+1])
			}
			i += n.HMSet(key, fields)
		}
		conn.WriteInt64(i)
	})
}

func hGet(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("HGET requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		field := cmd.Args[1]
		v := string(n.HGet(key, field))
		if v == "" {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(v)
	})
}

func hDel(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("HDEL requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		conn.WriteInt64(n.HDel(key, cmd.Args[1:]...))
	})
}

func hLen(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HLEN requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		conn.WriteInt64(n.HLen(key))
	})
}

func hKeys(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HKEYS requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		results := n.HKeys(key)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func hExists(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("HEXISTS requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		field := cmd.Args[1]
		is := n.HExists(key, field)
		var r int64 = 0
		if is {
			r = 1
		}
		conn.WriteInt64(r)
	})
}

func hGetAll(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HGETALL requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		results := n.HGetAll(key)
		conn.WriteArray(len(results) * 2)
		for k, v := range results {
			conn.WriteBulk(string(k))
			conn.WriteBulk(string(v))
		}
	})
}

func hIncrBy(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("HINCRBY requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	value, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		v, err := n.HIncrBy(key, field, value)
		if err != nil {
			conn.WriteError(err.Error())
			return
		}
		conn.WriteInt64(v)
	})
}

func hIncrByFloat(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("HINCRBYFLOAT requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	value, err := redis.FormatFloat64(cmd.Args[2])
	if err != nil {
		conn.WriteError("ERR value is not a valid float")
		return
	}
	execCommand(conn, func() {
		v, err := n.HIncrByFloat(key, field, value)
		if err != nil {
			conn.WriteError(err.Error())
			return
		}
		f := strconv.FormatFloat(v, 'f', -1, 64)
		conn.WriteBulk(f)
	})
}

func hSetNX(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("HSETNX requires at least three arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		field := cmd.Args[1]
		value := cmd.Args[2]
		conn.WriteInt64(n.HSetNX(key, field, []byte(value)))
	})
}

func hMGet(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("HMGET requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		fields := cmd.Args[1:]
		results := n.HMGet(key, fields...)
		if results == nil {
			conn.WriteArray(len(fields))
			for range fields {
				conn.WriteBulkNull()
			}
			return
		}
		conn.WriteArray(len(results))
		for _, v := range results {
			if v == nil {
				conn.WriteBulkNull()
			} else {
				conn.WriteBulk(string(v))
			}
		}
	})
}

func hMSet(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("HMSET requires at least three arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		fields := make(map[string][]byte, len(cmd.Args)-1)
		for i := 1; i < len(cmd.Args); i += 2 {
			if i+2 > len(cmd.Args) {
				break
			}
			fields[cmd.Args[i]] = []byte(cmd.Args[i+1])
		}
		n.HMSet(key, fields)
		conn.WriteString("OK")
	})
}

func hClear(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HCLEAR requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		n.HClear(key)
		conn.WriteString("OK")
	})
}

func hStrLen(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("HSTRLEN requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		field := cmd.Args[1]
		conn.WriteInt64(n.HStrLen(key, field))
	})
}

func hScan(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HSCAN requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	cursor, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	var match = "*"
	var count int64
	if cmd.Options.MATCH > 1 {
		match = cmd.Args[cmd.Options.MATCH]
	}
	if cmd.Options.COUNT > 1 {
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.COUNT], 10, 64)
		if err != nil {
			conn.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	execCommand(conn, func() {
		_, results := n.HScan(key, cursor, match, count)
		conn.WriteArray(2)
		conn.WriteBulk(strconv.FormatInt(cursor, 10))
		conn.WriteArray(len(results) * 2)
		for k, v := range results {
			conn.WriteBulk(k)
			conn.WriteBulk(string(v))
		}
	})
}

func hVals(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HVALS requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		results := n.HVals(key)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(string(v))
		}
	})
}

func lPush(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LPUSH requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		values := make([][]byte, 0, len(cmd.Args)-1)
		for i := 1; i < len(cmd.Args); i++ {
			values = append(values, []byte(cmd.Args[i]))
		}
		conn.WriteInt64(n.LPush(key, values...))
	})
}

func rPush(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RPUSH requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		values := make([][]byte, 0, len(cmd.Args)-1)
		for i := 1; i < len(cmd.Args); i++ {
			values = append(values, []byte(cmd.Args[i]))
		}
		conn.WriteInt64(n.RPush(key, values...))
	})
}

func lPop(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("LPOP requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	var count int64 = 1
	if len(cmd.Args) > 1 {
		count, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			conn.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	execCommand(conn, func() {
		v := n.LPop(key, count)
		if v == nil {
			conn.WriteBulkNull()
			return
		}
		if count == 1 {
			conn.WriteBulk(string(v[0]))
			return
		}
		conn.WriteArray(len(v))
		for _, vv := range v {
			conn.WriteBulk(string(vv))
		}
	})
}

func rPop(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("RPOP requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	var count int64 = 1
	if len(cmd.Args) > 1 {
		count, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			conn.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	execCommand(conn, func() {
		v := n.RPop(key, count)
		if v == nil {
			conn.WriteBulkNull()
			return
		}
		if count == 1 {
			conn.WriteBulk(string(v[0]))
			return
		}
		conn.WriteArray(len(v))
		for _, vv := range v {
			conn.WriteBulk(string(vv))
		}
	})
}

func llen(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("LLEN requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		v := n.LLen(key)
		if v == -1 {
			conn.WriteBulkNull()
			return
		}
		conn.WriteInt64(v)
	})
}

func lIndex(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LINDEX requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		index, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
		v := n.LIndex(key, index)
		if v == nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(string(v))
	})
}

func lInsert(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 4 {
		conn.WriteError("LINSERT requires at least four arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		before := strings.ToUpper(cmd.Args[1]) == "BEFORE"
		pivot := []byte(cmd.Args[2])
		value := []byte(cmd.Args[3])
		conn.WriteInt64(n.LInsert(key, pivot, value, before))
	})
}

func lPushx(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LPUSHX requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		value := []byte(cmd.Args[1])
		conn.WriteInt64(n.LPushX(key, value))
	})
}

func rPushx(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RPUSHX requires at least two arguments")
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		value := []byte(cmd.Args[1])
		conn.WriteInt64(n.RPushX(key, value))
	})
}

func lRem(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LREM requires at least two arguments")
		return
	}
	var err error
	key := cmd.Args[0]
	count, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		value := []byte(cmd.Args[2])
		conn.WriteInt64(n.LRem(key, value, count))
	})
}

func lTrim(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LTRIM requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR start value is not an integer or out of range")
		return
	}
	end, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		conn.WriteError("ERR end value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		n.LTrim(key, start, end)
		conn.WriteString("OK")
	})
}

func lSet(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("LSET requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	index, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		if index >= n.LLen(key) {
			conn.WriteError("ERR index out of range")
			return
		}
		value := []byte(cmd.Args[2])
		n.LSet(key, index, value)
		conn.WriteOK()
	})
}

func lRange(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LRANGE requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	end, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		results := n.LRange(key, start, end)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(string(v))
		}
	})
}

func lPopRPush(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LPOPRPUSH requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		source := cmd.Args[0]
		destination := cmd.Args[1]
		v := n.LPopRPush(source, destination)
		if v == nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(string(v))
	})
}

func rPopLPush(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RPOPLPUSH requires at least two arguments")
		return
	}
	execCommand(conn, func() {
		source := cmd.Args[0]
		destination := cmd.Args[1]
		v := n.RPopLPush(source, destination)
		if v == nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(string(v))
	})
}

func bLPop(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("BLPOP requires at least two arguments")
		return
	}
	keys := make([]string, 0, len(cmd.Args)-1)
	for i := 0; i < len(cmd.Args)-1; i++ {
		keys = append(keys, cmd.Args[i])
	}
	timeout, err := strconv.ParseFloat(cmd.Args[len(cmd.Args)-1], 64)
	if err != nil {
		conn.WriteError("ERR timeout value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		k, v := n.BLPop(time.Duration(timeout*time.Second.Seconds())*time.Second, keys...)
		if k == "" {
			conn.WriteArrayNull()
			return
		}
		conn.WriteArray(2)
		conn.WriteBulk(k)
		conn.WriteBulk(string(v))
	})
}

func bRPop(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("BRPOP requires at least two arguments")
		return
	}
	keys := make([]string, 0, len(cmd.Args)-1)
	for i := 0; i < len(cmd.Args)-1; i++ {
		keys = append(keys, cmd.Args[i])
	}
	timeout, err := strconv.ParseFloat(cmd.Args[len(cmd.Args)-1], 64)
	if err != nil {
		conn.WriteError("ERR timeout value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		k, v := n.BRPop(time.Duration(timeout*time.Second.Seconds())*time.Second, keys...)
		if k == "" {
			conn.WriteArrayNull()
			return
		}
		conn.WriteArray(2)
		conn.WriteBulk(k)
		conn.WriteBulk(string(v))
	})
}

// ZADD key [NX | XX] [GT | LT] [CH] [INCR] score member [score member   ...]
func zAdd(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZADD requires at least three arguments")
		return
	}
	itemStart := max(cmd.Options.NX, cmd.Options.XX, cmd.Options.LT, cmd.Options.GT, cmd.Options.CH, cmd.Options.INCR)
	if itemStart+1 > len(cmd.Args) {
		conn.WriteError("ZADD requires at least one score-member pair")
		return
	}
	itemStart++
	key := cmd.Args[0]
	var count int64 = 0
	execCommand(conn, func() {
		for i := itemStart; i < len(cmd.Args); i += 2 {
			if i+2 > len(cmd.Args) {
				break
			}
			score, err := strconv.ParseFloat(cmd.Args[i], 64)
			if err != nil {
				conn.WriteError("ERR score value is not a valid float")
				return
			}
			member := cmd.Args[i+1]
			if cmd.Options.INCR > 0 {
				score = n.ZIncrBy(key, member, score)
				conn.WriteBulk(strconv.FormatFloat(score, 'f', -1, 64))
				return
			}
			if cmd.Options.XX > 0 {
				conn.WriteInt64(n.ZAddXX(key, member, score))
				return
			}
			if cmd.Options.NX > 0 {
				conn.WriteInt64(n.ZAddNX(key, member, score))
				return
			}
			if cmd.Options.LT > 0 {
				conn.WriteInt64(n.ZAddLT(key, member, score))
				return
			}
			if cmd.Options.GT > 0 {
				conn.WriteInt64(n.ZAddGT(key, member, score))
				return
			}
			count += n.ZAdd(key, member, score)
		}
		conn.WriteInt64(count)
	})
}

func zCard(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("ZCARD requires at least one argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		conn.WriteInt64(n.ZCard(key))
	})
}

// ZRANK key member [WITHSCORE]
func zRank(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ZRANK requires at least two argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		member := cmd.Args[1]
		if cmd.Options.WITHSCORES > 1 {
			rank, el := n.ZRankWithScore(key, member)
			if el != nil {
				conn.WriteArray(2)
				conn.WriteInt64(rank)
				conn.WriteBulk(el.Member)
				return
			}
			conn.WriteBulkNull()
			return
		}
		v, err := n.ZRank(key, member)
		if err != nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteInt64(v)
	})
}

func zRevRank(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("ZREVRANK requires at least two argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		member := cmd.Args[1]
		if cmd.Options.WITHSCORES > 1 {
			rank, el := n.ZRevRankWithScore(key, member)
			if el != nil {
				conn.WriteArray(2)
				conn.WriteInt64(rank)
				conn.WriteBulk(el.Member)
				return
			}
			conn.WriteBulkNull()
			return
		}
		v, err := n.ZRevRank(key, member)
		if err != nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteInt64(v)
	})
}

func zScore(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ZSCORE requires at least two argument")
		return
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		member := cmd.Args[1]
		score, err := n.ZScore(key, member)
		if err != nil {
			conn.WriteBulkNull()
			return
		}
		conn.WriteBulk(strconv.FormatFloat(score, 'f', -1, 64))
	})
}

func zIncrBy(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZINCRBY requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	score, err := redis.FormatFloat64(cmd.Args[1])
	if err != nil {
		conn.WriteError("ERR score value is not a valid float")
		return
	}
	execCommand(conn, func() {
		member := cmd.Args[2]
		v := n.ZIncrBy(key, member, score)
		conn.WriteBulk(strconv.FormatFloat(v, 'f', -1, 64))
	})
}

func zRange(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZRANGE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	var mode int
	if cmd.Options.BYSCORE > 2 {
		if cmd.Args[1][0] == '(' {
			mode = zset.MinOpen
		}
		var min, max float64
		var err error
		if cmd.Options.REV > 2 {
			min, err = redis.FormatFloat64(cmd.Args[2])
		} else {
			min, err = redis.FormatFloat64(cmd.Args[1])
		}
		if err != nil {
			conn.WriteError("ERR start value is not an integer or out of range")
			return
		}
		if cmd.Args[2][0] == '(' {
			mode |= zset.MaxOpen
		}
		if cmd.Options.REV > 2 {
			max, err = redis.FormatFloat64(cmd.Args[1])
		} else {
			max, err = redis.FormatFloat64(cmd.Args[2])
		}
		if err != nil {
			conn.WriteError("ERR stop value is not an integer or out of range")
			return
		}
		var offset, count int64 = 0, -1
		if cmd.Options.LIMIT > 2 {
			offset, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT], 10, 64)
			if err != nil {
				conn.WriteError("ERR offset value is not an integer or out of range")
				return
			}
			count, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT+1], 10, 64)
			if err != nil {
				conn.WriteError("ERR count value is not an integer or out of range")
				return
			}
		}
		execCommand(conn, func() {
			if cmd.Options.WITHSCORES > 2 {
				if cmd.Options.REV > 2 {
					results := n.ZRevRangeByScoreWithScores(key, min, max, offset, count, mode)
					conn.WriteArray(len(results) * 2)
					for _, v := range results {
						conn.WriteBulk(v.Member)
						conn.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
					}
					return
				}
				results := n.ZRangeByScoreWithScores(key, min, max, offset, count, mode)
				conn.WriteArray(len(results) * 2)
				for _, v := range results {
					conn.WriteBulk(v.Member)
					conn.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
				}
				return
			}
			if cmd.Options.REV > 2 {
				results := n.ZRevRangeByScore(key, min, max, offset, count, mode)
				conn.WriteArray(len(results))
				for _, v := range results {
					conn.WriteBulk(v)
				}
				return
			}
			results := n.ZRangeByScore(key, min, max, offset, count, mode)
			conn.WriteArray(len(results))
			for _, v := range results {
				conn.WriteBulk(v)
			}
		})
		return
	}
	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR start value is not an integer or out of range")
		return
	}
	stop, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		conn.WriteError("ERR stop value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		if cmd.Options.WITHSCORES > 2 {
			if cmd.Options.REV > 2 {
				results := n.ZRevRangeWithScores(key, start, stop)
				conn.WriteArray(len(results) * 2)
				for _, v := range results {
					conn.WriteBulk(v.Member)
					conn.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
				}
				return
			}
			results := n.ZRangeWithScores(key, start, stop)
			conn.WriteArray(len(results) * 2)
			for _, v := range results {
				conn.WriteBulk(v.Member)
				conn.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
			}
			return
		}
		if cmd.Options.REV > 2 {
			results := n.ZRevRange(key, start, stop)
			conn.WriteArray(len(results))
			for _, v := range results {
				conn.WriteBulk(v)
			}
			return
		}
		results := n.ZRange(key, start, stop)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func zRevRange(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZREVRANGE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR start value is not an integer or out of range")
		return
	}
	stop, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		conn.WriteError("ERR stop value is not an integer or out of range")
		return
	}
	execCommand(conn, func() {
		if cmd.Options.WITHSCORES > 2 {
			results := n.ZRevRangeWithScores(key, start, stop)
			conn.WriteArray(len(results) * 2)
			for _, v := range results {
				conn.WriteBulk(v.Member)
				conn.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
			}
			return
		}
		results := n.ZRevRange(key, start, stop)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

// ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
func zRangeByScore(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZRANGEBYSCORE requires at least three arguments")
		return
	}
	var mode = 0
	key := cmd.Args[0]
	if cmd.Args[1][0] == '(' {
		mode = zset.MinOpen
	}
	min, err := redis.FormatFloat64(cmd.Args[1])
	if err != nil {
		conn.WriteError("ERR min value is not a valid float")
		return
	}
	if cmd.Args[2][0] == '(' {
		mode |= zset.MaxOpen
	}
	max, err := redis.FormatFloat64(cmd.Args[2])
	if err != nil {
		conn.WriteError("ERR max value is not a valid float")
		return
	}
	var offset, count int64 = 0, -1
	if cmd.Options.LIMIT > 2 {
		offset, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT], 10, 64)
		if err != nil {
			conn.WriteError("ERR offset value is not an integer or out of range")
			return
		}
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT+1], 10, 64)
		if err != nil {
			conn.WriteError("ERR count value is not an integer or out of range")
			return
		}
	}
	execCommand(conn, func() {
		if cmd.Options.WITHSCORES > 2 {
			results := n.ZRangeByScoreWithScores(key, min, max, offset, count, mode)
			conn.WriteArray(len(results) * 2)
			for _, v := range results {
				conn.WriteBulk(v.Member)
				conn.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
			}
			return
		}
		results := n.ZRangeByScore(key, min, max, offset, count, mode)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func zRevRangeByScore(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZREVRANGEBYSCORE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	var mode int
	if cmd.Args[2][0] == '(' {
		mode = zset.MaxOpen
	}
	min, err := redis.FormatFloat64(cmd.Args[2])
	if err != nil {
		conn.WriteError("ERR min value is not a valid float")
		return
	}
	if cmd.Args[1][0] == '(' {
		mode |= zset.MinOpen
	}
	max, err := redis.FormatFloat64(cmd.Args[1])
	if err != nil {
		conn.WriteError("ERR max value is not a valid float")
		return
	}
	var offset, count int64 = 0, -1
	if cmd.Options.LIMIT > 2 {
		offset, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT], 10, 64)
		if err != nil {
			conn.WriteError("ERR offset value is not an integer or out of range")
			return
		}
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT+1], 10, 64)
		if err != nil {
			conn.WriteError("ERR count value is not an integer or out of range")
			return
		}
	}
	execCommand(conn, func() {
		if cmd.Options.WITHSCORES > 2 {
			results := n.ZRevRangeByScoreWithScores(key, min, max, offset, count, mode)
			conn.WriteArray(len(results) * 2)
			for _, v := range results {
				conn.WriteBulk(v.Member)
				conn.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
			}
			return
		}
		results := n.ZRevRangeByScore(key, min, max, offset, count, mode)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func zCount(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZCOUNT requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	mode := 0
	if cmd.Args[1][0] == '(' {
		mode = zset.MinOpen
	}
	min, err := redis.FormatFloat64(cmd.Args[1])
	if err != nil {
		conn.WriteError("ERR min value is not a valid float")
		return
	}
	if cmd.Args[2][0] == '(' {
		mode |= zset.MaxOpen
	}
	max, err := redis.FormatFloat64(cmd.Args[2])
	if err != nil {
		conn.WriteError("ERR max value is not a valid float")
		return
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.ZCount(key, min, max, mode))
	})
}

func zRem(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ZREM requires at least two arguments")
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		conn.WriteInt64(n.ZRem(key, cmd.Args[1:]...))
	})
}

func zRemRangeByRank(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZREMRANGEBYRANK requires at least three arguments")
	}
	key := cmd.Args[0]
	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR start value is not a valid float")
		return
	}
	stop, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		conn.WriteError("ERR stop value is not a valid float")
		return
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.ZRemRangeByRank(key, start, stop))
	})
}

func zRemRangeByScore(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZREMRANGEBYSCORE requires at least three arguments")
	}
	key := cmd.Args[0]
	var mode int
	if cmd.Args[1][0] == '(' {
		mode = zset.MinOpen
	}
	min, err := redis.FormatFloat64(cmd.Args[1])
	if err != nil {
		conn.WriteError("ERR min value is not a valid float")
		return
	}
	max, err := redis.FormatFloat64(cmd.Args[2])
	if err != nil {
		conn.WriteError("ERR max value is not a valid float")
		return
	}
	if cmd.Args[2][0] == '(' {
		mode |= zset.MaxOpen
	}
	execCommand(conn, func() {
		conn.WriteInt64(n.ZRemRangeByScore(key, min, max, mode))
	})
}

func zUnionStore(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZUNIONSTORE requires at least three arguments")
		return
	}
	destination := cmd.Args[0]
	numKeys, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR numkeys value is not a valid integer")
		return
	}
	keys := cmd.Args[2 : 2+numKeys]
	var weights []float64
	var aggregate string
	if cmd.Options.WEIGHTS > 2 {
		if len(cmd.Args) < cmd.Options.WEIGHTS+int(numKeys) {
			conn.WriteError("ERR syntax error")
			return
		}
		weights = make([]float64, numKeys)
		for i := 0; i < len(weights); i++ {
			weights[i], err = redis.FormatFloat64(cmd.Args[i+cmd.Options.WEIGHTS])
			if err != nil {
				conn.WriteError("ERR weight value is not a valid float")
				return
			}
		}
	}
	if cmd.Options.AGGREGATE > 2 {
		aggregate = cmd.Args[cmd.Options.AGGREGATE]
	}
	execCommand(conn, func() {
		n.ZUnionStore(destination, keys, weights, strings.ToUpper(aggregate))
		conn.WriteInt64(n.ZCard(destination))
	})
}

func zInterStore(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZINTERSTORE requires at least three arguments")
		return
	}
	destination := cmd.Args[0]
	numKeys, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR numkeys value is not a valid integer")
		return
	}
	keys := cmd.Args[2 : 2+numKeys]
	var weights []float64
	var aggregate string
	if cmd.Options.WEIGHTS > 2 {
		if len(cmd.Args) < cmd.Options.WEIGHTS+int(numKeys) {
			conn.WriteError("ERR syntax error")
			return
		}
		weights = make([]float64, numKeys)
		for i := 0; i < len(weights); i++ {
			weights[i], err = redis.FormatFloat64(cmd.Args[i+cmd.Options.WEIGHTS])
			if err != nil {
				conn.WriteError("ERR weight value is not a valid float")
				return
			}
		}
	}
	if cmd.Options.AGGREGATE > 2 {
		aggregate = cmd.Args[cmd.Options.AGGREGATE]
	}
	execCommand(conn, func() {
		n.ZInterStore(destination, keys, weights, strings.ToUpper(aggregate))
		conn.WriteInt64(n.ZCard(destination))
	})
}

func zClear(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("ZCLEAR requires at least one argument")
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		n.ZClear(key)
		conn.WriteString("OK")
	})
}

func zExists(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ZEXISTS requires at least two arguments")
	}
	execCommand(conn, func() {
		key := cmd.Args[0]
		member := cmd.Args[1]
		is := n.ZExists(key, member)
		var r int64 = 0
		if is {
			r = 1
		}
		conn.WriteInt64(r)
	})
}

func zScan(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ZSCAN requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	cursor, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		conn.WriteError("ERR cursor value is not a valid integer")
		return
	}
	var match = "*"
	if cmd.Options.MATCH > 0 {
		match = cmd.Args[cmd.Options.MATCH]
	}
	var count int64 = 10
	if cmd.Options.COUNT > 0 {
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.COUNT], 10, 64)
		if err != nil {
			conn.WriteError("ERR count value is not a valid integer")
			return
		}
	}
	execCommand(conn, func() {
		_, results := n.ZScan(key, cursor, match, count)
		conn.WriteArray(2)
		conn.WriteBulk(strconv.FormatInt(cursor, 10))
		conn.WriteArray(len(results) * 2)
		for _, v := range results {
			conn.WriteBulk(v.Member)
			conn.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
		}
	})
}

func save(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	execCommand(conn, func() {
		n.store.flush()
		conn.WriteString("OK")
	})
}

func geoAdd(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 4 {
		conn.WriteError("GEOADD requires at least four arguments")
		return
	}
	key := cmd.Args[0]
	var items = make([]*GeoMember, 0)
	var args []string
	if cmd.Options.NX == 2 || cmd.Options.XX == 2 {
		args = cmd.Args[2:]
	} else if cmd.Options.NX == 1 || cmd.Options.XX == 1 {
		if cmd.Options.CH == 2 {
			args = cmd.Args[2:]
		} else {
			args = cmd.Args[1:]
		}
	} else {
		args = cmd.Args[1:]
	}
	if len(args) < 3 {
		conn.WriteError("GEOADD requires at least four arguments")
		return
	}
	if len(args)%3 != 0 {
		conn.WriteError("syntax error")
		return
	}
	for i := 0; i < len(args); i += 3 {
		longitude, err := redis.FormatFloat64(args[i])
		if err != nil {
			conn.WriteError("ERR longitude value is not a valid float")
			return
		}
		latitude, err := redis.FormatFloat64(args[i+1])
		if err != nil {
			conn.WriteError("ERR latitude value is not a valid float")
			return
		}
		items = append(items, &GeoMember{Member: args[i+2], Longitude: longitude, Latitude: latitude})
	}
	execCommand(conn, func() {
		if cmd.Options.NX == 1 {
			conn.WriteInt64(n.GeoAddNX(key, items...))
			return
		}
		if cmd.Options.XX == 1 {
			conn.WriteInt64(n.GeoAddXX(key, items...))
			return
		}
		v := n.GeoAdd(key, items...)
		conn.WriteInt64(v)
	})
}

func geoDist(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("GEODIST requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	member1 := cmd.Args[1]
	member2 := cmd.Args[2]
	execCommand(conn, func() {
		v, err := n.GeoDist(key, member1, member2)
		if err != nil {
			conn.WriteBulkNull()
			return
		}
		switch true {
		case cmd.Options.KM == 3:
			conn.WriteBulk(fmt.Sprintf("%0.4f", v/1000))
		case cmd.Options.MI == 3:
			conn.WriteBulk(fmt.Sprintf("%0.4f", v/1609.34))
		case cmd.Options.FT == 3:
			conn.WriteBulk(fmt.Sprintf("%0.4f", v/0.3048))
		default:
			conn.WriteBulk(fmt.Sprintf("%0.4f", v))
		}
	})
}

func geoHash(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("GEOHASH requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	execCommand(conn, func() {
		results := n.GeoHash(key, cmd.Args[1:]...)
		conn.WriteArray(len(results))
		for _, v := range results {
			conn.WriteBulk(v)
		}
	})
}

func geoPos(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("GEOPOS requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	execCommand(conn, func() {
		results := n.GeoPos(key, cmd.Args[1:]...)
		conn.WriteArray(len(results))
		for _, v := range results {
			if v == nil {
				conn.WriteBulkNull()
				continue
			}
			conn.WriteArray(2)
			conn.WriteBulk(strconv.FormatFloat(v.Longitude, 'f', -1, 64))
			conn.WriteBulk(strconv.FormatFloat(v.Latitude, 'f', -1, 64))
		}
	})
}

func geoRadius(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 4 {
		conn.WriteError("GEORADIUS requires at least four arguments")
		return
	}
	key := cmd.Args[0]
	longitude, err := redis.FormatFloat64(cmd.Args[1])
	if err != nil {
		conn.WriteError("ERR longitude value is not a valid float")
		return
	}
	latitude, err := redis.FormatFloat64(cmd.Args[2])
	if err != nil {
		conn.WriteError("ERR latitude value is not a valid float")
		return
	}
	radius, err := redis.FormatFloat64(cmd.Args[3])
	if err != nil {
		conn.WriteError("ERR radius value is not a valid float")
		return
	}
	if cmd.Options.KM > 3 {
		radius *= 1000
	}
	if cmd.Options.MI > 3 {
		radius *= 1609.34
	}
	if cmd.Options.FT > 3 {
		radius *= 0.3048
	}
	var count int64 = -1
	if cmd.Options.COUNT > 3 && cmd.Options.ANY == 0 {
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.COUNT], 10, 64)
		if err != nil {
			conn.WriteError("ERR count value is not a valid integer")
			return
		}
	}
	execCommand(conn, func() {
		var results []*GeoMember
		var err error
		results, err = n.GeoRadius(key, longitude, latitude, radius, count, cmd.Options.DESC > 3)
		if err != nil {
			conn.WriteArrayNull()
			return
		}
		if cmd.Options.WITHCOORD == 0 && cmd.Options.WITHDIST == 0 && cmd.Options.WITHHASH == 0 {
			conn.WriteArray(len(results))
			for _, v := range results {
				conn.WriteBulk(v.Member)
			}
			return
		}
		conn.WriteArray(len(results))
		l := 1
		if cmd.Options.WITHCOORD > 3 {
			l++
		}
		if cmd.Options.WITHDIST > 3 {
			l++
		}
		if cmd.Options.WITHHASH > 3 {
			l++
		}
		for _, v := range results {
			conn.WriteArray(l)
			conn.WriteBulk(v.Member)
			if cmd.Options.WITHDIST > 3 {
				h, _ := geohash.EncodeWGS84(longitude, latitude)
				dist := geohash.DistBetweenGeoHashWGS84(h, v.Hash())
				if cmd.Options.KM > 3 {
					conn.WriteBulk(fmt.Sprintf("%0.4f", dist/1000))
				} else if cmd.Options.MI > 3 {
					conn.WriteBulk(fmt.Sprintf("%0.4f", dist/1609.34))
				} else if cmd.Options.FT > 3 {
					conn.WriteBulk(fmt.Sprintf("%0.4f", dist/0.3048))
				} else {
					conn.WriteBulk(fmt.Sprintf("%0.4f", dist))
				}
			}
			if cmd.Options.WITHHASH > 3 {
				conn.WriteUInt64(v.Hash())
			}
			if cmd.Options.WITHCOORD > 3 {
				conn.WriteArray(2)
				conn.WriteBulk(strconv.FormatFloat(v.Longitude, 'f', -1, 64))
				conn.WriteBulk(strconv.FormatFloat(v.Latitude, 'f', -1, 64))
			}
		}
	})
}

func geoRadiusByMember(n *Nodis, conn *redis.Conn, cmd redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("GEORADIUSBYMEMBER requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	radius, err := redis.FormatFloat64(cmd.Args[2])
	if err != nil {
		conn.WriteError("ERR radius value is not a valid float")
		return
	}
	if cmd.Options.KM > 3 {
		radius *= 1000
	}
	if cmd.Options.MI > 3 {
		radius *= 1609.34
	}
	if cmd.Options.FT > 3 {
		radius *= 0.3048
	}
	var count int64 = -1
	if cmd.Options.COUNT > 3 && cmd.Options.ANY == 0 {
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.COUNT], 10, 64)
		if err != nil {
			conn.WriteError("ERR count value is not a valid integer")
			return
		}
	}
	execCommand(conn, func() {
		var results []*GeoMember
		var err error
		results, err = n.GeoRadiusByMember(key, member, radius, count, cmd.Options.DESC > 3)
		if err != nil {
			conn.WriteArrayNull()
			return
		}
		if cmd.Options.WITHCOORD == 0 && cmd.Options.WITHDIST == 0 && cmd.Options.WITHHASH == 0 {
			conn.WriteArray(len(results))
			for _, v := range results {
				conn.WriteBulk(v.Member)
			}
			return
		}
		conn.WriteArray(len(results))
		l := 1
		if cmd.Options.WITHCOORD > 3 {
			l++
		}
		if cmd.Options.WITHDIST > 3 {
			l++
		}
		if cmd.Options.WITHHASH > 3 {
			l++
		}
		for _, v := range results {
			conn.WriteArray(l)
			conn.WriteBulk(v.Member)
			if cmd.Options.WITHDIST > 3 {
				h, _ := geohash.EncodeWGS84(v.Longitude, v.Latitude)
				dist := geohash.DistBetweenGeoHashWGS84(h, v.Hash())
				if cmd.Options.KM > 3 {
					conn.WriteBulk(fmt.Sprintf("%0.4f", dist/1000))
				}
				if cmd.Options.MI > 3 {
					conn.WriteBulk(fmt.Sprintf("%0.4f", dist/1609.34))
				}
				if cmd.Options.FT > 3 {
					conn.WriteBulk(fmt.Sprintf("%0.4f", dist/0.3048))
				}
				conn.WriteBulk(fmt.Sprintf("%0.4f", dist))
			}
			if cmd.Options.WITHHASH > 3 {
				conn.WriteUInt64(v.Hash())
			}
			if cmd.Options.WITHCOORD > 3 {
				conn.WriteArray(2)
				conn.WriteBulk(strconv.FormatFloat(v.Longitude, 'f', -1, 64))
				conn.WriteBulk(strconv.FormatFloat(v.Latitude, 'f', -1, 64))
			}
		}
	})
}
