package geohash

import "errors"

const (
	/* These are constraints from EPSG:900913 / EPSG:3785 / OSGEO:41001
	We can't geocode at the north/south pole.*/
	WGS84_LAT_MIN  = -85.05112878
	WGS84_LAT_MAX  = 85.05112878
	WGS84_LONG_MIN = -180
	WGS84_LONG_MAX = 180

	/* Use 26*2 = 52bits to encode a position in WGS84, the precision is 0.6m.*/
	WGS84_GEO_STEP = 26
)

var (
	WGS84_LONG_RANGE = &Range{Max: WGS84_LONG_MAX, Min: WGS84_LONG_MIN}
	WGS84_LAT_RANGE  = &Range{Max: WGS84_LAT_MAX, Min: WGS84_LAT_MIN}
)

var (
	ErrPositionOutOfRange = errors.New("position out of range of WGS84")
	ErrStepOutOfRange     = errors.New("geohash encode step must less-equal than 32 and greater than 0")
)

func EncodeWGS84(longitude, latitude float64) (uint64, error) {
	if hash, err := Encode(
		WGS84_LONG_RANGE,
		WGS84_LAT_RANGE,
		longitude, latitude,
		WGS84_GEO_STEP); err != nil {
		return 0, err
	} else {
		return hash.Bits, nil
	}
}

func DecodeWGS84(bits uint64) *Area {
	return decode(
		WGS84_LONG_RANGE,
		WGS84_LAT_RANGE,
		HashBits{Bits: bits, Step: WGS84_GEO_STEP},
	)
}

func DecodeToLongLatWGS84(bits uint64) (float64, float64) {
	area := DecodeWGS84(bits)
	return DecodeAreaToLongLat(area)
}

func Encode(
	longRange, latRange *Range,
	longitude, latitude float64,
	step uint8) (HashBits, error) {

	hash := HashBits{}

	if step > 32 || step == 0 {
		return hash, ErrStepOutOfRange
	}

	/* Return an error when trying to index outside the supported constraints. */
	if longitude > WGS84_LONG_MAX || longitude < WGS84_LONG_MIN ||
		latitude > WGS84_LAT_MAX || latitude < WGS84_LAT_MIN {
		return hash, ErrPositionOutOfRange
	}

	if longitude > longRange.Max || longitude < longRange.Min ||
		latitude > latRange.Max || latitude < latRange.Min {
		return hash, ErrPositionOutOfRange
	}

	var latOffset float64 = (latitude - latRange.Min) / (latRange.Max - latRange.Min)
	var lonOffset float64 = (longitude - longRange.Min) / (longRange.Max - longRange.Min)

	var x uint64 = (1 << step)
	latOffset *= float64(x)
	lonOffset *= float64(x)

	hash.Bits = interleave64(uint32(latOffset), uint32(lonOffset))
	hash.Step = step
	return hash, nil
}

func DecodeAreaToLongLat(area *Area) (float64, float64) {
	return (area.Longitude.Max + area.Longitude.Min) / 2,
		(area.Latitude.Max + area.Latitude.Min) / 2
}

func decode(lonRange, latRange *Range, hash HashBits) *Area {
	latScale := latRange.Max - latRange.Min
	lonScale := lonRange.Max - lonRange.Min

	ilato, ilono := deinterleave64(hash.Bits)

	area := &Area{
		Hash: hash,
	}

	var x float64 = float64((uint64(1) << hash.Step))

	area.Latitude.Min =
		latRange.Min + (float64(ilato)*1.0/x)*latScale
	area.Latitude.Max =
		latRange.Min + ((float64(ilato)+1)*1.0/x)*latScale

	area.Longitude.Min =
		lonRange.Min + (float64(ilono)*1.0/x)*lonScale
	area.Longitude.Max =
		lonRange.Min + ((float64(ilono)+1)*1.0/x)*lonScale

	return area
}
