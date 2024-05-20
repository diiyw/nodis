package geo

import (
	"fmt"
	"strconv"

	"github.com/mmcloughlin/geohash"
)

type Member struct {
	Longitude float64
	Latitude  float64
	Name      string
}

func (m *Member) Hash() uint64 {
	return geohash.EncodeInt(m.Latitude, m.Longitude)
}

func Parse(name, lat, lon string) (*Member, error) {
	l, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return nil, err
	}
	n, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		return nil, err
	}
	if l < -90 || l > 90 || n < -180 || n > 180 {
		return nil, fmt.Errorf("ERR invalid longitude,latitude pair %s,%s", lat, lon)
	}
	return &Member{
		Latitude:  l,
		Longitude: n,
		Name:      name,
	}, nil
}
