package nodis

import (
	"github.com/diiyw/nodis/ds/zset"
	"github.com/diiyw/nodis/internal/geohash"
)

type GeoMember struct {
	Longitude float64
	Latitude  float64
	Score     float64
	Member    string
}

func (m *GeoMember) Hash() uint64 {
	v, _ := geohash.EncodeWGS84(m.Longitude, m.Latitude)
	return v
}

func (n *Nodis) GeoAdd(key string, members ...*GeoMember) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		for _, member := range members {
			v += meta.value.(*zset.SortedSet).ZAdd(member.Member, float64(member.Hash()))
		}
		return nil
	})
	return v
}

// GeoAddXX adds the specified members to the key only if the member already exists.
func (n *Nodis) GeoAddXX(key string, members ...*GeoMember) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, nil)
		if !meta.isOk() {
			return nil
		}
		for _, member := range members {
			v += meta.value.(*zset.SortedSet).ZAddXX(member.Member, float64(member.Hash()))
		}
		return nil
	})
	return v
}

// GeoAddNX adds the specified members to the key only if the member does not already exist.
func (n *Nodis) GeoAddNX(key string, members ...*GeoMember) int64 {
	var v int64
	_ = n.exec(func(tx *Tx) error {
		meta := tx.writeKey(key, n.newZSet)
		if meta.isOk() {
			return nil
		}
		for _, member := range members {
			v += meta.value.(*zset.SortedSet).ZAddNX(member.Member, float64(member.Hash()))
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
		v = geohash.DistBetweenGeoHashWGS84(uint64(score1), uint64(score2))
		return nil
	})
	return v, err
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
			latitude, longitude := geohash.DecodeToLongLatWGS84(uint64(score))
			code, _ := geohash.Encode(
				&geohash.Range{Max: 180, Min: -180},
				&geohash.Range{Max: 90, Min: -90},
				longitude,
				latitude,
				geohash.WGS84_GEO_STEP,
			)
			v = append(v, string(geohash.EncodeToBase32(code.Bits)))
		}
		return nil
	})
	return v
}

func (n *Nodis) GeoPos(key string, members ...string) []*GeoMember {
	var v = make([]*GeoMember, len(members))
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
			lat, lng := geohash.DecodeToLongLatWGS84(uint64(score))
			v[i] = &GeoMember{Member: member, Latitude: lat, Longitude: lng}
		}
		return nil
	})
	return v
}

func (n *Nodis) GeoRadius(key string, longitude, latitude, radiusMeters float64, count int64, desc bool) ([]*GeoMember, error) {
	var v []*GeoMember
	err := n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		radiusArea, err := geohash.GetAreasByRadiusWGS84(longitude, latitude, radiusMeters)
		if err != nil {
			return err
		}
		v = n.geoMembersOfAllNeighbors(meta.value.(*zset.SortedSet), radiusArea, longitude, latitude, radiusMeters, count)
		return nil
	})
	return v, err
}

func (n *Nodis) GeoRadiusByMember(key, member string, radiusMeters float64, count int64, desc bool) ([]*GeoMember, error) {
	var v []*GeoMember
	err := n.exec(func(tx *Tx) error {
		meta := tx.readKey(key)
		if !meta.isOk() {
			return nil
		}
		score, err := meta.value.(*zset.SortedSet).ZScore(member)
		if err != nil {
			return err
		}
		lat, lng := geohash.DecodeToLongLatWGS84(uint64(score))
		radiusArea, err := geohash.GetAreasByRadiusWGS84(lng, lat, radiusMeters)
		if err != nil {
			return err
		}
		v = n.geoMembersOfAllNeighbors(meta.value.(*zset.SortedSet), radiusArea, lng, lat, radiusMeters, count)
		return nil
	})
	return v, err
}

func (n *Nodis) geoMembersOfAllNeighbors(v *zset.SortedSet, geoRadius *geohash.Radius, lon, lat, radius float64, count int64) []*GeoMember {
	neighbors := [9]*geohash.HashBits{
		&geoRadius.Hash,
		&geoRadius.North,
		&geoRadius.South,
		&geoRadius.East,
		&geoRadius.West,
		&geoRadius.NorthEast,
		&geoRadius.NorthWest,
		&geoRadius.SouthEast,
		&geoRadius.SouthWest,
	}

	var lastProcessed int = 0
	plist := make([]*GeoMember, 0, 64)
	for i, area := range neighbors {
		if area.IsZero() {
			continue
		}
		// When a huge Radius (in the 5000 km range or more) is used,
		// adjacent neighbors can be the same, leading to duplicated
		// elements. Skip every range which is the same as the one
		// processed previously.
		if lastProcessed != 0 &&
			area.Bits == neighbors[lastProcessed].Bits &&
			area.Step == neighbors[lastProcessed].Step {
			continue
		}
		ps := n.membersOfGeoHashBox(v, lon, lat, radius, area, count)
		plist = append(plist, ps...)
		lastProcessed = i
	}
	return plist
}

// Compute the sorted set scores min (inclusive), max (exclusive) we should
// query in order to retrieve all the elements inside the specified area
// 'hash'. The two scores are returned by reference in *min and *max.
func scoresOfGeoHashBox(hash *geohash.HashBits) (min, max uint64) {
	min = hash.Bits << (geohash.WGS84_GEO_STEP*2 - hash.Step*2)
	bits := hash.Bits + 1
	max = bits << (geohash.WGS84_GEO_STEP*2 - hash.Step*2)
	return
}

// Obtain all members between the min/max of this geohash bounding box.
func (n *Nodis) membersOfGeoHashBox(v *zset.SortedSet, longitude, latitude, radius float64, hash *geohash.HashBits, count int64) []*GeoMember {
	points := make([]*GeoMember, 0, 32)
	min, max := scoresOfGeoHashBox(hash)
	vlist := v.ZRangeByScore(float64(min), float64(max), 0, count, 0)
	for _, v := range vlist {
		x, y := geohash.DecodeToLongLatWGS84(uint64(v.Score))
		dist := geohash.GetDistance(x, y, longitude, latitude)
		if radius >= dist {
			p := &GeoMember{
				Longitude: x,
				Latitude:  y,
				Score:     v.Score,
				Member:    v.Member,
			}
			points = append(points, p)
		}
	}
	return points
}
