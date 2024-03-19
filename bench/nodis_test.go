package bench

import (
	"strconv"
	"testing"

	"github.com/diiyw/nodis"
)

func BenchmarkSet(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/set",
		FileSize: nodis.FileSizeGB,
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.Set(id, []byte("value"+id), 0)
	}
}

func BenchmarkGet(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/get",
		FileSize: nodis.FileSizeGB,
	})
	for i := 0; i < 100000; i++ {
		id := strconv.Itoa(i)
		n.Set(id, []byte("value"+id), 0)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.Get(id)
	}
}

func BenchmarkLPush(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/lPush",
		FileSize: nodis.FileSizeGB,
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.LPush(id, []byte("value"+id))
	}
}

func BenchmarkLPop(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/lPop",
		FileSize: nodis.FileSizeGB,
	})
	for i := 0; i < 100000; i++ {
		id := strconv.Itoa(i)
		n.LPush(id, []byte("value"+id))
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.LPop(id)
	}
}

func BenchmarkSAdd(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/sAdd",
		FileSize: nodis.FileSizeGB,
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.SAdd(id, "value"+id)
	}
}

func BenchmarkSMembers(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/sMembers",
		FileSize: nodis.FileSizeGB,
	})
	for i := 0; i < 100000; i++ {
		id := strconv.Itoa(i)
		n.SAdd(id, "value"+id)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.SMembers(id)
	}
}

func BenchmarkZAdd(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/zAdd",
		FileSize: nodis.FileSizeGB,
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.ZAdd("key", "value"+id, float64(i))
	}
}

func BenchmarkZRank(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/zRank",
		FileSize: nodis.FileSizeGB,
	})
	for i := 0; i < 100000; i++ {
		id := strconv.Itoa(i)
		n.ZAdd(id, "value"+id, float64(i))
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.ZRank(id, "value"+id)
	}
}

func BenchmarkHSet(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/hSet",
		FileSize: nodis.FileSizeGB,
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.HSet(id, "field"+id, []byte("value"+id))
	}
}

func BenchmarkHGet(b *testing.B) {
	n := nodis.Open(&nodis.Options{
		Path:     "../testdata/hGet",
		FileSize: nodis.FileSizeGB,
	})
	for i := 0; i < 100000; i++ {
		id := strconv.Itoa(i)
		n.HSet(id, "field"+id, []byte("value"+id))
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i)
		n.HGet(id, "field"+id)
	}
}
