package nodis

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/redis"
	"github.com/diiyw/nodis/utils"
)

func getCommand(name string) func(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	switch name {
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
	case "FLUSHALL":
		return flushDB
	case "SAVE":
		return save
	case "INFO":
		return info
	case "DEL":
		return del
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
	case "PERSIST":
		return Persist
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
	case "MGET":
		return mGet
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
	}
	return nil
}

func client(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("CLIENT subcommand must be provided")
		return
	}
	switch utils.ToUpper(cmd.Args[0]) {
	case "LIST":
		conn.WriteString("id=1 addr=" + conn.Network.RemoteAddr().String() + " fd=5 name= age=0 idle=0 flags=N db=0 sub=0 psub=0 multi=-1 qbuf=0 qbuf-free=0 obl=0 oll=0 omem=0 events=r cmd=client")
	case "SETNAME":
		conn.WriteString("OK")
	default:
		conn.WriteError("CLIENT subcommand must be provided")
	}
}

func config(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("CONFIG GET requires at least two argument")
		return
	}
	if cmd.Options.GET > 0 {
		if cmd.Options.GET > len(cmd.Args) {
			conn.WriteError("CONFIG GET requires at least one argument")
			return
		}
		if utils.ToUpper(cmd.Args[cmd.Options.GET]) == "DATABASES" {
			conn.WriteArray(2)
			conn.WriteBulk("databases")
			conn.WriteBulk("0")
			return
		}
	}
	conn.WriteNull()
}

func dbSize(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	conn.WriteInteger(int64(n.store.keys.Len()))
}

func info(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	usedMemory := strconv.FormatUint(memStats.HeapInuse+memStats.StackInuse, 10)
	pid := strconv.Itoa(os.Getpid())
	conn.WriteBulk(`# Server
redis_version:1.6.0
os:` + runtime.GOOS + `
process_id:` + pid + `
# Memory
used_memory:` + usedMemory + `
`)
}

func ping(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteBulk("PONG")
		return
	}
	conn.WriteBulk(cmd.Args[0])
}

func echo(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteNull()
		return
	}
	conn.WriteBulk(cmd.Args[0])
}

func quit(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	conn.WriteOK()
	conn.Network.Close()
}

func flushDB(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	n.Clear()
	conn.WriteOK()
}

func del(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("DEL requires at least one argument")
		return
	}
	conn.WriteInteger(n.Del(cmd.Args...))
}

func exists(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("EXISTS requires at least one argument")
		return
	}
	conn.WriteInteger(n.Exists(cmd.Args...))
}

// EXPIRE key seconds [NX | XX | GT | LT]
func expire(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("EXPIRE requires at least two arguments")
		return
	}
	seconds, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	if cmd.Options.NX > 0 {
		conn.WriteInteger(n.ExpireNX(cmd.Args[0], seconds))
		return
	}
	if cmd.Options.XX > 0 {
		conn.WriteInteger(n.ExpireXX(cmd.Args[0], seconds))
		return
	}
	if cmd.Options.LT > 0 {
		conn.WriteInteger(n.ExpireLT(cmd.Args[0], seconds))
		return
	}
	if cmd.Options.GT > 0 {
		conn.WriteInteger(n.ExpireGT(cmd.Args[0], seconds))
		return
	}
	conn.WriteInteger(n.Expire(cmd.Args[0], seconds))
}

// EXPIREAT key unix-time-seconds [NX | XX | GT | LT]
func expireAt(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("EXPIREAT requires at least two arguments")
		return
	}
	timestamp, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	e := time.Unix(timestamp, 0)
	if cmd.Options.NX > 0 {
		conn.WriteInteger(n.ExpireAtNX(cmd.Args[0], e))
		return
	}
	if cmd.Options.XX > 0 {
		conn.WriteInteger(n.ExpireAtXX(cmd.Args[0], e))
		return
	}
	if cmd.Options.LT > 0 {
		conn.WriteInteger(n.ExpireAtLT(cmd.Args[0], e))
		return
	}
	if cmd.Options.GT > 0 {
		conn.WriteInteger(n.ExpireAtGT(cmd.Args[0], e))
		return
	}
	conn.WriteInteger(n.ExpireAt(cmd.Args[0], e))
}

func keys(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("KEYS requires at least one argument")
		return
	}
	keys := n.Keys(cmd.Args[0])
	conn.WriteArray(len(keys))
	for _, v := range keys {
		conn.WriteBulk(v)
	}
}

func ttl(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("TTL requires at least one argument")
		return
	}
	v := n.TTL(cmd.Args[0])
	if v == -1 {
		conn.WriteInteger(-1)
		return
	}
	if v == -2 {
		conn.WriteInteger(-2)
		return
	}
	conn.WriteInteger(int64(v.Seconds()))
}

func Persist(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("PERSIST requires at least one argument")
		return
	}
	conn.WriteInteger(n.Persist(cmd.Args[0]))
}

func randomKey(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	key := n.RandomKey()
	if key == "" {
		conn.WriteNull()
		return
	}
	conn.WriteBulk(key)
}

func rename(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RENAME requires at least two arguments")
		return
	}
	oldKey := cmd.Args[0]
	newKey := cmd.Args[1]
	v := n.Rename(oldKey, newKey)
	if v == nil {
		conn.WriteInteger(1)
		return
	}
	conn.WriteInteger(0)
}

func renameNx(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RENAMENX requires at least two arguments")
		return
	}
	oldKey := cmd.Args[0]
	newKey := cmd.Args[1]
	v := n.RenameNX(oldKey, newKey)
	if v == nil {
		conn.WriteInteger(1)
		return
	}
	conn.WriteInteger(0)
}

func typ(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("TYPE requires at least one argument")
		return
	}
	key := cmd.Args[0]
	conn.WriteString(n.Type(key))
}

// SCAN cursor [MATCH pattern] [COUNT count] [TYPE type]
func scan(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	var count int64
	if cmd.Options.MATCH > 0 {
		match = cmd.Args[cmd.Options.MATCH]
	}
	if cmd.Options.COUNT > 0 {
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.COUNT], 10, 64)
		if err != nil {
			conn.WriteError("ERR count value is not an integer or out of range")
			return
		}
	}
	nextCursor, keys := n.Scan(cursor, match, count)
	conn.WriteArray(2)
	conn.WriteBulk(strconv.FormatInt(nextCursor, 10))
	conn.WriteArray(len(keys))
	for _, v := range keys {
		conn.WriteBulk(v)
	}
}

// SET key value [NX | XX] [GET] [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL]
func setString(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SET requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value := []byte(cmd.Args[1])
	var get []byte
	if cmd.Options.GET > 0 {
		get = n.Get(key)
	}
	if cmd.Options.NX > 0 {
		n.SetNX(key, value)
	} else if cmd.Options.XX > 0 {
		n.SetXX(key, value)
	} else {
		n.Set(key, value)
	}
	if cmd.Options.EX > 0 {
		seconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.EX], 10, 64)
		n.Expire(key, seconds)
		conn.WriteOK()
		return
	}
	if cmd.Options.PX > 0 {
		milliseconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.PX], 10, 64)
		n.ExpirePX(key, milliseconds)
		conn.WriteOK()
		return
	}
	if cmd.Options.EXAT > 0 {
		seconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.EXAT], 10, 64)
		n.ExpireAt(key, time.Unix(seconds, 0))
		conn.WriteOK()
		return
	}
	if cmd.Options.PXAT > 0 {
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
}

func mSet(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 || len(cmd.Args)%2 != 0 {
		conn.WriteError("MSET requires at least two arguments")
		return
	}
	for i := 0; i < len(cmd.Args); i += 2 {
		n.Set(cmd.Args[i], []byte(cmd.Args[i+1]))
	}
	conn.WriteOK()
}

func appendString(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("APPEND requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value := []byte(cmd.Args[1])
	conn.WriteInteger(n.Append(key, value))
}

func setex(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("SETEX requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	seconds, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	value := []byte(cmd.Args[2])
	n.SetEX(key, value, seconds)
	conn.WriteOK()
}

func setnx(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SETNX requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value := []byte(cmd.Args[1])
	if n.SetNX(key, value) {
		conn.WriteInteger(1)
		return
	}
	conn.WriteInteger(0)
}

func incr(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("INCR requires at least one argument")
		return
	}
	key := cmd.Args[0]
	v, err := n.Incr(key)
	if err != nil {
		conn.WriteNull()
		return
	}
	conn.WriteInteger(v)
}

func incrBy(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	v, err := n.IncrBy(key, value)
	if err != nil {
		conn.WriteNull()
		return
	}
	conn.WriteInteger(v)
}

func decr(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("DECR requires at least one argument")
		return
	}
	key := cmd.Args[0]
	v, err := n.Decr(key)
	if err != nil {
		conn.WriteNull()
		return
	}
	conn.WriteInteger(v)
}

func decrBy(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	v, err := n.DecrBy(key, value)
	if err != nil {
		conn.WriteNull()
		return
	}
	conn.WriteInteger(v)
}

func incrByFloat(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("INCRBYFLOAT requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value, err := redis.FormatFloat64(cmd.Args[1], 0)
	if err != nil {
		conn.WriteError("ERR value is not a valid float")
		return
	}
	v, err := n.IncrByFloat(key, value)
	if err != nil {
		conn.WriteNull()
		return
	}
	conn.WriteBulk(strconv.FormatFloat(v, 'f', -1, 64))
}

func getString(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("GET requires at least one argument")
		return
	}
	v := n.Get(cmd.Args[0])
	if v == nil {
		conn.WriteNull()
		return
	}
	conn.WriteBulk(string(v))
}

func mGet(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("MGET requires at least one argument")
		return
	}
	conn.WriteArray(len(cmd.Args))
	for _, v := range cmd.Args {
		value := n.Get(v)
		if value == nil {
			conn.WriteNull()
			continue
		}
		conn.WriteBulk(string(value))
	}
}

func getRange(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	v := n.GetRange(key, start, end)
	conn.WriteBulk(string(v))
}

func strLen(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("STRLEN requires at least one argument")
		return
	}
	key := cmd.Args[0]
	conn.WriteInteger(n.StrLen(key))
}

func setBit(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	conn.WriteInteger(n.SetBit(key, offset, value == 1))
}

func getBit(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	conn.WriteInteger(n.GetBit(key, offset))
}

func bitCount(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	conn.WriteInteger(n.BitCount(key, start, end, cmd.Options.BIT > 0))
}

func sAdd(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SADD requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	conn.WriteInteger(n.SAdd(key, cmd.Args[1:]...))
}

func sMove(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("SMOVE requires at least three arguments")
		return
	}
	src := cmd.Args[0]
	dst := cmd.Args[1]
	member := cmd.Args[2]
	if n.SMove(src, dst, member) {
		conn.WriteInteger(1)
		return
	}
	conn.WriteInteger(0)
}

// SSCAN key cursor [MATCH pattern] [COUNT count]
func sScan(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	if cmd.Options.MATCH > 0 {
		match = cmd.Args[cmd.Options.MATCH]
	}
	if cmd.Options.COUNT > 0 {
		count, _ = strconv.ParseInt(cmd.Args[cmd.Options.COUNT], 10, 64)
	}
	cursor, keys := n.SScan(key, cursor, match, count)
	conn.WriteArray(2)
	conn.WriteBulk(strconv.FormatInt(cursor, 10))
	conn.WriteArray(len(keys))
	for _, v := range keys {
		conn.WriteBulk(v)
	}
}

func sPop(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	results := n.SPop(key, count)
	if results == nil {
		conn.WriteNull()
		return
	}
	if len(results) == 0 {
		conn.WriteNull()
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
}

func scard(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("SCARD requires at least one argument")
		return
	}
	key := cmd.Args[0]
	conn.WriteInteger(n.SCard(key))
}

func sDiff(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SDIFF requires at least two arguments")
		return
	}
	results := n.SDiff(cmd.Args...)
	conn.WriteArray(len(results))
	for _, v := range results {
		conn.WriteBulk(v)
	}
}

func sDiffStore(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SDIFFSTORE requires at least two arguments")
		return
	}
	dst := cmd.Args[0]
	keys := cmd.Args[1:]
	if n.Exists(keys...) != int64(len(keys)) {
		conn.WriteInteger(0)
		return
	}
	results := n.SDiffStore(dst, keys...)
	conn.WriteInteger(results)
}

func sInter(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SINTER requires at least two arguments")
		return
	}
	results := n.SInter(cmd.Args...)
	conn.WriteArray(len(results))
	for _, v := range results {
		conn.WriteBulk(v)
	}
}

func sInterStore(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SINTERSTORE requires at least two arguments")
		return
	}
	dst := cmd.Args[0]
	keys := cmd.Args[1:]
	if n.Exists(keys...) != int64(len(keys)) {
		conn.WriteInteger(0)
		return
	}
	results := n.SInterStore(dst, keys...)
	conn.WriteInteger(results)
}

func sUnion(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SUNION requires at least two arguments")
		return
	}
	results := n.SUnion(cmd.Args...)
	conn.WriteArray(len(results))
	for _, v := range results {
		conn.WriteBulk(v)
	}
}

func sUnionStore(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SUNIONSTORE requires at least two arguments")
		return
	}
	dst := cmd.Args[0]
	keys := cmd.Args[1:]
	if n.Exists(keys...) == 0 {
		conn.WriteInteger(0)
		return
	}
	results := n.SUnionStore(dst, keys...)
	conn.WriteInteger(results)
}
func sIsMember(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SISMEMBER requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	var r int64 = 0
	is := n.SIsMember(key, member)
	if is {
		r = 1
	}
	conn.WriteInteger(r)
}

func sMembers(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("SMEMBERS requires at least one argument")
		return
	}
	key := cmd.Args[0]
	results := n.SMembers(key)
	conn.WriteArray(len(results))
	for _, v := range results {
		conn.WriteBulk(v)
	}
}

func sRandMember(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("SRANDMEMBER requires at least one argument")
		return
	}
	key := cmd.Args[0]
	var count int64 = 1
	if len(cmd.Args) > 1 {
		var err error
		count, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			conn.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	results := n.SRandMember(key, count)
	if len(results) == 0 {
		conn.WriteNull()
		return
	}
	if count == 1 {
		conn.WriteBulk(results[0])
		return
	}
	conn.WriteArray(len(results))
	for _, v := range results {
		conn.WriteBulk(v)
	}
}

func sRem(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("SREM requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	conn.WriteInteger(n.SRem(key, cmd.Args[1:]...))
}

func hSet(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("HSET requires at least three arguments")
		return
	}
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
	conn.WriteInteger(i)
}

func hGet(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("HGET requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	v := string(n.HGet(key, field))
	if v == "" {
		conn.WriteNull()
		return
	}
	conn.WriteBulk(v)
}

func hDel(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("HDEL requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	conn.WriteInteger(n.HDel(key, cmd.Args[1:]...))
}

func hLen(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HLEN requires at least one argument")
		return
	}
	key := cmd.Args[0]
	conn.WriteInteger(n.HLen(key))
}

func hKeys(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HKEYS requires at least one argument")
		return
	}
	key := cmd.Args[0]
	results := n.HKeys(key)
	conn.WriteArray(len(results))
	for _, v := range results {
		conn.WriteBulk(v)
	}
}

func hExists(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("HEXISTS requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	is := n.HExists(key, field)
	var r int64 = 0
	if is {
		r = 1
	}
	conn.WriteInteger(r)
}

func hGetAll(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HGETALL requires at least one argument")
		return
	}
	key := cmd.Args[0]
	results := n.HGetAll(key)
	conn.WriteArray(len(results) * 2)
	for k, v := range results {
		conn.WriteBulk(string(k))
		conn.WriteBulk(string(v))
	}
}

func hIncrBy(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	v, err := n.HIncrBy(key, field, value)
	if err != nil {
		conn.WriteError(err.Error())
		return
	}
	conn.WriteInteger(v)
}

func hIncrByFloat(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("HINCRBYFLOAT requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	value, err := redis.FormatFloat64(cmd.Args[2], 0)
	if err != nil {
		conn.WriteError("ERR value is not a valid float")
		return
	}
	v, err := n.HIncrByFloat(key, field, value)
	if err != nil {
		conn.WriteError(err.Error())
		return
	}
	f := strconv.FormatFloat(v, 'f', -1, 64)
	conn.WriteBulk(f)
}

func hSetNX(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("HSETNX requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	value := cmd.Args[2]
	conn.WriteInteger(n.HSetNX(key, field, []byte(value)))
}

func hMGet(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("HMGET requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	fields := cmd.Args[1:]
	results := n.HMGet(key, fields...)
	conn.WriteArray(len(results))
	for _, v := range results {
		if v == nil {
			conn.WriteNull()
		} else {
			conn.WriteBulk(string(v))
		}
	}
}

func hMSet(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("HMSET requires at least three arguments")
		return
	}
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
}

func hClear(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HCLEAR requires at least one argument")
		return
	}
	key := cmd.Args[0]
	n.HClear(key)
	conn.WriteString("OK")
}

func hScan(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HSCAN requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	cursor, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	var match = "*"
	var count int64
	if cmd.Options.MATCH > 0 {
		match = cmd.Args[cmd.Options.MATCH]
	}
	if cmd.Options.COUNT > 0 {
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.COUNT], 10, 64)
		if err != nil {
			conn.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	_, results := n.HScan(key, cursor, match, count)
	conn.WriteArray(2)
	conn.WriteBulk(strconv.FormatInt(cursor, 10))
	conn.WriteArray(len(results) * 2)
	for k, v := range results {
		conn.WriteBulk(k)
		conn.WriteBulk(string(v))
	}
}

func hVals(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("HVALS requires at least one argument")
		return
	}
	key := cmd.Args[0]
	results := n.HVals(key)
	conn.WriteArray(len(results))
	for _, v := range results {
		conn.WriteBulk(string(v))
	}
}

func lPush(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LPUSH requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	values := make([][]byte, 0, len(cmd.Args)-1)
	for i := 1; i < len(cmd.Args); i++ {
		values = append(values, []byte(cmd.Args[i]))
	}
	conn.WriteInteger(n.LPush(key, values...))
}

func rPush(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RPUSH requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	values := make([][]byte, 0, len(cmd.Args)-1)
	for i := 1; i < len(cmd.Args); i++ {
		values = append(values, []byte(cmd.Args[i]))
	}

	conn.WriteInteger(n.RPush(key, values...))
}

func lPop(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	v := n.LPop(key, count)
	if v == nil {
		conn.WriteNull()
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
}

func rPop(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	v := n.RPop(key, count)
	if v == nil {
		conn.WriteNull()
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
}

func llen(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("LLEN requires at least one argument")
		return
	}
	key := cmd.Args[0]
	v := n.LLen(key)
	if v == -1 {
		conn.WriteNull()
		return
	}
	conn.WriteInteger(v)
}

func lIndex(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LINDEX requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	index, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	v := n.LIndex(key, index)
	if v == nil {
		conn.WriteNull()
		return
	}
	conn.WriteBulk(string(v))
}

func lInsert(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 4 {
		conn.WriteError("LINSERT requires at least four arguments")
		return
	}
	key := cmd.Args[0]
	before := utils.ToUpper(cmd.Args[1]) == "BEFORE"
	pivot := []byte(cmd.Args[2])
	value := []byte(cmd.Args[3])
	conn.WriteInteger(n.LInsert(key, pivot, value, before))
}

func lPushx(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LPUSHX requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value := []byte(cmd.Args[1])
	conn.WriteInteger(n.LPushX(key, value))
}

func rPushx(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RPUSHX requires at least two arguments")
	}
	key := cmd.Args[0]
	value := []byte(cmd.Args[1])
	conn.WriteInteger(n.RPushX(key, value))
}

func lRem(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	value := []byte(cmd.Args[2])
	conn.WriteInteger(n.LRem(key, value, count))
}

func lTrim(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	n.LTrim(key, start, end)
	conn.WriteString("OK")
}

func lSet(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	value := []byte(cmd.Args[2])
	n.LSet(key, index, value)
	conn.WriteString("OK")
}

func lRange(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	results := n.LRange(key, start, end)
	conn.WriteArray(len(results))
	for _, v := range results {
		conn.WriteBulk(string(v))
	}
}

func lPopRPush(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("LPOPRPUSH requires at least two arguments")
		return
	}
	source := cmd.Args[0]
	destination := cmd.Args[1]
	v := n.LPopRPush(source, destination)
	if v == nil {
		conn.WriteNull()
		return
	}
	conn.WriteBulk(string(v))
}

func rPopLPush(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("RPOPLPUSH requires at least two arguments")
		return
	}
	source := cmd.Args[0]
	destination := cmd.Args[1]
	v := n.RPopLPush(source, destination)
	if v == nil {
		conn.WriteNull()
		return
	}
	conn.WriteBulk(string(v))
}

func bLPop(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("BLPOP requires at least two arguments")
		return
	}
	keys := make([]string, 0, len(cmd.Args)-1)
	for i := 0; i < len(cmd.Args)-1; i++ {
		keys = append(keys, cmd.Args[i])
	}
	timeout, err := strconv.ParseInt(cmd.Args[len(cmd.Args)-1], 10, 64)
	if err != nil {
		conn.WriteError("ERR timeout value is not an integer or out of range")
		return
	}
	k, v := n.BLPop(time.Duration(timeout)*time.Second, keys...)
	if k == "" {
		conn.WriteNull()
		return
	}
	conn.WriteArray(2)
	conn.WriteBulk(k)
	conn.WriteBulk(string(v))
}

func bRPop(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("BRPOP requires at least two arguments")
		return
	}
	keys := make([]string, 0, len(cmd.Args)-1)
	for i := 0; i < len(cmd.Args)-1; i++ {
		keys = append(keys, cmd.Args[i])
	}
	timeout, err := strconv.ParseInt(cmd.Args[len(cmd.Args)-1], 10, 64)
	if err != nil {
		conn.WriteError("ERR timeout value is not an integer or out of range")
		return
	}
	k, v := n.BRPop(time.Duration(timeout)*time.Second, keys...)
	if k == "" {
		conn.WriteNull()
		return
	}
	conn.WriteArray(2)
	conn.WriteBulk(k)
	conn.WriteBulk(string(v))
}

// ZADD key [NX | XX] [GT | LT] [CH] [INCR] score member [score member   ...]
func zAdd(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
			conn.WriteInteger(n.ZAddXX(key, member, score))
			return
		}
		if cmd.Options.NX > 0 {
			conn.WriteInteger(n.ZAddNX(key, member, score))
			return
		}
		if cmd.Options.LT > 0 {
			conn.WriteInteger(n.ZAddLT(key, member, score))
			return
		}
		if cmd.Options.GT > 0 {
			conn.WriteInteger(n.ZAddGT(key, member, score))
			return
		}
		count += n.ZAdd(key, member, score)
	}
	conn.WriteInteger(count)
}

func zCard(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("ZCARD requires at least one argument")
		return
	}
	key := cmd.Args[0]
	conn.WriteInteger(n.ZCard(key))
}

// ZRANK key member [WITHSCORE]
func zRank(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ZRANK requires at least two argument")
		return
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	if cmd.Options.WITHSCORES > 0 {
		rank, el := n.ZRankWithScore(key, member)
		if el != nil {
			conn.WriteArray(2)
			conn.WriteInteger(rank)
			conn.WriteBulk(el.Member)
			return
		}
		conn.WriteNull()
		return
	}
	conn.WriteInteger(n.ZRank(key, member))
}

func zRevRank(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("ZREVRANK requires at least two argument")
		return
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	if cmd.Options.WITHSCORES > 0 {
		rank, el := n.ZRevRankWithScore(key, member)
		if el != nil {
			conn.WriteArray(2)
			conn.WriteInteger(rank)
			conn.WriteBulk(el.Member)
			return
		}
		conn.WriteNull()
		return
	}
	conn.WriteInteger(n.ZRevRank(key, member))
}

func zScore(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ZSCORE requires at least two argument")
		return
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	score := n.ZScore(key, member)
	conn.WriteBulk(strconv.FormatFloat(score, 'f', -1, 64))
}

func zIncrBy(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZINCRBY requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	score, err := redis.FormatFloat64(cmd.Args[1], 0)
	if err != nil {
		conn.WriteError("ERR score value is not a valid float")
		return
	}
	member := cmd.Args[2]
	v := n.ZIncrBy(key, member, score)
	conn.WriteBulk(strconv.FormatFloat(v, 'f', -1, 64))
}

func zRange(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZRANGE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	var mode int
	if cmd.Options.BYSCORE > 0 {
		if cmd.Args[1][0] == '(' {
			mode = zset.MinOpen
		}
		var min, max float64
		var err error
		if cmd.Options.REV > 0 {
			min, err = redis.FormatFloat64(cmd.Args[2], n.ZMin(key).Score)
		} else {
			min, err = redis.FormatFloat64(cmd.Args[1], n.ZMin(key).Score)
		}
		if err != nil {
			conn.WriteError("ERR start value is not an integer or out of range")
			return
		}
		if cmd.Args[2][0] == '(' {
			mode |= zset.MaxOpen
		}
		if cmd.Options.REV > 0 {
			max, err = redis.FormatFloat64(cmd.Args[1], n.ZMax(key).Score)
		} else {
			max, err = redis.FormatFloat64(cmd.Args[2], n.ZMax(key).Score)
		}
		if err != nil {
			conn.WriteError("ERR stop value is not an integer or out of range")
			return
		}
		var offset, count int64 = 0, -1
		if cmd.Options.LIMIT > 0 {
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
		if cmd.Options.WITHSCORES > 0 {
			if cmd.Options.REV > 0 {
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
		if cmd.Options.REV > 0 {
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
	if cmd.Options.WITHSCORES > 0 {
		if cmd.Options.REV > 0 {
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
	if cmd.Options.REV > 0 {
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
}

func zRevRange(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	if cmd.Options.WITHSCORES > 0 {
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
}

// ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
func zRangeByScore(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZRANGEBYSCORE requires at least three arguments")
		return
	}
	var mode = 0
	key := cmd.Args[0]
	if cmd.Args[1][0] == '(' {
		mode = zset.MinOpen
	}
	min, err := redis.FormatFloat64(cmd.Args[1], n.ZMin(key).Score)
	if err != nil {
		conn.WriteError("ERR min value is not a valid float")
		return
	}
	if cmd.Args[2][0] == '(' {
		mode |= zset.MaxOpen
	}
	max, err := redis.FormatFloat64(cmd.Args[2], n.ZMax(key).Score)
	if err != nil {
		conn.WriteError("ERR max value is not a valid float")
		return
	}
	var offset, count int64 = 0, -1
	if cmd.Options.LIMIT > 0 {
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
	if cmd.Options.WITHSCORES > 0 {
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
}

func zRevRangeByScore(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZREVRANGEBYSCORE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	var mode int
	if cmd.Args[2][0] == '(' {
		mode = zset.MaxOpen
	}
	min, err := redis.FormatFloat64(cmd.Args[2], n.ZMin(key).Score)
	if err != nil {
		conn.WriteError("ERR min value is not a valid float")
		return
	}
	if cmd.Args[1][0] == '(' {
		mode |= zset.MinOpen
	}
	max, err := redis.FormatFloat64(cmd.Args[1], n.ZMax(key).Score)
	if err != nil {
		conn.WriteError("ERR max value is not a valid float")
		return
	}
	var offset, count int64 = 0, -1
	if cmd.Options.LIMIT > 0 {
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
	if cmd.Options.WITHSCORES > 0 {
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
}

func zCount(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZCOUNT requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	mode := 0
	if cmd.Args[1][0] == '(' {
		mode = zset.MinOpen
	}
	min, err := redis.FormatFloat64(cmd.Args[1], n.ZMin(key).Score)
	if err != nil {
		conn.WriteError("ERR min value is not a valid float")
		return
	}
	if cmd.Args[2][0] == '(' {
		mode |= zset.MaxOpen
	}
	max, err := redis.FormatFloat64(cmd.Args[2], n.ZMax(key).Score)
	if err != nil {
		conn.WriteError("ERR max value is not a valid float")
		return
	}
	conn.WriteInteger(n.ZCount(key, min, max, mode))
}

func zRem(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ZREM requires at least two arguments")
	}
	key := cmd.Args[0]
	conn.WriteInteger(n.ZRem(key, cmd.Args[1:]...))
}

func zRemRangeByRank(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	conn.WriteInteger(n.ZRemRangeByRank(key, start, stop))
}

func zRemRangeByScore(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		conn.WriteError("ZREMRANGEBYSCORE requires at least three arguments")
	}
	key := cmd.Args[0]
	min, err := redis.FormatFloat64(cmd.Args[1], n.ZMin(key).Score)
	if err != nil {
		conn.WriteError("ERR min value is not a valid float")
		return
	}
	max, err := redis.FormatFloat64(cmd.Args[1], n.ZMax(key).Score)
	if err != nil {
		conn.WriteError("ERR max value is not a valid float")
		return
	}
	conn.WriteInteger(n.ZRemRangeByScore(key, min, max))
}

func zUnionStore(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	if cmd.Options.WEIGHTS > 0 {
		if len(cmd.Args) < cmd.Options.WEIGHTS+int(numKeys) {
			conn.WriteError("ERR syntax error")
			return
		}
		weights = make([]float64, numKeys)
		for i := 0; i < len(weights); i++ {
			weights[i], err = redis.FormatFloat64(cmd.Args[i+cmd.Options.WEIGHTS], 1)
			if err != nil {
				conn.WriteError("ERR weight value is not a valid float")
				return
			}
		}
	}
	if cmd.Options.AGGREGATE > 0 {
		aggregate = cmd.Args[cmd.Options.AGGREGATE]
	}
	n.ZUnionStore(destination, keys, weights, utils.ToUpper(aggregate))
	conn.WriteInteger(n.ZCard(destination))
}

func zInterStore(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
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
	if cmd.Options.WEIGHTS > 0 {
		if len(cmd.Args) < cmd.Options.WEIGHTS+int(numKeys) {
			conn.WriteError("ERR syntax error")
			return
		}
		weights = make([]float64, numKeys)
		for i := 0; i < len(weights); i++ {
			weights[i], err = redis.FormatFloat64(cmd.Args[i+cmd.Options.WEIGHTS], 1)
			if err != nil {
				conn.WriteError("ERR weight value is not a valid float")
				return
			}
		}
	}
	if cmd.Options.AGGREGATE > 0 {
		aggregate = cmd.Args[cmd.Options.AGGREGATE]
	}
	n.ZInterStore(destination, keys, weights, utils.ToUpper(aggregate))
	conn.WriteInteger(n.ZCard(destination))
}

func zClear(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		conn.WriteError("ZCLEAR requires at least one argument")
	}
	key := cmd.Args[0]
	n.ZClear(key)
	conn.WriteString("OK")
}

func zExists(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		conn.WriteError("ZEXISTS requires at least two arguments")
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	is := n.ZExists(key, member)
	var r int64 = 0
	if is {
		r = 1
	}
	conn.WriteInteger(r)
}

func save(n *Nodis, conn *redis.Conn, cmd *redis.Command) {
	n.store.save()
	conn.WriteString("OK")
}
