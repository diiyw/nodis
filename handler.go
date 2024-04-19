package nodis

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/btree"

	"github.com/diiyw/nodis/redis"
)

var redisCommands btree.Map[string, func(*Nodis, redis.Value, []redis.Value) redis.Value]

func init() {
	redisCommands.Set("CLIENT", client)
	redisCommands.Set("CONFIG", config)
	redisCommands.Set("PING", ping)
	redisCommands.Set("QUIT", quit)
	redisCommands.Set("FLUSHDB", flushdb)
	redisCommands.Set("INFO", info)
	redisCommands.Set("DEL", del)
	redisCommands.Set("EXISTS", exists)
	redisCommands.Set("EXPIRE", expire)
	redisCommands.Set("EXPIREAT", expireAt)
	redisCommands.Set("KEYS", keys)
	redisCommands.Set("TTL", ttl)
	redisCommands.Set("RENAME", rename)
	redisCommands.Set("TYPE", typ)
	redisCommands.Set("SCAN", scan)
	redisCommands.Set("SET", setString)
	redisCommands.Set("GET", getString)
	redisCommands.Set("INCR", incr)
	redisCommands.Set("DESR", decr)
	redisCommands.Set("SETBIT", setBit)
	redisCommands.Set("GETBIT", getBit)
	redisCommands.Set("BITCOUNT", bitCount)
	redisCommands.Set("SADD", sAdd)
	redisCommands.Set("SSCAN", sScan)
	redisCommands.Set("SCARD", scard)
	redisCommands.Set("SPOP", sPop)
	redisCommands.Set("SDIFF", sDiff)
	redisCommands.Set("SINTER", sInter)
	redisCommands.Set("SISMEMBER", sIsMember)
	redisCommands.Set("SMEMBERS", sMembers)
	redisCommands.Set("SREM", sRem)
	redisCommands.Set("HSET", hSet)
	redisCommands.Set("HGET", hGet)
	redisCommands.Set("HDEL", hDel)
	redisCommands.Set("HLEN", hLen)
	redisCommands.Set("HKEYS", hKeys)
	redisCommands.Set("HEXISTS", hExists)
	redisCommands.Set("HGETALL", hGetAll)
	redisCommands.Set("HINCRBY", hIncrBy)
	redisCommands.Set("HINCRBYFLOAT", hIncrByFloat)
	redisCommands.Set("HSETNX", hSetNX)
	redisCommands.Set("HMGET", hMGet)
	redisCommands.Set("HMSET", hMSet)
	redisCommands.Set("HCLEAR", hClear)
	redisCommands.Set("HSCAN", hScan)
	redisCommands.Set("HVALS", hVals)
	redisCommands.Set("LPUSH", lPush)
	redisCommands.Set("RPUSH", rPush)
	redisCommands.Set("LPOP", lPop)
	redisCommands.Set("RPOP", rPop)
	redisCommands.Set("LLEN", llen)
	redisCommands.Set("LINDEX", lIndex)
	redisCommands.Set("LINSERT", lInsert)
	redisCommands.Set("LPUSHX", lPushx)
	redisCommands.Set("RPUSHX", rPushx)
	redisCommands.Set("LREM", lRem)
	redisCommands.Set("LSET", lSet)
	redisCommands.Set("LRANGE", lRange)
	redisCommands.Set("LPOPRPUSH", lPopRPush)
	redisCommands.Set("RPOPLPUSH", rPopLPush)
	redisCommands.Set("ZADD", zAdd)
	redisCommands.Set("ZCARD", zCard)
	redisCommands.Set("ZRANK", zRank)
	redisCommands.Set("ZREVRANK", zRevRank)
	redisCommands.Set("ZSCORE", zScore)
	redisCommands.Set("ZINCRBY", zIncrBy)
	redisCommands.Set("ZRANGE", zRange)
	redisCommands.Set("ZREVRANGE", zRevRange)
	redisCommands.Set("ZRANGEBYSCORE", zRangeByScore)
	redisCommands.Set("ZREVRANGEBYSCORE", zRevRangeByScore)
	redisCommands.Set("ZREM", zRem)
	redisCommands.Set("ZREMRANGEBYRANK", zRemRangeByRank)
	redisCommands.Set("ZREMRANGEBYSCORE", zRemRangeByScore)
	redisCommands.Set("ZCLEAR", zClear)
	redisCommands.Set("ZEXISTS", zExists)
}

func client(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("CLIENT subcommand must be provided")
	}
	switch strings.ToUpper(args[0].Bulk) {
	case "LIST":
		return redis.ArrayValue()
	default:
		return redis.ErrorValue("CLIENT subcommand must be provided")
	}
}

func config(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("CONFIG subcommand must be provided")
	}
	switch strings.ToUpper(args[0].Bulk) {
	case "GET":
		if len(args) == 1 {
			return redis.ErrorValue("CONFIG GET requires at least one argument")
		}
		if args[1].Bulk == "databases" {
			return redis.ArrayValue(redis.StringValue("databases"), redis.IntegerValue(0))
		}
		return redis.BulkValue("CONFIG GET")
	case "SET":
		if len(args) == 1 {
			return redis.ErrorValue("CONFIG SET requires at least one argument")
		}
		return redis.BulkValue("CONFIG SET")
	default:
		return redis.ErrorValue("CONFIG subcommand must be provided")
	}
}

func info(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	usedMemory := strconv.FormatUint(memStats.Sys, 10)
	pid := strconv.Itoa(os.Getpid())
	return redis.BulkValue(`# Server
redis_version:6.0.0
nodis_version:1.3.0
os:` + runtime.GOOS + `
process_id:` + pid + `
# Memory
used_memory:` + usedMemory)
}

func flushdb(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	n.Clear()
	return redis.StringValue("OK")

}

func quit(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	return redis.StringValue("OK")
}

func ping(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.StringValue("PONG")
	}
	return redis.BulkValue(args[0].Bulk)
}

func del(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("DEL requires at least one argument")
	}
	keys := make([]string, 0, len(args))
	for _, arg := range args {
		keys = append(keys, arg.Bulk)
	}
	return redis.IntegerValue(n.Del(keys...))
}

func exists(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("EXISTS requires at least one argument")
	}
	keys := make([]string, 0, len(args))
	for _, arg := range args {
		keys = append(keys, arg.Bulk)
	}
	return redis.IntegerValue(n.Exists(keys...))
}

func expire(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("EXPIRE requires at least two arguments")
	}
	key := args[0].Bulk
	seconds, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	if _, ok := cmd.Args["NX"]; ok {
		return redis.IntegerValue(n.ExpireNX(key, seconds))
	}
	if _, ok := cmd.Args["XX"]; ok {
		return redis.IntegerValue(n.ExpireXX(key, seconds))
	}
	if _, ok := cmd.Args["LT"]; ok {
		return redis.IntegerValue(n.ExpireLT(key, seconds))
	}
	if _, ok := cmd.Args["GT"]; ok {
		return redis.IntegerValue(n.ExpireGT(key, seconds))
	}
	return redis.IntegerValue(n.Expire(key, seconds))
}

func expireAt(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("EXPIREAT requires at least two arguments")
	}
	key := args[0].Bulk
	timestamp, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	e := time.Unix(timestamp, 0)
	if _, ok := cmd.Args["NX"]; ok {
		return redis.IntegerValue(n.ExpireAtNX(key, e))
	}
	if _, ok := cmd.Args["XX"]; ok {
		return redis.IntegerValue(n.ExpireAtXX(key, e))
	}
	if _, ok := cmd.Args["LT"]; ok {
		return redis.IntegerValue(n.ExpireAtLT(key, e))
	}
	if _, ok := cmd.Args["GT"]; ok {
		return redis.IntegerValue(n.ExpireAtGT(key, e))
	}
	return redis.IntegerValue(n.ExpireAt(key, e))
}

func keys(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("KEYS requires at least one argument")
	}
	pattern := args[0].Bulk
	keys := n.Keys(pattern)
	var k = make([]redis.Value, 0, len(keys))
	for _, v := range keys {
		k = append(k, redis.BulkValue(v))
	}
	return redis.ArrayValue(k...)
}

func ttl(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("TTL requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(int64(n.TTL(key).Seconds()))
}

func rename(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("RENAME requires at least two arguments")
	}
	oldKey := args[0].Bulk
	newKey := args[1].Bulk
	v := n.Rename(oldKey, newKey)
	if v == nil {
		return redis.IntegerValue(1)
	}
	return redis.ErrorValue(v.Error())
}

func typ(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("TYPE requires at least one argument")
	}
	key := args[0].Bulk
	return redis.BulkValue(n.Type(key))
}

func scan(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("SCAN requires at least one argument")
	}
	cursor, _ := strconv.ParseInt(args[0].Bulk, 10, 64)
	var match = "*"
	var count int64
	if v, ok := cmd.Args["MATCH"]; ok {
		match = v.Bulk
	}
	if v, ok := cmd.Args["COUNT"]; ok {
		count, _ = strconv.ParseInt(v.Bulk, 10, 64)
	}
	_, keys := n.Scan(cursor, match, count)
	var r = make([]redis.Value, 2)
	r[0] = redis.BulkValue(strconv.FormatInt(cursor, 10))
	var k = make([]redis.Value, 0, len(keys))
	for _, v := range keys {
		k = append(k, redis.BulkValue(v))
	}
	r[1] = redis.ArrayValue(k...)
	return redis.ArrayValue(r...)
}

func setString(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("SET requires at least two arguments")
	}
	key := args[0].Bulk
	value := []byte(args[1].Bulk)
	var get []byte
	if _, ok := cmd.Args["GET"]; ok {
		get = n.Get(key)
	}
	if _, ok := cmd.Args["NX"]; ok {
		n.SetNX(key, value)
	} else if _, ok = cmd.Args["XX"]; ok {
		n.SetXX(key, value)
	} else {
		n.Set(key, value)
	}
	if _, ok := cmd.Args["EX"]; ok {
		seconds, _ := strconv.ParseInt(cmd.Args["EX"].Bulk, 10, 64)
		n.Expire(key, seconds)
		return redis.StringValue("OK")
	}
	if _, ok := cmd.Args["PX"]; ok {
		milliseconds, _ := strconv.ParseInt(cmd.Args["PX"].Bulk, 10, 64)
		n.ExpirePX(key, milliseconds)
		return redis.StringValue("OK")
	}
	if _, ok := cmd.Args["EXAT"]; ok {
		seconds, _ := strconv.ParseInt(cmd.Args["EX"].Bulk, 10, 64)
		n.ExpireAt(key, time.Unix(seconds, 0))
		return redis.StringValue("OK")
	}
	if _, ok := cmd.Args["PXAT"]; ok {
		milliseconds, _ := strconv.ParseInt(cmd.Args["PX"].Bulk, 10, 64)
		seconds := milliseconds / 1000
		ns := (milliseconds - seconds*1000) * 1000 * 1000
		n.ExpireAt(key, time.Unix(seconds, ns))
		return redis.StringValue("OK")
	}
	if get != nil {
		return redis.BulkValue(string(get))
	}
	return redis.StringValue("OK")
}

func incr(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("INCR requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.Incr(key))
}

func decr(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("DECR requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.Incr(key))
}

func getString(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("GET requires at least one argument")
	}
	key := args[0].Bulk
	v := n.Get(key)
	if v == nil {
		return redis.NullValue()
	}
	return redis.BulkValue(string(v))
}

func setBit(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("SETBIT requires at least two arguments")
	}
	key := args[0].Bulk
	offset, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	value, _ := strconv.ParseBool(args[2].Bulk)
	n.SetBit(key, offset, value)
	return redis.IntegerValue(1)
}

func getBit(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("GETBIT requires at least two arguments")
	}
	key := args[0].Bulk
	offset, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	return redis.IntegerValue(n.GetBit(key, offset))
}

func bitCount(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("BITCOUNT requires at least one argument")
	}
	key := args[0].Bulk
	var start, end int64
	if len(args) > 1 {
		start, _ = strconv.ParseInt(args[1].Bulk, 10, 64)
	}
	if len(args) > 2 {
		end, _ = strconv.ParseInt(args[2].Bulk, 10, 64)
	}
	return redis.IntegerValue(n.BitCount(key, start, end))
}

func sAdd(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("SADD requires at least two arguments")
	}
	key := args[0].Bulk
	members := make([]string, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		members = append(members, args[i].Bulk)
	}
	return redis.IntegerValue(n.SAdd(key, members...))
}

func sScan(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 1 {
		return redis.ErrorValue("SSCAN requires at least one argument")
	}
	key := args[0].Bulk
	cursor, _ := strconv.ParseInt(args[0].Bulk, 10, 64)
	var match = "*"
	var count int64
	if v, ok := cmd.Args["MATCH"]; ok {
		match = v.Bulk
	}
	if v, ok := cmd.Args["COUNT"]; ok {
		count, _ = strconv.ParseInt(v.Bulk, 10, 64)
	}
	cursor, keys := n.SScan(key, cursor, match, count)
	var r = make([]redis.Value, 2)
	r[0] = redis.BulkValue(strconv.FormatInt(cursor, 10))
	var k = make([]redis.Value, 0, len(keys))
	for _, v := range keys {
		k = append(k, redis.BulkValue(v))
	}
	r[1] = redis.ArrayValue(k...)
	return redis.ArrayValue(r...)
}

func sPop(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("SPOP requires at least one argument")
	}
	key := args[0].Bulk
	var count int64 = 1
	if len(args) > 1 {
		count, _ = strconv.ParseInt(args[1].Bulk, 10, 64)
	}
	results := n.SPop(key, count)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(v))
	}
	return redis.ArrayValue(r...)
}

func scard(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("SCARD requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.SCard(key))
}

func sDiff(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("SDIFF requires at least two arguments")
	}
	keys := make([]string, 0, len(args))
	for _, arg := range args {
		keys = append(keys, arg.Bulk)
	}
	results := n.SDiff(keys...)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(v))
	}
	return redis.ArrayValue(r...)
}

func sInter(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("SINTER requires at least two arguments")
	}
	keys := make([]string, 0, len(args))
	for _, arg := range args {
		keys = append(keys, arg.Bulk)
	}
	results := n.SInter(keys...)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(v))
	}
	return redis.ArrayValue(r...)
}

func sIsMember(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("SISMEMBER requires at least two arguments")
	}
	key := args[0].Bulk
	member := args[1].Bulk
	var r int64 = 0
	is := n.SIsMember(key, member)
	if is {
		r = 1
	}
	return redis.IntegerValue(r)
}

func sMembers(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("SMEMBERS requires at least one argument")
	}
	key := args[0].Bulk
	results := n.SMembers(key)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(v))
	}
	return redis.ArrayValue(r...)
}

func sRem(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("SREM requires at least two arguments")
	}
	key := args[0].Bulk
	members := make([]string, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		members = append(members, args[i].Bulk)
	}
	return redis.IntegerValue(n.SRem(key, members...))
}

func hSet(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("HSET requires at least three arguments")
	}
	key := args[0].Bulk
	field := args[1].Bulk
	value := args[2].Bulk
	var i int64 = 1
	n.HSet(key, field, []byte(value))
	if len(args) > 3 {
		var fields = make(map[string][]byte, len(args)-3)
		for i := 3; i < len(args); i += 2 {
			if i+1 >= len(args) {
				break
			}
			fields[args[i].Bulk] = []byte(args[i+1].Bulk)
		}
		i++
		n.HMSet(key, fields)
	}
	return redis.IntegerValue(i)
}

func hGet(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("HGET requires at least two arguments")
	}
	key := args[0].Bulk
	field := args[1].Bulk
	return redis.BulkValue(string(n.HGet(key, field)))
}

func hDel(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("HDEL requires at least two arguments")
	}
	key := args[0].Bulk
	fields := make([]string, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		fields = append(fields, args[i].Bulk)
	}
	return redis.IntegerValue(n.HDel(key, fields...))
}

func hLen(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HLEN requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.HLen(key))
}

func hKeys(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HKEYS requires at least one argument")
	}
	key := args[0].Bulk
	results := n.HKeys(key)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(v))
	}
	return redis.ArrayValue(r...)
}

func hExists(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("HEXISTS requires at least two arguments")
	}
	key := args[0].Bulk
	field := args[1].Bulk
	is := n.HExists(key, field)
	var r int64 = 0
	if is {
		r = 1
	}
	return redis.IntegerValue(r)
}

func hGetAll(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HGETALL requires at least one argument")
	}
	key := args[0].Bulk
	results := n.HGetAll(key)
	var r = make(map[string]redis.Value)
	for k, v := range results {
		r[k] = redis.BulkValue(string(v))
	}
	return redis.MapValue(r)
}

func hIncrBy(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("HINCRBY requires at least three arguments")
	}
	key := args[0].Bulk
	field := args[1].Bulk
	value, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	return redis.IntegerValue(n.HIncrBy(key, field, value))
}

func hIncrByFloat(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("HINCRBYFLOAT requires at least three arguments")
	}
	key := args[0].Bulk
	field := args[1].Bulk
	value, _ := strconv.ParseFloat(args[2].Bulk, 64)
	f := strconv.FormatFloat(n.HIncrByFloat(key, field, value), 'f', -1, 64)
	return redis.BulkValue(f)
}

func hSetNX(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("HSETNX requires at least three arguments")
	}
	key := args[0].Bulk
	field := args[1].Bulk
	value := args[2].Bulk
	if n.HSetNX(key, field, []byte(value)) {
		return redis.IntegerValue(1)
	}
	return redis.IntegerValue(0)
}

func hMGet(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("HMGET requires at least two arguments")
	}
	key := args[0].Bulk
	fields := make([]string, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		fields = append(fields, args[i].Bulk)
	}
	results := n.HMGet(key, fields...)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(string(v)))
	}
	return redis.ArrayValue(r...)
}

func hMSet(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("HMSET requires at least three arguments")
	}
	key := args[0].Bulk
	fields := make(map[string][]byte, len(args)-1)
	for i := 1; i < len(args); i += 2 {
		fields[args[i].Bulk] = []byte(args[i+1].Bulk)
	}
	n.HMSet(key, fields)
	return redis.StringValue("OK")
}

func hClear(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HCLEAR requires at least one argument")
	}
	key := args[0].Bulk
	n.HClear(key)
	return redis.StringValue("OK")
}

func hScan(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HSCAN requires at least one argument")
	}
	key := args[0].Bulk
	cursor, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	var match = "*"
	var count int64
	if v, ok := cmd.Args["MATCH"]; ok {
		match = v.Bulk
	}
	if v, ok := cmd.Args["COUNT"]; ok {
		count, _ = strconv.ParseInt(v.Bulk, 10, 64)
	}
	_, results := n.HScan(key, cursor, match, count)
	var r = make([]redis.Value, 2)
	r[0] = redis.BulkValue(strconv.FormatInt(cursor, 10))
	var ret = make([]redis.Value, 0, len(results))
	for k, v := range results {
		ret = append(ret, redis.StringValue(k), redis.BulkValue(string(v)))
	}
	r[1] = redis.ArrayValue(ret...)
	return redis.ArrayValue(r...)
}

func hVals(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HVALS requires at least one argument")
	}
	key := args[0].Bulk
	results := n.HVals(key)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(string(v)))
	}
	return redis.ArrayValue(r...)
}

func lPush(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("LPUSH requires at least two arguments")
	}
	key := args[0].Bulk
	values := make([][]byte, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		values = append(values, []byte(args[i].Bulk))
	}
	return redis.IntegerValue(n.LPush(key, values...))
}

func rPush(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("RPUSH requires at least two arguments")
	}
	key := args[0].Bulk
	values := make([][]byte, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		values = append(values, []byte(args[i].Bulk))
	}

	return redis.IntegerValue(n.RPush(key, values...))
}

func lPop(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("LPOP requires at least one argument")
	}
	key := args[0].Bulk
	var count int64 = 1
	if len(args) > 1 {
		count, _ = strconv.ParseInt(args[1].Bulk, 10, 64)
	}
	v := n.LPop(key, count)
	if v == nil {
		return redis.NullValue()
	}
	if count == 1 {
		return redis.BulkValue(string(v[0]))
	}
	var r = make([]redis.Value, 0, len(v))
	for _, vv := range v {
		r = append(r, redis.BulkValue(string(vv)))
	}
	return redis.ArrayValue(r...)
}

func rPop(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("RPOP requires at least one argument")
	}
	key := args[0].Bulk
	var count int64 = 1
	if len(args) > 1 {
		count, _ = strconv.ParseInt(args[1].Bulk, 10, 64)
	}
	v := n.RPop(key, count)
	if v == nil {
		return redis.NullValue()
	}
	if count == 1 {
		return redis.BulkValue(string(v[0]))
	}
	var r = make([]redis.Value, 0, len(v))
	for _, vv := range v {
		r = append(r, redis.BulkValue(string(vv)))
	}
	return redis.ArrayValue(r...)
}

func llen(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("LLEN requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.LLen(key))
}

func lIndex(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("LINDEX requires at least two arguments")
	}
	key := args[0].Bulk
	index, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	v := n.LIndex(key, index)
	if v == nil {
		return redis.NullValue()
	}
	return redis.BulkValue(string(v))
}

func lInsert(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 4 {
		return redis.ErrorValue("LINSERT requires at least four arguments")
	}
	key := args[0].Bulk
	before := strings.ToUpper(args[1].Bulk) == "BEFORE"
	pivot := []byte(args[2].Bulk)
	value := []byte(args[3].Bulk)
	return redis.IntegerValue(n.LInsert(key, pivot, value, before))
}

func lPushx(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("LPUSHX requires at least two arguments")
	}
	key := args[0].Bulk
	value := []byte(args[1].Bulk)
	return redis.IntegerValue(n.LPushX(key, value))
}

func rPushx(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("RPUSHX requires at least two arguments")
	}
	key := args[0].Bulk
	value := []byte(args[1].Bulk)
	return redis.IntegerValue(n.RPushX(key, value))
}

func lRem(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("LREM requires at least two arguments")
	}
	key := args[0].Bulk
	count, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	value := []byte(args[2].Bulk)
	return redis.IntegerValue(n.LRem(key, count, value))
}

func lSet(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("LSET requires at least three arguments")
	}
	key := args[0].Bulk
	index, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	value := []byte(args[2].Bulk)
	n.LSet(key, index, value)
	return redis.StringValue("OK")
}

func lRange(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("LRANGE requires at least two arguments")
	}
	key := args[0].Bulk
	start, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	end, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	results := n.LRange(key, start, end)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(string(v)))
	}
	return redis.ArrayValue(r...)
}

func lPopRPush(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("LPOPRPUSH requires at least two arguments")
	}
	source := args[0].Bulk
	destination := args[1].Bulk
	v := n.LPopRPush(source, destination)
	if v == nil {
		return redis.NullValue()
	}
	return redis.BulkValue(string(v))
}

func rPopLPush(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("RPOPLPUSH requires at least two arguments")
	}
	source := args[0].Bulk
	destination := args[1].Bulk
	v := n.RPopLPush(source, destination)
	if v == nil {
		return redis.NullValue()
	}
	return redis.BulkValue(string(v))
}

func zAdd(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZADD requires at least three arguments")
	}
	key := args[0].Bulk
	score, _ := strconv.ParseFloat(args[1].Bulk, 64)
	member := args[2].Bulk
	if cmd.Options["INCR"] {
		score = n.ZIncrBy(key, member, score)
		return redis.DoubleValue(score)
	}
	if cmd.Options["XX"] {
		return redis.IntegerValue(n.ZAddXX(key, member, score))
	}
	if cmd.Options["NX"] {
		return redis.IntegerValue(n.ZAddNX(key, member, score))
	}
	if cmd.Options["LT"] {
		return redis.IntegerValue(n.ZAddLT(key, member, score))
	}
	if cmd.Options["GT"] {
		return redis.IntegerValue(n.ZAddGT(key, member, score))
	}
	n.ZAdd(key, member, score)
	return redis.IntegerValue(1)
}

func zCard(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("ZCARD requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.ZCard(key))
}

func zRank(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("ZRANK requires at least two argument")
	}
	key := args[0].Bulk
	member := args[1].Bulk
	if len(args) > 2 {
		rank, el := n.ZRankWithScore(key, member)
		if el != nil {
			return redis.ArrayValue(redis.IntegerValue(rank), redis.BulkValue(el.Member))
		}
		return redis.NullValue()
	}
	return redis.IntegerValue(n.ZRank(key, member))
}

func zRevRank(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("ZREVRANK requires at least two argument")
	}
	key := args[0].Bulk
	member := args[1].Bulk
	if len(args) > 2 {
		rank, el := n.ZRevRankWithScore(key, member)
		if el != nil {
			return redis.ArrayValue(redis.IntegerValue(rank), redis.BulkValue(el.Member))
		}
		return redis.NullValue()
	}
	return redis.IntegerValue(n.ZRevRank(key, member))
}

func zScore(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("ZSCORE requires at least two argument")
	}
	key := args[0].Bulk
	member := args[1].Bulk
	return redis.DoubleValue(n.ZScore(key, member))
}

func zIncrBy(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZINCRBY requires at least three arguments")
	}
	key := args[0].Bulk
	score, _ := strconv.ParseFloat(args[1].Bulk, 64)
	member := args[2].Bulk
	v := n.ZIncrBy(key, member, score)
	return redis.DoubleValue(v)
}

func zRange(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZRANGE requires at least three arguments")
	}
	key := args[0].Bulk
	start, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	stop, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	if cmd.Options["WITHSCORES"] {
		results := n.ZRangeWithScores(key, start, stop)
		var r = make([]redis.Value, 0, len(results)*2)
		for _, v := range results {
			r = append(r, redis.StringValue(v.Member), redis.BulkValue(strconv.FormatFloat(v.Score, 'f', -1, 64)))
		}
		return redis.ArrayValue(r...)
	}
	results := n.ZRange(key, start, stop)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(v))
	}
	return redis.ArrayValue(r...)
}

func zRevRange(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZREVRANGE requires at least three arguments")
	}
	key := args[0].Bulk
	start, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	stop, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	if cmd.Options["WITHSCORES"] {
		results := n.ZRevRangeWithScores(key, start, stop)
		var r = make([]redis.Value, 0, len(results)*2)
		for _, v := range results {
			r = append(r, redis.StringValue(v.Member), redis.BulkValue(strconv.FormatFloat(v.Score, 'f', -1, 64)))
		}
		return redis.ArrayValue(r...)
	}
	results := n.ZRevRange(key, start, stop)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(v))
	}
	return redis.ArrayValue(r...)
}

func zRangeByScore(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZRANGEBYSCORE requires at least three arguments")
	}
	key := args[0].Bulk
	min, _ := strconv.ParseFloat(args[1].Bulk, 64)
	max, _ := strconv.ParseFloat(args[1].Bulk, 64)
	results := n.ZRangeByScore(key, min, max)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(v))
	}
	return redis.ArrayValue(r...)
}

func zRevRangeByScore(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZREVRANGEBYSCORE requires at least three arguments")
	}
	key := args[0].Bulk
	min, _ := strconv.ParseFloat(args[1].Bulk, 64)
	max, _ := strconv.ParseFloat(args[1].Bulk, 64)
	results := n.ZRevRangeByScore(key, min, max)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.BulkValue(v))
	}
	return redis.ArrayValue(r...)
}

func zRem(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("ZREM requires at least two arguments")
	}
	key := args[0].Bulk
	members := make([]string, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		members = append(members, args[i].Bulk)
	}
	return redis.IntegerValue(n.ZRem(key, members...))
}

func zRemRangeByRank(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZREMRANGEBYRANK requires at least three arguments")
	}
	key := args[0].Bulk
	start, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	stop, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	return redis.IntegerValue(n.ZRemRangeByRank(key, start, stop))
}

func zRemRangeByScore(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZREMRANGEBYSCORE requires at least three arguments")
	}
	key := args[0].Bulk
	min, _ := strconv.ParseFloat(args[1].Bulk, 64)
	max, _ := strconv.ParseFloat(args[1].Bulk, 64)
	return redis.IntegerValue(n.ZRemRangeByScore(key, min, max))
}

func zClear(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("ZCLEAR requires at least one argument")
	}
	key := args[0].Bulk
	n.ZClear(key)
	return redis.StringValue("OK")
}

func zExists(n *Nodis, cmd redis.Value, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("ZEXISTS requires at least two arguments")
	}
	key := args[0].Bulk
	member := args[1].Bulk
	is := n.ZExists(key, member)
	var r int64 = 0
	if is {
		r = 1
	}
	return redis.IntegerValue(r)
}
