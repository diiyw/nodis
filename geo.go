package nodis

import (
	"math"

	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/geo"
	"github.com/mmcloughlin/geohash"
)

var (
	earthRadius = 6372797.560856
	dr          = math.Pi / 180.0
)

func (n *Nodis) GeoAdd(key string, members ...*geo.Member) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		for _, member := range members {
			v += meta.value.(*zset.SortedSet).ZAdd(member.Name, float64(member.Hash()))
		}
		return nil
	})
	return v
}

// GeoAddXX adds the specified members to the key only if the member already exists.
func (n *Nodis) GeoAddXX(key string, members ...*geo.Member) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		for _, member := range members {
			v += meta.value.(*zset.SortedSet).ZAddXX(member.Name, float64(member.Hash()))
		}
		return nil
	})
	return v
}

// GeoAddNX adds the specified members to the key only if the member does not already exist.
func (n *Nodis) GeoAddNX(key string, members ...*geo.Member) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		if meta.isOk() {
			return nil
		}
		for _, member := range members {
			v += meta.value.(*zset.SortedSet).ZAddNX(member.Name, float64(member.Hash()))
		}
		return nil
	})
	return v
}

func (n *Nodis) GeoDist(key string, member1, member2 string) (float64, error) {
	var v float64
	err := n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		score1, err := meta.value.(*zset.SortedSet).ZScore(member1)
		if err != nil {
			return err
		}
		score2, err := meta.value.(*zset.SortedSet).ZScore(member2)
		if err != nil {
			return err
		}
		lat1, lng1 := geohash.DecodeInt(uint64(score1))
		lat2, lng2 := geohash.DecodeInt(uint64(score2))
		v = distance(lat1, lng1, lat2, lng2)
		return nil
	})
	return v, err
}

// distance computes the distance between two given coordinates in meter
func distance(latitude1, longitude1, latitude2, longitude2 float64) float64 {
	radLat1 := latitude1 * dr
	radLat2 := latitude2 * dr
	a := radLat1 - radLat2
	b := longitude1*dr - longitude2*dr
	return 2 * earthRadius * math.Asin(math.Sqrt(math.Pow(math.Sin(a/2), 2)+
		math.Cos(radLat1)*math.Cos(radLat2)*math.Pow(math.Sin(b/2), 2)))
}

func (n *Nodis) GeoHash(key string, members ...string) []string {
	var v []string
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		for _, member := range members {
			score, err := meta.value.(*zset.SortedSet).ZScore(member)
			if err != nil {
				return err
			}
			lat, lng := geohash.DecodeInt(uint64(score))
			v = append(v, geohash.Encode(lat, lng))
		}
		return nil
	})
	return v
}

func (n *Nodis) GeoPos(key string, members ...string) []*geo.Member {
	var v = make([]*geo.Member, len(members))
	_ = n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if meta == nil {
			return nil
		}
		for i, member := range members {
			score, err := meta.value.(*zset.SortedSet).ZScore(member)
			if err != nil {
				continue
			}
			lat, lng := geohash.DecodeInt(uint64(score))
			v[i] = &geo.Member{Name: member, Latitude: lat, Longitude: lng}
		}
		return nil
	})
	return v
}

func (n *Nodis) GeoRadius(key string, longitude, latitude, radius float64, count int64, desc bool) (map[string]*geo.Member, error) {
	var v = make(map[string]*geo.Member)
	err := n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		bits := estimatePrecisionByRadius(radius, latitude)
		hash := geohash.EncodeInt(latitude, longitude)
		neighbors := geohash.NeighborsInt(hash)
		for _, lower := range neighbors {
			var items []*zset.Item
			r := uint64(1 << (64 - bits))
			upper := lower + r
			if desc {
				items = meta.value.(*zset.SortedSet).ZRangeByScore(float64(lower), float64(upper), 0, count, 0)
			} else {
				items = meta.value.(*zset.SortedSet).ZRevRangeByScore(float64(lower), float64(upper), 0, count, 0)
			}
			for _, item := range items {
				lat, lng := geohash.DecodeInt(uint64(item.Score))
				v[item.Member] = &geo.Member{Name: item.Member, Latitude: lat, Longitude: lng}
			}
		}
		return nil
	})
	return v, err
}

const (
	MercatorMax = 20037726.37
)

func estimatePrecisionByRadius(radiusMeters float64, lat float64) uint {
	if radiusMeters == 0 {
		return 63
	}
	var precision uint = 1
	for radiusMeters < MercatorMax {
		radiusMeters *= 2
		precision++
	}
	/* Make sure range is included in most of the base cases. */
	precision -= 2
	if lat > 66 || lat < -66 {
		precision--
		if lat > 80 || lat < -80 {
			precision--
		}
	}
	if precision < 1 {
		precision = 1
	}
	if precision > 32 {
		precision = 32
	}
	return precision*2 - 1
}
