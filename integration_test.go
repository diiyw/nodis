package nodis

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// startTestServer starts a Nodis server on a random available port and returns the client and cleanup function
func startTestServer(t *testing.T) (*redis.Client, func()) {
	t.Helper()
	_ = os.RemoveAll("testdata")

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	opt := DefaultOptions
	n := Open(opt)

	// Start server in background
	go func() {
		_ = n.Serve(addr)
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	cleanup := func() {
		client.Close()
		n.Close()
		os.RemoveAll("testdata")
	}

	return client, cleanup
}

func TestIntegration_String(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Test SET and GET
	err := client.Set(ctx, "key1", "value1", 0).Err()
	if err != nil {
		t.Fatalf("SET failed: %v", err)
	}

	val, err := client.Get(ctx, "key1").Result()
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("GET = %v, want %v", val, "value1")
	}

	// Test SETNX
	ok, err := client.SetNX(ctx, "key1", "value2", 0).Result()
	if err != nil {
		t.Fatalf("SETNX failed: %v", err)
	}
	if ok {
		t.Error("SETNX should return false for existing key")
	}

	ok, err = client.SetNX(ctx, "key2", "value2", 0).Result()
	if err != nil {
		t.Fatalf("SETNX failed: %v", err)
	}
	if !ok {
		t.Error("SETNX should return true for new key")
	}

	// Test INCR/DECR
	err = client.Set(ctx, "counter", "10", 0).Err()
	if err != nil {
		t.Fatalf("SET counter failed: %v", err)
	}

	n, err := client.Incr(ctx, "counter").Result()
	if err != nil {
		t.Fatalf("INCR failed: %v", err)
	}
	if n != 11 {
		t.Errorf("INCR = %v, want %v", n, 11)
	}

	n, err = client.Decr(ctx, "counter").Result()
	if err != nil {
		t.Fatalf("DECR failed: %v", err)
	}
	if n != 10 {
		t.Errorf("DECR = %v, want %v", n, 10)
	}

	// Test APPEND
	length, err := client.Append(ctx, "key1", "_appended").Result()
	if err != nil {
		t.Fatalf("APPEND failed: %v", err)
	}
	// "value1" (6) + "_appended" (9) = 15
	if length != 15 {
		t.Errorf("APPEND length = %v, want %v", length, 15)
	}

	val, _ = client.Get(ctx, "key1").Result()
	if val != "value1_appended" {
		t.Errorf("GET after APPEND = %v, want %v", val, "value1_appended")
	}

	// Test STRLEN
	strLen, err := client.StrLen(ctx, "key1").Result()
	if err != nil {
		t.Fatalf("STRLEN failed: %v", err)
	}
	if strLen != 15 {
		t.Errorf("STRLEN = %v, want %v", strLen, 15)
	}

	// Test MSET/MGET
	err = client.MSet(ctx, "mkey1", "mval1", "mkey2", "mval2").Err()
	if err != nil {
		t.Fatalf("MSET failed: %v", err)
	}

	vals, err := client.MGet(ctx, "mkey1", "mkey2", "nonexistent").Result()
	if err != nil {
		t.Fatalf("MGET failed: %v", err)
	}
	if vals[0] != "mval1" || vals[1] != "mval2" || vals[2] != nil {
		t.Errorf("MGET = %v, want [mval1, mval2, nil]", vals)
	}
}

func TestIntegration_List(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Test LPUSH and RPUSH
	n, err := client.LPush(ctx, "list1", "a", "b", "c").Result()
	if err != nil {
		t.Fatalf("LPUSH failed: %v", err)
	}
	if n != 3 {
		t.Errorf("LPUSH = %v, want %v", n, 3)
	}

	n, err = client.RPush(ctx, "list1", "d", "e").Result()
	if err != nil {
		t.Fatalf("RPUSH failed: %v", err)
	}
	if n != 5 {
		t.Errorf("RPUSH = %v, want %v", n, 5)
	}

	// Test LRANGE
	vals, err := client.LRange(ctx, "list1", 0, -1).Result()
	if err != nil {
		t.Fatalf("LRANGE failed: %v", err)
	}
	expected := []string{"c", "b", "a", "d", "e"}
	if len(vals) != len(expected) {
		t.Errorf("LRANGE length = %v, want %v", len(vals), len(expected))
	}
	for i, v := range vals {
		if v != expected[i] {
			t.Errorf("LRANGE[%d] = %v, want %v", i, v, expected[i])
		}
	}

	// Test LLEN
	length, err := client.LLen(ctx, "list1").Result()
	if err != nil {
		t.Fatalf("LLEN failed: %v", err)
	}
	if length != 5 {
		t.Errorf("LLEN = %v, want %v", length, 5)
	}

	// Test LPOP and RPOP
	val, err := client.LPop(ctx, "list1").Result()
	if err != nil {
		t.Fatalf("LPOP failed: %v", err)
	}
	if val != "c" {
		t.Errorf("LPOP = %v, want %v", val, "c")
	}

	val, err = client.RPop(ctx, "list1").Result()
	if err != nil {
		t.Fatalf("RPOP failed: %v", err)
	}
	if val != "e" {
		t.Errorf("RPOP = %v, want %v", val, "e")
	}

	// Test LINDEX
	val, err = client.LIndex(ctx, "list1", 0).Result()
	if err != nil {
		t.Fatalf("LINDEX failed: %v", err)
	}
	if val != "b" {
		t.Errorf("LINDEX = %v, want %v", val, "b")
	}

	// Test LSET
	err = client.LSet(ctx, "list1", 0, "new_b").Err()
	if err != nil {
		t.Fatalf("LSET failed: %v", err)
	}

	val, _ = client.LIndex(ctx, "list1", 0).Result()
	if val != "new_b" {
		t.Errorf("LINDEX after LSET = %v, want %v", val, "new_b")
	}
}

func TestIntegration_Hash(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Test HSET and HGET
	n, err := client.HSet(ctx, "hash1", "field1", "value1").Result()
	if err != nil {
		t.Fatalf("HSET failed: %v", err)
	}
	if n != 1 {
		t.Errorf("HSET = %v, want %v", n, 1)
	}

	val, err := client.HGet(ctx, "hash1", "field1").Result()
	if err != nil {
		t.Fatalf("HGET failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("HGET = %v, want %v", val, "value1")
	}

	// Test HMSET and HMGET
	err = client.HMSet(ctx, "hash1", "field2", "value2", "field3", "value3").Err()
	if err != nil {
		t.Fatalf("HMSET failed: %v", err)
	}

	vals, err := client.HMGet(ctx, "hash1", "field1", "field2", "field3", "nonexistent").Result()
	if err != nil {
		t.Fatalf("HMGET failed: %v", err)
	}
	if vals[0] != "value1" || vals[1] != "value2" || vals[2] != "value3" || vals[3] != nil {
		t.Errorf("HMGET = %v, want [value1, value2, value3, nil]", vals)
	}

	// Test HGETALL
	all, err := client.HGetAll(ctx, "hash1").Result()
	if err != nil {
		t.Fatalf("HGETALL failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("HGETALL length = %v, want %v", len(all), 3)
	}

	// Test HDEL
	deleted, err := client.HDel(ctx, "hash1", "field1").Result()
	if err != nil {
		t.Fatalf("HDEL failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("HDEL = %v, want %v", deleted, 1)
	}

	// Test HEXISTS
	exists, err := client.HExists(ctx, "hash1", "field1").Result()
	if err != nil {
		t.Fatalf("HEXISTS failed: %v", err)
	}
	if exists {
		t.Error("HEXISTS should return false for deleted field")
	}

	exists, err = client.HExists(ctx, "hash1", "field2").Result()
	if err != nil {
		t.Fatalf("HEXISTS failed: %v", err)
	}
	if !exists {
		t.Error("HEXISTS should return true for existing field")
	}

	// Test HLEN
	hlen, err := client.HLen(ctx, "hash1").Result()
	if err != nil {
		t.Fatalf("HLEN failed: %v", err)
	}
	if hlen != 2 {
		t.Errorf("HLEN = %v, want %v", hlen, 2)
	}

	// Test HINCRBY
	client.HSet(ctx, "hash1", "counter", "10")
	newVal, err := client.HIncrBy(ctx, "hash1", "counter", 5).Result()
	if err != nil {
		t.Fatalf("HINCRBY failed: %v", err)
	}
	if newVal != 15 {
		t.Errorf("HINCRBY = %v, want %v", newVal, 15)
	}
}

func TestIntegration_Set(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Test SADD
	n, err := client.SAdd(ctx, "set1", "a", "b", "c").Result()
	if err != nil {
		t.Fatalf("SADD failed: %v", err)
	}
	if n != 3 {
		t.Errorf("SADD = %v, want %v", n, 3)
	}

	// Test SISMEMBER
	isMember, err := client.SIsMember(ctx, "set1", "a").Result()
	if err != nil {
		t.Fatalf("SISMEMBER failed: %v", err)
	}
	if !isMember {
		t.Error("SISMEMBER should return true for existing member")
	}

	isMember, err = client.SIsMember(ctx, "set1", "d").Result()
	if err != nil {
		t.Fatalf("SISMEMBER failed: %v", err)
	}
	if isMember {
		t.Error("SISMEMBER should return false for non-existing member")
	}

	// Test SMEMBERS
	members, err := client.SMembers(ctx, "set1").Result()
	if err != nil {
		t.Fatalf("SMEMBERS failed: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("SMEMBERS length = %v, want %v", len(members), 3)
	}

	// Test SCARD
	card, err := client.SCard(ctx, "set1").Result()
	if err != nil {
		t.Fatalf("SCARD failed: %v", err)
	}
	if card != 3 {
		t.Errorf("SCARD = %v, want %v", card, 3)
	}

	// Test SREM
	removed, err := client.SRem(ctx, "set1", "a").Result()
	if err != nil {
		t.Fatalf("SREM failed: %v", err)
	}
	if removed != 1 {
		t.Errorf("SREM = %v, want %v", removed, 1)
	}

	card, _ = client.SCard(ctx, "set1").Result()
	if card != 2 {
		t.Errorf("SCARD after SREM = %v, want %v", card, 2)
	}

	// Test set operations
	client.SAdd(ctx, "set2", "b", "c", "d")

	// SINTER
	inter, err := client.SInter(ctx, "set1", "set2").Result()
	if err != nil {
		t.Fatalf("SINTER failed: %v", err)
	}
	if len(inter) != 2 { // b, c
		t.Errorf("SINTER length = %v, want %v", len(inter), 2)
	}

	// SUNION
	union, err := client.SUnion(ctx, "set1", "set2").Result()
	if err != nil {
		t.Fatalf("SUNION failed: %v", err)
	}
	if len(union) != 3 { // b, c, d
		t.Errorf("SUNION length = %v, want %v", len(union), 3)
	}

	// SDIFF
	diff, err := client.SDiff(ctx, "set2", "set1").Result()
	if err != nil {
		t.Fatalf("SDIFF failed: %v", err)
	}
	if len(diff) != 1 { // d
		t.Errorf("SDIFF length = %v, want %v", len(diff), 1)
	}
}

func TestIntegration_SortedSet(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Test ZADD
	n, err := client.ZAdd(ctx, "zset1", redis.Z{Score: 1, Member: "a"}, redis.Z{Score: 2, Member: "b"}, redis.Z{Score: 3, Member: "c"}).Result()
	if err != nil {
		t.Fatalf("ZADD failed: %v", err)
	}
	if n != 3 {
		t.Errorf("ZADD = %v, want %v", n, 3)
	}

	// Test ZSCORE
	score, err := client.ZScore(ctx, "zset1", "b").Result()
	if err != nil {
		t.Fatalf("ZSCORE failed: %v", err)
	}
	if score != 2 {
		t.Errorf("ZSCORE = %v, want %v", score, 2)
	}

	// Test ZRANK
	rank, err := client.ZRank(ctx, "zset1", "b").Result()
	if err != nil {
		t.Fatalf("ZRANK failed: %v", err)
	}
	if rank != 1 {
		t.Errorf("ZRANK = %v, want %v", rank, 1)
	}

	// Test ZRANGE
	members, err := client.ZRange(ctx, "zset1", 0, -1).Result()
	if err != nil {
		t.Fatalf("ZRANGE failed: %v", err)
	}
	expected := []string{"a", "b", "c"}
	if len(members) != len(expected) {
		t.Errorf("ZRANGE length = %v, want %v", len(members), len(expected))
	}
	for i, m := range members {
		if m != expected[i] {
			t.Errorf("ZRANGE[%d] = %v, want %v", i, m, expected[i])
		}
	}

	// Test ZREVRANGE
	members, err = client.ZRevRange(ctx, "zset1", 0, -1).Result()
	if err != nil {
		t.Fatalf("ZREVRANGE failed: %v", err)
	}
	expected = []string{"c", "b", "a"}
	for i, m := range members {
		if m != expected[i] {
			t.Errorf("ZREVRANGE[%d] = %v, want %v", i, m, expected[i])
		}
	}

	// Test ZCARD
	card, err := client.ZCard(ctx, "zset1").Result()
	if err != nil {
		t.Fatalf("ZCARD failed: %v", err)
	}
	if card != 3 {
		t.Errorf("ZCARD = %v, want %v", card, 3)
	}

	// Test ZINCRBY
	newScore, err := client.ZIncrBy(ctx, "zset1", 2.5, "a").Result()
	if err != nil {
		t.Fatalf("ZINCRBY failed: %v", err)
	}
	if newScore != 3.5 {
		t.Errorf("ZINCRBY = %v, want %v", newScore, 3.5)
	}

	// Test ZREM
	removed, err := client.ZRem(ctx, "zset1", "a").Result()
	if err != nil {
		t.Fatalf("ZREM failed: %v", err)
	}
	if removed != 1 {
		t.Errorf("ZREM = %v, want %v", removed, 1)
	}

	// Test ZCOUNT
	count, err := client.ZCount(ctx, "zset1", "1", "3").Result()
	if err != nil {
		t.Fatalf("ZCOUNT failed: %v", err)
	}
	if count != 2 { // b(2), c(3)
		t.Errorf("ZCOUNT = %v, want %v", count, 2)
	}
}

func TestIntegration_Keys(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Setup some keys
	client.Set(ctx, "key1", "value1", 0)
	client.Set(ctx, "key2", "value2", 0)
	client.Set(ctx, "other", "value3", 0)

	// Test EXISTS
	n, err := client.Exists(ctx, "key1", "key2", "nonexistent").Result()
	if err != nil {
		t.Fatalf("EXISTS failed: %v", err)
	}
	if n != 2 {
		t.Errorf("EXISTS = %v, want %v", n, 2)
	}

	// Test DEL
	deleted, err := client.Del(ctx, "key1").Result()
	if err != nil {
		t.Fatalf("DEL failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("DEL = %v, want %v", deleted, 1)
	}

	// Test KEYS
	keys, err := client.Keys(ctx, "key*").Result()
	if err != nil {
		t.Fatalf("KEYS failed: %v", err)
	}
	if len(keys) != 1 { // only key2 remains
		t.Errorf("KEYS length = %v, want %v", len(keys), 1)
	}

	// Test RENAME
	err = client.Rename(ctx, "key2", "key2_renamed").Err()
	if err != nil {
		t.Fatalf("RENAME failed: %v", err)
	}

	val, err := client.Get(ctx, "key2_renamed").Result()
	if err != nil {
		t.Fatalf("GET after RENAME failed: %v", err)
	}
	if val != "value2" {
		t.Errorf("GET after RENAME = %v, want %v", val, "value2")
	}

	// Test TYPE
	typ, err := client.Type(ctx, "key2_renamed").Result()
	if err != nil {
		t.Fatalf("TYPE failed: %v", err)
	}
	if typ != "string" {
		t.Errorf("TYPE = %v, want %v", typ, "string")
	}

	// Test TTL and EXPIRE
	client.Set(ctx, "expiring", "value", 0)
	ttl, err := client.TTL(ctx, "expiring").Result()
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	// go-redis v9 returns -1 and -2 as sentinel values (not multiplied by precision)
	// -1 means key exists but has no expiration
	if ttl != time.Duration(-1) {
		t.Errorf("TTL = %v, want %v", ttl, time.Duration(-1))
	}

	ok, err := client.Expire(ctx, "expiring", 10*time.Second).Result()
	if err != nil {
		t.Fatalf("EXPIRE failed: %v", err)
	}
	if !ok {
		t.Error("EXPIRE should return true")
	}

	ttl, _ = client.TTL(ctx, "expiring").Result()
	if ttl <= 0 || ttl > 10*time.Second {
		t.Errorf("TTL after EXPIRE = %v, want between 0 and 10s", ttl)
	}

	// Test PERSIST
	ok, err = client.Persist(ctx, "expiring").Result()
	if err != nil {
		t.Fatalf("PERSIST failed: %v", err)
	}
	if !ok {
		t.Error("PERSIST should return true")
	}

	ttl, _ = client.TTL(ctx, "expiring").Result()
	// go-redis v9 returns -1 as sentinel value
	if ttl != time.Duration(-1) {
		t.Errorf("TTL after PERSIST = %v, want %v", ttl, time.Duration(-1))
	}
}

func TestIntegration_PingAndEcho(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Test PING
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("PING failed: %v", err)
	}
	if pong != "PONG" {
		t.Errorf("PING = %v, want %v", pong, "PONG")
	}

	// Test ECHO
	echo, err := client.Echo(ctx, "Hello").Result()
	if err != nil {
		t.Fatalf("ECHO failed: %v", err)
	}
	if echo != "Hello" {
		t.Errorf("ECHO = %v, want %v", echo, "Hello")
	}
}

func TestIntegration_DBSize(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Get initial size (may not be 0 due to storage initialization)
	initialSize, err := client.DBSize(ctx).Result()
	if err != nil {
		t.Fatalf("DBSIZE failed: %v", err)
	}

	// Add some keys
	client.Set(ctx, "dbsize_key1", "value1", 0)
	client.Set(ctx, "dbsize_key2", "value2", 0)
	client.LPush(ctx, "dbsize_list1", "a")

	size, err := client.DBSize(ctx).Result()
	if err != nil {
		t.Fatalf("DBSIZE failed: %v", err)
	}
	if size != initialSize+3 {
		t.Errorf("DBSIZE = %v, want %v (initial %v + 3)", size, initialSize+3, initialSize)
	}
}
