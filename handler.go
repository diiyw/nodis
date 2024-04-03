package nodis

import (
	"strconv"
	"strings"
	"time"

	"github.com/diiyw/nodis/redis"
)

var redisHandlers = map[string]func(*Nodis, []redis.Value) redis.Value{
	"PING":             ping,
	"DEL":              del,
	"EXISTS":           exists,
	"EXPIRE":           expire,
	"EXPIREAT":         expireAt,
	"KEYS":             keys,
	"TTL":              ttl,
	"RENAME":           rename,
	"TYPE":             typ,
	"SCAN":             scan,
	"SET":              setString,
	"GET":              getString,
	"SETBIT":           setBit,
	"GETBIT":           getBit,
	"BITCOUNT":         bitCount,
	"SADD":             sAdd,
	"SCARD":            scard,
	"SDIFF":            sDiff,
	"SINTER":           sInter,
	"SISMEMBER":        sIsMember,
	"SMEMBERS":         sMembers,
	"SREM":             sRem,
	"HSET":             hSet,
	"HGET":             hGet,
	"HDEL":             hDel,
	"HLEN":             hLen,
	"HKEYS":            hKeys,
	"HEXISTS":          hExists,
	"HGETALL":          hGetAll,
	"HINCRBY":          hIncrBy,
	"HINCRBYFLOAT":     hIncrByFloat,
	"HSETNX":           hSetNX,
	"HMGET":            hMGet,
	"HMSET":            hMSet,
	"HCLEAR":           hClear,
	"HSCAN":            hScan,
	"HVALS":            hVals,
	"LPUSH":            lPush,
	"RPUSH":            rPush,
	"LPOP":             lPop,
	"RPOP":             rPop,
	"LLEN":             llen,
	"LINDEX":           lIndex,
	"LINSERT":          lInsert,
	"LPUSHX":           lPushx,
	"RPUSHX":           rPushx,
	"LREM":             lRem,
	"LSET":             lSet,
	"LRANGE":           lRange,
	"LPOPRPUSH":        lPopRPush,
	"RPOPLPUSH":        rPopLPush,
	"ZADD":             zAdd,
	"ZCARD":            zCard,
	"ZRANK":            zRank,
	"ZREVRANK":         zRevRank,
	"ZSCORE":           zScore,
	"ZINCRBY":          zIncrBy,
	"ZRANGE":           zRange,
	"ZREVRANGE":        zRevRange,
	"ZRANGEBYSCORE":    zRangeByScore,
	"ZREVRANGEBYSCORE": zRevRangeByScore,
	"ZREM":             zRem,
	"ZREMRANGEBYRANK":  zRemRangeByRank,
	"ZREMRANGEBYSCORE": zRemRangeByScore,
	"ZCLEAR":           zClear,
	"ZEXISTS":          zExists,
}

func ping(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.StringValue("PONG")
	}
	return redis.BulkValue(args[0].Bulk)
}

func del(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("DEL requires at least one argument")
	}
	keys := make([]string, 0, len(args))
	for _, arg := range args {
		keys = append(keys, arg.Bulk)
	}
	return redis.IntegerValue(n.Del(keys...))
}

func exists(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("EXISTS requires at least one argument")
	}
	keys := make([]string, 0, len(args))
	for _, arg := range args {
		keys = append(keys, arg.Bulk)
	}
	return redis.IntegerValue(n.Exists(keys...))
}

func expire(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("EXPIRE requires at least two arguments")
	}
	key := args[0].Bulk
	seconds, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	return redis.IntegerValue(n.Expire(key, int64(seconds)))
}

func expireAt(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("EXPIREAT requires at least two arguments")
	}
	key := args[0].Bulk
	timestamp, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	return redis.IntegerValue(n.ExpireAt(key, time.Unix(int64(timestamp), 0)))
}

func keys(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("KEYS requires at least one argument")
	}
	pattern := args[0].Bulk
	keys := n.Keys(pattern)
	var k = make([]redis.Value, 0, len(keys))
	for _, v := range keys {
		k = append(k, redis.StringValue(v))
	}
	return redis.ArrayValue(k...)
}

func ttl(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("TTL requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(int64(n.TTL(key).Seconds()))
}

func rename(n *Nodis, args []redis.Value) redis.Value {
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

func typ(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("TYPE requires at least one argument")
	}
	key := args[0].Bulk
	return redis.StringValue(n.Type(key))
}

func scan(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("SCAN requires at least one argument")
	}
	cursor, _ := strconv.ParseInt(args[0].Bulk, 10, 64)
	var match string = "*"
	var count int64
	if len(args) > 1 {
		match = args[1].Bulk
	}
	if len(args) > 2 {
		count, _ = strconv.ParseInt(args[2].Bulk, 10, 64)
	}
	_, keys := n.Scan(cursor, match, count)
	var k = make([]redis.Value, 0, len(keys))
	for _, v := range keys {
		k = append(k, redis.StringValue(v))
	}
	return redis.ArrayValue(k...)
}

func setString(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("SET requires at least two arguments")
	}
	key := args[0].Bulk
	value := args[1].Bulk
	n.Set(key, []byte(value), 0)
	return redis.StringValue("OK")
}

func getString(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("GET requires at least one argument")
	}
	key := args[0].Bulk
	v := n.Get(key)
	return redis.StringValue(string(v))
}

func setBit(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("SETBIT requires at least two arguments")
	}
	key := args[0].Bulk
	offset, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	value, _ := strconv.ParseBool(args[2].Bulk)
	n.SetBit(key, offset, value)
	return redis.IntegerValue(1)
}

func getBit(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("GETBIT requires at least two arguments")
	}
	key := args[0].Bulk
	offset, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	return redis.IntegerValue(n.GetBit(key, offset))
}

func bitCount(n *Nodis, args []redis.Value) redis.Value {
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

func sAdd(n *Nodis, args []redis.Value) redis.Value {
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

func scard(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("SCARD requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.SCard(key))
}

func sDiff(n *Nodis, args []redis.Value) redis.Value {
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
		r = append(r, redis.StringValue(v))
	}
	return redis.ArrayValue(r...)
}

func sInter(n *Nodis, args []redis.Value) redis.Value {
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
		r = append(r, redis.StringValue(v))
	}
	return redis.ArrayValue(r...)
}

func sIsMember(n *Nodis, args []redis.Value) redis.Value {
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

func sMembers(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("SMEMBERS requires at least one argument")
	}
	key := args[0].Bulk
	results := n.SMembers(key)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.StringValue(v))
	}
	return redis.ArrayValue(r...)
}

func sRem(n *Nodis, args []redis.Value) redis.Value {
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

func hSet(n *Nodis, args []redis.Value) redis.Value {
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

func hGet(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("HGET requires at least two arguments")
	}
	key := args[0].Bulk
	field := args[1].Bulk
	return redis.StringValue(string(n.HGet(key, field)))
}

func hDel(n *Nodis, args []redis.Value) redis.Value {
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

func hLen(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HLEN requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.HLen(key))
}

func hKeys(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HKEYS requires at least one argument")
	}
	key := args[0].Bulk
	results := n.HKeys(key)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.StringValue(v))
	}
	return redis.ArrayValue(r...)
}

func hExists(n *Nodis, args []redis.Value) redis.Value {
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

func hGetAll(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HGETALL requires at least one argument")
	}
	key := args[0].Bulk
	results := n.HGetAll(key)
	var r = make(map[string]redis.Value)
	for k, v := range results {
		r[k] = redis.StringValue(string(v))
	}
	return redis.MapValue(r)
}

func hIncrBy(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("HINCRBY requires at least three arguments")
	}
	key := args[0].Bulk
	field := args[1].Bulk
	value, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	return redis.IntegerValue(n.HIncrBy(key, field, value))
}

func hIncrByFloat(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("HINCRBYFLOAT requires at least three arguments")
	}
	key := args[0].Bulk
	field := args[1].Bulk
	value, _ := strconv.ParseFloat(args[2].Bulk, 64)
	f := strconv.FormatFloat(n.HIncrByFloat(key, field, value), 'f', -1, 64)
	return redis.BulkValue(f)
}

func hSetNX(n *Nodis, args []redis.Value) redis.Value {
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

func hMGet(n *Nodis, args []redis.Value) redis.Value {
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
		r = append(r, redis.StringValue(string(v)))
	}
	return redis.ArrayValue(r...)
}

func hMSet(n *Nodis, args []redis.Value) redis.Value {
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

func hClear(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HCLEAR requires at least one argument")
	}
	key := args[0].Bulk
	n.HClear(key)
	return redis.StringValue("OK")
}

func hScan(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HSCAN requires at least one argument")
	}
	key := args[0].Bulk
	cursor, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	var match string = "*"
	var count int64
	if len(args) > 2 {
		match = args[2].Bulk
	}
	if len(args) > 3 {
		count, _ = strconv.ParseInt(args[3].Bulk, 10, 64)
	}
	_, results := n.HScan(key, cursor, match, count)
	var r = make([]redis.Value, 0, len(results)*2)
	for k, v := range results {
		r = append(r, redis.StringValue(k), redis.StringValue(string(v)))
	}
	return redis.ArrayValue(r...)
}

func hVals(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("HVALS requires at least one argument")
	}
	key := args[0].Bulk
	results := n.HVals(key)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.StringValue(string(v)))
	}
	return redis.ArrayValue(r...)
}

func lPush(n *Nodis, args []redis.Value) redis.Value {
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

func rPush(n *Nodis, args []redis.Value) redis.Value {
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

func lPop(n *Nodis, args []redis.Value) redis.Value {
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
		r = append(r, redis.StringValue(string(vv)))
	}
	return redis.ArrayValue(r...)
}

func rPop(n *Nodis, args []redis.Value) redis.Value {
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
		r = append(r, redis.StringValue(string(vv)))
	}
	return redis.ArrayValue(r...)
}

func llen(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("LLEN requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.LLen(key))
}

func lIndex(n *Nodis, args []redis.Value) redis.Value {
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

func lInsert(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 4 {
		return redis.ErrorValue("LINSERT requires at least four arguments")
	}
	key := args[0].Bulk
	before := strings.ToUpper(args[1].Bulk) == "BEFORE"
	pivot := []byte(args[2].Bulk)
	value := []byte(args[3].Bulk)
	return redis.IntegerValue(n.LInsert(key, pivot, value, before))
}

func lPushx(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("LPUSHX requires at least two arguments")
	}
	key := args[0].Bulk
	value := []byte(args[1].Bulk)
	return redis.IntegerValue(n.LPushX(key, value))
}

func rPushx(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("RPUSHX requires at least two arguments")
	}
	key := args[0].Bulk
	value := []byte(args[1].Bulk)
	return redis.IntegerValue(n.RPushX(key, value))
}

func lRem(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("LREM requires at least two arguments")
	}
	key := args[0].Bulk
	count, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	value := []byte(args[2].Bulk)
	return redis.IntegerValue(n.LRem(key, count, value))
}

func lSet(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("LSET requires at least three arguments")
	}
	key := args[0].Bulk
	index, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	value := []byte(args[2].Bulk)
	n.LSet(key, index, value)
	return redis.StringValue("OK")
}

func lRange(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 2 {
		return redis.ErrorValue("LRANGE requires at least two arguments")
	}
	key := args[0].Bulk
	start, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	end, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	results := n.LRange(key, start, end)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.StringValue(string(v)))
	}
	return redis.ArrayValue(r...)
}

func lPopRPush(n *Nodis, args []redis.Value) redis.Value {
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

func rPopLPush(n *Nodis, args []redis.Value) redis.Value {
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

func zAdd(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZADD requires at least three arguments")
	}
	key := args[0].Bulk
	p1 := strings.ToUpper(args[1].Bulk)
	if p1 == "INCR" {
		score, _ := strconv.ParseFloat(args[2].Bulk, 64)
		member := args[3].Bulk
		v := n.ZIncrBy(key, member, score)
		f := strconv.FormatFloat(v, 'f', -1, 64)
		return redis.BulkValue(f)
	}
	score, _ := strconv.ParseFloat(args[1].Bulk, 64)
	member := args[2].Bulk
	n.ZAdd(key, member, score)
	return redis.IntegerValue(1)
}

func zCard(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("ZCARD requires at least one argument")
	}
	key := args[0].Bulk
	return redis.IntegerValue(n.ZCard(key))
}

func zRank(n *Nodis, args []redis.Value) redis.Value {
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

func zRevRank(n *Nodis, args []redis.Value) redis.Value {
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

func zScore(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("ZSCORE requires at least two argument")
	}
	key := args[0].Bulk
	member := args[1].Bulk
	return redis.DoubleValue(n.ZScore(key, member))
}

func zIncrBy(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZINCRBY requires at least three arguments")
	}
	key := args[0].Bulk
	score, _ := strconv.ParseFloat(args[1].Bulk, 64)
	member := args[2].Bulk
	v := n.ZIncrBy(key, member, score)
	return redis.DoubleValue(v)
}

func zRange(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZRANGE requires at least three arguments")
	}
	key := args[0].Bulk
	start, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	stop, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	results := n.ZRange(key, start, stop)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.StringValue(v))
	}
	return redis.ArrayValue(r...)
}

func zRevRange(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZREVRANGE requires at least three arguments")
	}
	key := args[0].Bulk
	start, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	stop, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	results := n.ZRevRange(key, start, stop)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.StringValue(v))
	}
	return redis.ArrayValue(r...)
}

func zRangeByScore(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZRANGEBYSCORE requires at least three arguments")
	}
	key := args[0].Bulk
	min, _ := strconv.ParseFloat(args[1].Bulk, 64)
	max, _ := strconv.ParseFloat(args[1].Bulk, 64)
	results := n.ZRangeByScore(key, min, max)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.StringValue(v))
	}
	return redis.ArrayValue(r...)
}

func zRevRangeByScore(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZREVRANGEBYSCORE requires at least three arguments")
	}
	key := args[0].Bulk
	min, _ := strconv.ParseFloat(args[1].Bulk, 64)
	max, _ := strconv.ParseFloat(args[1].Bulk, 64)
	results := n.ZRevRangeByScore(key, min, max)
	var r = make([]redis.Value, 0, len(results))
	for _, v := range results {
		r = append(r, redis.StringValue(v))
	}
	return redis.ArrayValue(r...)
}

func zRem(n *Nodis, args []redis.Value) redis.Value {
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

func zRemRangeByRank(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZREMRANGEBYRANK requires at least three arguments")
	}
	key := args[0].Bulk
	start, _ := strconv.ParseInt(args[1].Bulk, 10, 64)
	stop, _ := strconv.ParseInt(args[2].Bulk, 10, 64)
	return redis.IntegerValue(n.ZRemRangeByRank(key, start, stop))
}

func zRemRangeByScore(n *Nodis, args []redis.Value) redis.Value {
	if len(args) < 3 {
		return redis.ErrorValue("ZREMRANGEBYSCORE requires at least three arguments")
	}
	key := args[0].Bulk
	min, _ := strconv.ParseFloat(args[1].Bulk, 64)
	max, _ := strconv.ParseFloat(args[1].Bulk, 64)
	return redis.IntegerValue(n.ZRemRangeByScore(key, min, max))
}

func zClear(n *Nodis, args []redis.Value) redis.Value {
	if len(args) == 0 {
		return redis.ErrorValue("ZCLEAR requires at least one argument")
	}
	key := args[0].Bulk
	n.ZClear(key)
	return redis.StringValue("OK")
}

func zExists(n *Nodis, args []redis.Value) redis.Value {
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
