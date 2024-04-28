package nodis

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/diiyw/nodis/redis"
	"github.com/diiyw/nodis/utils"
)

func getCommand(name string) func(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	switch name {
	case "CLIENT":
		return client
	case "CONFIG":
		return config
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
	case "RENAME":
		return rename
	case "TYPE":
		return typ
	case "SCAN":
		return scan
	case "SET":
		return setString
	case "SETEX":
		return setex
	case "GET":
		return getString
	case "INCR":
		return incr
	case "INCRBY":
		return incrBy
	case "DESR":
		return decr
	case "DECRBY":
		return decrBy
	case "SETBIT":
		return setBit
	case "GETBIT":
		return getBit
	case "BITCOUNT":
		return bitCount
	case "SADD":
		return sAdd
	case "SSCAN":
		return sScan
	case "SCARD":
		return scard
	case "SPOP":
		return sPop
	case "SDIFF":
		return sDiff
	case "SINTER":
		return sInter
	case "SISMEMBER":
		return sIsMember
	case "SMEMBERS":
		return sMembers
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
	case "ZREMRANGEBYRANK":
		return zRemRangeByRank
	case "ZREMRANGEBYSCORE":
		return zRemRangeByScore
	case "ZCLEAR":
		return zClear
	case "ZEXISTS":
		return zExists
	}
	return nil
}

func client(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("CLIENT subcommand must be provided")
		return
	}
	switch utils.ToUpper(cmd.Args[0]) {
	case "LIST":
		w.WriteArray(0)
	case "SETNAME":
		w.WriteString("OK")
	default:
		w.WriteError("CLIENT subcommand must be provided")
	}
}

func config(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("CONFIG GET requires at least two argument")
		return
	}
	if cmd.Options.GET > 0 {
		if cmd.Options.GET > len(cmd.Args) {
			w.WriteError("CONFIG GET requires at least one argument")
			return
		}
		if utils.ToUpper(cmd.Args[cmd.Options.GET]) == "DATABASES" {
			w.WriteArray(2)
			w.WriteBulk("databases")
			w.WriteBulk("0")
			return
		}
	}
	w.WriteNull()
}

func info(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	usedMemory := strconv.FormatUint(memStats.HeapInuse+memStats.StackInuse, 10)
	pid := strconv.Itoa(os.Getpid())
	w.WriteBulk(`# Server
redis_version:nodis-1.5.0
os:` + runtime.GOOS + `
process_id:` + pid + `
# Memory
used_memory:` + usedMemory + `
`)
}

func ping(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteBulk("PONG")
		return
	}
	w.WriteBulk(cmd.Args[0])
}

func echo(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteNull()
		return
	}
	w.WriteBulk(cmd.Args[0])
}

func quit(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	w.WriteOK()
}

func flushDB(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	n.Clear()
	w.WriteOK()
}

func del(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("DEL requires at least one argument")
		return
	}
	w.WriteInteger(n.Del(cmd.Args...))
}

func exists(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("EXISTS requires at least one argument")
		return
	}
	w.WriteInteger(n.Exists(cmd.Args...))
}

// EXPIRE key seconds [NX | XX | GT | LT]
func expire(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("EXPIRE requires at least two arguments")
		return
	}
	seconds, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	if cmd.Options.NX > 0 {
		w.WriteInteger(n.ExpireNX(cmd.Args[0], seconds))
		return
	}
	if cmd.Options.XX > 0 {
		w.WriteInteger(n.ExpireXX(cmd.Args[0], seconds))
		return
	}
	if cmd.Options.LT > 0 {
		w.WriteInteger(n.ExpireLT(cmd.Args[0], seconds))
		return
	}
	if cmd.Options.GT > 0 {
		w.WriteInteger(n.ExpireGT(cmd.Args[0], seconds))
		return
	}
	w.WriteInteger(n.Expire(cmd.Args[0], seconds))
}

// EXPIREAT key unix-time-seconds [NX | XX | GT | LT]
func expireAt(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("EXPIREAT requires at least two arguments")
		return
	}
	timestamp, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	e := time.Unix(timestamp, 0)
	if cmd.Options.NX > 0 {
		w.WriteInteger(n.ExpireAtNX(cmd.Args[0], e))
		return
	}
	if cmd.Options.XX > 0 {
		w.WriteInteger(n.ExpireAtXX(cmd.Args[0], e))
		return
	}
	if cmd.Options.LT > 0 {
		w.WriteInteger(n.ExpireAtLT(cmd.Args[0], e))
		return
	}
	if cmd.Options.GT > 0 {
		w.WriteInteger(n.ExpireAtGT(cmd.Args[0], e))
		return
	}
	w.WriteInteger(n.ExpireAt(cmd.Args[0], e))
}

func keys(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("KEYS requires at least one argument")
		return
	}
	keys := n.Keys(cmd.Args[0])
	w.WriteArray(len(keys))
	for _, v := range keys {
		w.WriteBulk(v)
	}
}

func ttl(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("TTL requires at least one argument")
		return
	}
	w.WriteInteger(int64(n.TTL(cmd.Args[0]).Seconds()))
}

func randomKey(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	key := n.RandomKey()
	if key == "" {
		w.WriteNull()
		return
	}
	w.WriteBulk(key)
}

func rename(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("RENAME requires at least two arguments")
		return
	}
	oldKey := cmd.Args[0]
	newKey := cmd.Args[1]
	v := n.Rename(oldKey, newKey)
	if v == nil {
		w.WriteInteger(1)
		return
	}
	w.WriteError(v.Error())
}

func typ(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("TYPE requires at least one argument")
		return
	}
	key := cmd.Args[0]
	w.WriteBulk(n.Type(key))
}

// SCAN cursor [MATCH pattern] [COUNT count] [TYPE type]
func scan(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("SCAN requires at least one argument")
		return
	}
	cursor, err := strconv.ParseInt(cmd.Args[0], 10, 64)
	if err != nil {
		w.WriteError("ERR cursor value is not an integer or out of range")
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
			w.WriteError("ERR count value is not an integer or out of range")
			return
		}
	}
	nextCursor, keys := n.Scan(cursor, match, count)
	w.WriteArray(2)
	w.WriteBulk(strconv.FormatInt(nextCursor, 10))
	w.WriteArray(len(keys))
	for _, v := range keys {
		w.WriteBulk(v)
	}
}

// SET key value [NX | XX] [GET] [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL]
func setString(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("SET requires at least two arguments")
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
		w.WriteOK()
		return
	}
	if cmd.Options.PX > 0 {
		milliseconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.PX], 10, 64)
		n.ExpirePX(key, milliseconds)
		w.WriteOK()
		return
	}
	if cmd.Options.EXAT > 0 {
		seconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.EXAT], 10, 64)
		n.ExpireAt(key, time.Unix(seconds, 0))
		w.WriteOK()
		return
	}
	if cmd.Options.PXAT > 0 {
		milliseconds, _ := strconv.ParseInt(cmd.Args[cmd.Options.PXAT], 10, 64)
		seconds := milliseconds / 1000
		ns := (milliseconds - seconds*1000) * 1000 * 1000
		n.ExpireAt(key, time.Unix(seconds, ns))
		w.WriteOK()
		return
	}
	if get != nil {
		w.WriteBulk(string(get))
		return
	}
	w.WriteOK()
}

func setex(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("SETEX requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	seconds, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	value := []byte(cmd.Args[2])
	n.SetEX(key, value, seconds)
	w.WriteOK()
}

func incr(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("INCR requires at least one argument")
		return
	}
	key := cmd.Args[0]
	w.WriteInteger(n.Incr(key))
}

func incrBy(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("INCRBY requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	w.WriteInteger(n.IncrBy(key, value))
}

func decr(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("DECR requires at least one argument")
		return
	}
	key := cmd.Args[0]
	w.WriteInteger(n.Decr(key))
}

func decrBy(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("INCRBY requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	w.WriteInteger(n.DecrBy(key, value))
}

func getString(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("GET requires at least one argument")
		return
	}
	v := n.Get(cmd.Args[0])
	if v == nil {
		w.WriteNull()
		return
	}
	w.WriteBulk(string(v))
}

func setBit(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("SETBIT requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	offset, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil || offset < 0 {
		w.WriteError("ERR offset value is not an integer or out of range")
		return
	}
	value, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil || (value != 0 && value != 1) {
		w.WriteError("ERR bit value is not a valid integer")
		return
	}
	w.WriteInteger(n.SetBit(key, offset, value == 1))
}

func getBit(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("GETBIT requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	offset, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil || offset < 0 {
		w.WriteError("ERR offset value is not an integer or out of range")
		return
	}
	w.WriteInteger(n.GetBit(key, offset))
}

func bitCount(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("BITCOUNT requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	var start, end int64
	if len(cmd.Args) > 1 {
		start, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			w.WriteError("ERR start value is not an integer or out of range")
			return
		}
	}
	if len(cmd.Args) > 2 {
		end, err = strconv.ParseInt(cmd.Args[2], 10, 64)
		if err != nil {
			w.WriteError("ERR end value is not an integer or out of range")
			return
		}
	}
	w.WriteInteger(n.BitCount(key, start, end, cmd.Options.BIT > 0))
}

func sAdd(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("SADD requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	w.WriteInteger(n.SAdd(key, cmd.Args[1:]...))
}

// SSCAN key cursor [MATCH pattern] [COUNT count]
func sScan(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 1 {
		w.WriteError("SSCAN requires at least one argument")
		return
	}
	key := cmd.Args[0]
	cursor, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		w.WriteError("ERR value is not an integer or out of range")
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
	w.WriteArray(2)
	w.WriteBulk(strconv.FormatInt(cursor, 10))
	w.WriteArray(len(keys))
	for _, v := range keys {
		w.WriteBulk(v)
	}
}

func sPop(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 1 {
		w.WriteError("SPOP requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	var count int64 = 1
	if len(cmd.Args) > 1 {
		count, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			w.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	results := n.SPop(key, count)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(v)
	}
}

func scard(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("SCARD requires at least one argument")
		return
	}
	key := cmd.Args[0]
	w.WriteInteger(n.SCard(key))
}

func sDiff(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("SDIFF requires at least two arguments")
		return
	}
	results := n.SDiff(cmd.Args...)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(v)
	}
}

func sInter(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("SINTER requires at least two arguments")
		return
	}
	results := n.SInter(cmd.Args...)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(v)
	}
}

func sIsMember(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("SISMEMBER requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	var r int64 = 0
	is := n.SIsMember(key, member)
	if is {
		r = 1
	}
	w.WriteInteger(r)
}

func sMembers(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("SMEMBERS requires at least one argument")
		return
	}
	key := cmd.Args[0]
	results := n.SMembers(key)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(v)
	}
}

func sRem(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("SREM requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	w.WriteInteger(n.SRem(key, cmd.Args[1:]...))
}

func hSet(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("HSET requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	value := cmd.Args[2]
	var i int64 = 1
	n.HSet(key, field, []byte(value))
	if len(cmd.Args) > 3 {
		var fields = make(map[string][]byte, len(cmd.Args)-3)
		for i := 3; i < len(cmd.Args); i += 2 {
			if i+1 >= len(cmd.Args) {
				break
			}
			fields[cmd.Args[i]] = []byte(cmd.Args[i+1])
		}
		i++
		n.HMSet(key, fields)
	}
	w.WriteInteger(i)
}

func hGet(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("HGET requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	v := string(n.HGet(key, field))
	if v == "" {
		w.WriteNull()
		return
	}
	w.WriteBulk(v)
}

func hDel(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("HDEL requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	w.WriteInteger(n.HDel(key, cmd.Args[1:]...))
}

func hLen(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("HLEN requires at least one argument")
		return
	}
	key := cmd.Args[0]
	w.WriteInteger(n.HLen(key))
}

func hKeys(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("HKEYS requires at least one argument")
		return
	}
	key := cmd.Args[0]
	results := n.HKeys(key)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(v)
	}
}

func hExists(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("HEXISTS requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	is := n.HExists(key, field)
	var r int64 = 0
	if is {
		r = 1
	}
	w.WriteInteger(r)
}

func hGetAll(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("HGETALL requires at least one argument")
		return
	}
	key := cmd.Args[0]
	results := n.HGetAll(key)
	w.WriteArray(len(results) * 2)
	for k, v := range results {
		w.WriteBulk(string(k))
		w.WriteBulk(string(v))
	}
}

func hIncrBy(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("HINCRBY requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	value, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		w.WriteError("ERR value is not an integer or out of range")
		return
	}
	w.WriteInteger(n.HIncrBy(key, field, value))
}

func hIncrByFloat(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("HINCRBYFLOAT requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	value, err := strconv.ParseFloat(cmd.Args[2], 64)
	if err != nil {
		w.WriteError("ERR value is not a valid float")
		return
	}
	f := strconv.FormatFloat(n.HIncrByFloat(key, field, value), 'f', -1, 64)
	w.WriteBulk(f)
}

func hSetNX(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("HSETNX requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	field := cmd.Args[1]
	value := cmd.Args[2]
	if n.HSetNX(key, field, []byte(value)) {
		w.WriteInteger(1)
	}
	w.WriteInteger(0)
}

func hMGet(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("HMGET requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	fields := cmd.Args[1:]
	results := n.HMGet(key, fields...)
	w.WriteArray(len(results))
	for _, v := range results {
		if v == nil {
			w.WriteNull()
		} else {
			w.WriteBulk(string(v))
		}
	}
}

func hMSet(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("HMSET requires at least three arguments")
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
	w.WriteString("OK")
}

func hClear(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("HCLEAR requires at least one argument")
		return
	}
	key := cmd.Args[0]
	n.HClear(key)
	w.WriteString("OK")
}

func hScan(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("HSCAN requires at least one argument")
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
			w.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	_, results := n.HScan(key, cursor, match, count)
	w.WriteArray(2)
	w.WriteBulk(strconv.FormatInt(cursor, 10))
	w.WriteArray(len(results) * 2)
	for k, v := range results {
		w.WriteBulk(k)
		w.WriteBulk(string(v))
	}
}

func hVals(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("HVALS requires at least one argument")
		return
	}
	key := cmd.Args[0]
	results := n.HVals(key)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(string(v))
	}
}

func lPush(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("LPUSH requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	values := make([][]byte, 0, len(cmd.Args)-1)
	for i := 1; i < len(cmd.Args); i++ {
		values = append(values, []byte(cmd.Args[i]))
	}
	w.WriteInteger(n.LPush(key, values...))
}

func rPush(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("RPUSH requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	values := make([][]byte, 0, len(cmd.Args)-1)
	for i := 1; i < len(cmd.Args); i++ {
		values = append(values, []byte(cmd.Args[i]))
	}

	w.WriteInteger(n.RPush(key, values...))
}

func lPop(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("LPOP requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	var count int64 = 1
	if len(cmd.Args) > 1 {
		count, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			w.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	v := n.LPop(key, count)
	if v == nil {
		w.WriteNull()
		return
	}
	if count == 1 {
		w.WriteBulk(string(v[0]))
		return
	}
	w.WriteArray(len(v))
	for _, vv := range v {
		w.WriteBulk(string(vv))
	}
}

func rPop(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("RPOP requires at least one argument")
		return
	}
	var err error
	key := cmd.Args[0]
	var count int64 = 1
	if len(cmd.Args) > 1 {
		count, err = strconv.ParseInt(cmd.Args[1], 10, 64)
		if err != nil {
			w.WriteError("ERR value is not an integer or out of range")
			return
		}
	}
	v := n.RPop(key, count)
	if v == nil {
		w.WriteNull()
		return
	}
	if count == 1 {
		w.WriteBulk(string(v[0]))
		return
	}
	w.WriteArray(len(v))
	for _, vv := range v {
		w.WriteBulk(string(vv))
	}
}

func llen(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("LLEN requires at least one argument")
		return
	}
	key := cmd.Args[0]
	w.WriteInteger(n.LLen(key))
}

func lIndex(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("LINDEX requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	index, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	v := n.LIndex(key, index)
	if v == nil {
		w.WriteNull()
		return
	}
	w.WriteBulk(string(v))
}

func lInsert(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 4 {
		w.WriteError("LINSERT requires at least four arguments")
		return
	}
	key := cmd.Args[0]
	before := utils.ToUpper(cmd.Args[1]) == "BEFORE"
	pivot := []byte(cmd.Args[2])
	value := []byte(cmd.Args[3])
	w.WriteInteger(n.LInsert(key, pivot, value, before))
}

func lPushx(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("LPUSHX requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	value := []byte(cmd.Args[1])
	w.WriteInteger(n.LPushX(key, value))
}

func rPushx(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("RPUSHX requires at least two arguments")
	}
	key := cmd.Args[0]
	value := []byte(cmd.Args[1])
	w.WriteInteger(n.RPushX(key, value))
}

func lRem(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("LREM requires at least two arguments")
		return
	}
	var err error
	key := cmd.Args[0]
	count, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		w.WriteError("ERR value is not an integer or out of range")
		return
	}
	value := []byte(cmd.Args[2])
	w.WriteInteger(n.LRem(key, count, value))
}

func lSet(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("LSET requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	index, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		w.WriteError("ERR value is not an integer or out of range")
		return
	}
	value := []byte(cmd.Args[2])
	n.LSet(key, index, value)
	w.WriteString("OK")
}

func lRange(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("LRANGE requires at least two arguments")
		return
	}
	key := cmd.Args[0]
	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		w.WriteError("ERR value is not an integer or out of range")
		return
	}
	end, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		w.WriteError("ERR value is not an integer or out of range")
		return
	}
	results := n.LRange(key, start, end)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(string(v))
	}
}

func lPopRPush(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("LPOPRPUSH requires at least two arguments")
		return
	}
	source := cmd.Args[0]
	destination := cmd.Args[1]
	v := n.LPopRPush(source, destination)
	if v == nil {
		w.WriteNull()
		return
	}
	w.WriteBulk(string(v))
}

func rPopLPush(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("RPOPLPUSH requires at least two arguments")
		return
	}
	source := cmd.Args[0]
	destination := cmd.Args[1]
	v := n.RPopLPush(source, destination)
	if v == nil {
		w.WriteNull()
		return
	}
	w.WriteBulk(string(v))
}

func bLPop(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("BLPOP requires at least two arguments")
		return
	}
	keys := make([]string, 0, len(cmd.Args)-1)
	for i := 0; i < len(cmd.Args)-1; i++ {
		keys = append(keys, cmd.Args[i])
	}
	timeout, err := strconv.ParseInt(cmd.Args[len(cmd.Args)-1], 10, 64)
	if err != nil {
		w.WriteError("ERR timeout value is not an integer or out of range")
		return
	}
	k, v := n.BLPop(time.Duration(timeout)*time.Second, keys...)
	if k == "" {
		w.WriteNull()
		return
	}
	w.WriteArray(2)
	w.WriteBulk(k)
	w.WriteBulk(string(v))
}

func bRPop(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("BRPOP requires at least two arguments")
		return
	}
	keys := make([]string, 0, len(cmd.Args)-1)
	for i := 0; i < len(cmd.Args)-1; i++ {
		keys = append(keys, cmd.Args[i])
	}
	timeout, err := strconv.ParseInt(cmd.Args[len(cmd.Args)-1], 10, 64)
	if err != nil {
		w.WriteError("ERR timeout value is not an integer or out of range")
		return
	}
	k, v := n.BRPop(time.Duration(timeout)*time.Second, keys...)
	if k == "" {
		w.WriteNull()
		return
	}
	w.WriteArray(2)
	w.WriteBulk(k)
	w.WriteBulk(string(v))
}

// ZADD key [NX | XX] [GT | LT] [CH] [INCR] score member [score member   ...]
func zAdd(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("ZADD requires at least three arguments")
		return
	}
	itemStart := max(cmd.Options.NX, cmd.Options.XX, cmd.Options.LT, cmd.Options.GT, cmd.Options.CH, cmd.Options.INCR)
	if itemStart+1 > len(cmd.Args) {
		w.WriteError("ZADD requires at least one score-member pair")
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
			w.WriteError("ERR score value is not a valid float")
			return
		}
		member := cmd.Args[i+1]
		if cmd.Options.INCR > 0 {
			score = n.ZIncrBy(key, member, score)
			w.WriteBulk(strconv.FormatFloat(score, 'f', -1, 64))
		}
		if cmd.Options.XX > 0 {
			w.WriteInteger(n.ZAddXX(key, member, score))
		}
		if cmd.Options.NX > 0 {
			w.WriteInteger(n.ZAddNX(key, member, score))
		}
		if cmd.Options.LT > 0 {
			w.WriteInteger(n.ZAddLT(key, member, score))
		}
		if cmd.Options.GT > 0 {
			w.WriteInteger(n.ZAddGT(key, member, score))
		}
		n.ZAdd(key, member, score)
		count++
	}
	w.WriteInteger(count)
}

func zCard(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("ZCARD requires at least one argument")
		return
	}
	key := cmd.Args[0]
	w.WriteInteger(n.ZCard(key))
}

// ZRANK key member [WITHSCORE]
func zRank(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("ZRANK requires at least two argument")
		return
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	if cmd.Options.WITHSCORES > 0 {
		rank, el := n.ZRankWithScore(key, member)
		if el != nil {
			w.WriteArray(2)
			w.WriteInteger(rank)
			w.WriteBulk(el.Member)
			return
		}
		w.WriteNull()
		return
	}
	w.WriteInteger(n.ZRank(key, member))
}

func zRevRank(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("ZREVRANK requires at least two argument")
		return
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	if cmd.Options.WITHSCORES > 0 {
		rank, el := n.ZRevRankWithScore(key, member)
		if el != nil {
			w.WriteArray(2)
			w.WriteInteger(rank)
			w.WriteBulk(el.Member)
			return
		}
		w.WriteNull()
		return
	}
	w.WriteInteger(n.ZRevRank(key, member))
}

func zScore(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("ZSCORE requires at least two argument")
		return
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	score := n.ZScore(key, member)
	w.WriteBulk(strconv.FormatFloat(score, 'f', -1, 64))
}

func zIncrBy(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("ZINCRBY requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	score, err := strconv.ParseFloat(cmd.Args[1], 64)
	if err != nil {
		w.WriteError("ERR score value is not a valid float")
		return
	}
	member := cmd.Args[2]
	v := n.ZIncrBy(key, member, score)
	w.WriteBulk(strconv.FormatFloat(v, 'f', -1, 64))
}

func zRange(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("ZRANGE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	start, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
	stop, _ := strconv.ParseInt(cmd.Args[2], 10, 64)
	if cmd.Options.WITHSCORES > 0 {
		results := n.ZRangeWithScores(key, start, stop)
		w.WriteArray(len(results))
		for _, v := range results {
			w.WriteArray(2)
			w.WriteBulk(v.Member)
			w.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
		}
		return
	}
	results := n.ZRange(key, start, stop)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(v)
	}
}

func zRevRange(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("ZREVRANGE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		w.WriteError("ERR start value is not an integer or out of range")
		return
	}
	stop, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		w.WriteError("ERR stop value is not an integer or out of range")
		return
	}
	if cmd.Options.WITHSCORES > 0 {
		results := n.ZRevRangeWithScores(key, start, stop)
		w.WriteArray(len(results))
		for _, v := range results {
			w.WriteArray(2)
			w.WriteBulk(v.Member)
			w.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
		}
		return
	}
	results := n.ZRevRange(key, start, stop)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(v)
	}
}

// ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
func zRangeByScore(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("ZRANGEBYSCORE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	min, err := strconv.ParseFloat(cmd.Args[1], 64)
	if err != nil {
		w.WriteError("ERR min value is not a valid float")
		return
	}
	max, err := strconv.ParseFloat(cmd.Args[2], 64)
	if err != nil {
		w.WriteError("ERR max value is not a valid float")
		return
	}
	var offset, count int64 = 0, -1
	if cmd.Options.LIMIT > 0 {
		offset, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT], 10, 64)
		if err != nil {
			w.WriteError("ERR offset value is not an integer or out of range")
			return
		}
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT+1], 10, 64)
		if err != nil {
			w.WriteError("ERR count value is not an integer or out of range")
			return
		}
	}
	if cmd.Options.WITHSCORES > 0 {
		results := n.ZRangeByScoreWithScores(key, min, max, offset, count)
		w.WriteArray(len(results))
		for _, v := range results {
			w.WriteArray(2)
			w.WriteBulk(v.Member)
			w.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
		}
		return

	}
	results := n.ZRangeByScore(key, min, max, offset, count)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(v)
	}
}

func zRevRangeByScore(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("ZREVRANGEBYSCORE requires at least three arguments")
		return
	}
	key := cmd.Args[0]
	min, err := strconv.ParseFloat(cmd.Args[1], 64)
	if err != nil {
		w.WriteError("ERR min value is not a valid float")
		return
	}
	max, err := strconv.ParseFloat(cmd.Args[2], 64)
	if err != nil {
		w.WriteError("ERR max value is not a valid float")
		return
	}
	var offset, count int64 = 0, -1
	if cmd.Options.LIMIT > 0 {
		offset, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT], 10, 64)
		if err != nil {
			w.WriteError("ERR offset value is not an integer or out of range")
			return
		}
		count, err = strconv.ParseInt(cmd.Args[cmd.Options.LIMIT+1], 10, 64)
		if err != nil {
			w.WriteError("ERR count value is not an integer or out of range")
			return
		}
	}
	if cmd.Options.WITHSCORES > 0 {
		results := n.ZRevRangeByScoreWithScores(key, min, max, offset, count)
		w.WriteArray(len(results))
		for _, v := range results {
			w.WriteArray(2)
			w.WriteBulk(v.Member)
			w.WriteBulk(strconv.FormatFloat(v.Score, 'f', -1, 64))
		}
		return
	}
	results := n.ZRevRangeByScore(key, min, max, offset, count)
	w.WriteArray(len(results))
	for _, v := range results {
		w.WriteBulk(v)
	}
}

func zRem(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("ZREM requires at least two arguments")
	}
	key := cmd.Args[0]
	w.WriteInteger(n.ZRem(key, cmd.Args[1:]...))
}

func zRemRangeByRank(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("ZREMRANGEBYRANK requires at least three arguments")
	}
	key := cmd.Args[0]
	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		w.WriteError("ERR start value is not a valid float")
		return
	}
	stop, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		w.WriteError("ERR stop value is not a valid float")
		return
	}
	w.WriteInteger(n.ZRemRangeByRank(key, start, stop))
}

func zRemRangeByScore(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 3 {
		w.WriteError("ZREMRANGEBYSCORE requires at least three arguments")
	}
	key := cmd.Args[0]
	min, err := strconv.ParseFloat(cmd.Args[1], 64)
	if err != nil {
		w.WriteError("ERR min value is not a valid float")
		return
	}
	max, err := strconv.ParseFloat(cmd.Args[1], 64)
	if err != nil {
		w.WriteError("ERR max value is not a valid float")
		return
	}
	w.WriteInteger(n.ZRemRangeByScore(key, min, max))
}

func zClear(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) == 0 {
		w.WriteError("ZCLEAR requires at least one argument")
	}
	key := cmd.Args[0]
	n.ZClear(key)
	w.WriteString("OK")
}

func zExists(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	if len(cmd.Args) < 2 {
		w.WriteError("ZEXISTS requires at least two arguments")
	}
	key := cmd.Args[0]
	member := cmd.Args[1]
	is := n.ZExists(key, member)
	var r int64 = 0
	if is {
		r = 1
	}
	w.WriteInteger(r)
}

func save(n *Nodis, w *redis.Writer, cmd *redis.Command) {
	n.store.mu.Lock()
	n.store.save()
	n.store.mu.Unlock()
	w.WriteString("OK")
}
