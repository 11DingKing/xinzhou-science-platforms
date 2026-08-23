package geo

import (
	"errors"
	"math"
	"sort"
	"strings"
)

type Point struct{ Latitude, Longitude float64 }
type Region struct {
	Name     string
	Center   Point
	RadiusKM float64
	Active   bool
}

func (p Point) Valid() bool {
	return p.Latitude >= -90 && p.Latitude <= 90 && p.Longitude >= -180 && p.Longitude <= 180
}
func DistanceKM(a, b Point) float64 {
	const earth = 6371
	rad := math.Pi / 180
	lat1 := a.Latitude * rad
	lat2 := b.Latitude * rad
	dlat := (b.Latitude - a.Latitude) * rad
	dlon := (b.Longitude - a.Longitude) * rad
	h := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	return earth * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}
func Match(regions []Region, p Point) []Region {
	out := []Region{}
	if !p.Valid() {
		return out
	}
	for _, r := range regions {
		if r.Active && r.Center.Valid() && r.RadiusKM >= 0 && DistanceKM(r.Center, p) <= r.RadiusKM {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
func NormalizeName(v string) string { return strings.ToLower(strings.Join(strings.Fields(v), " ")) }
func Validate(r Region) error {
	if strings.TrimSpace(r.Name) == "" || !r.Center.Valid() || r.RadiusKM < 0 {
		return errors.New("invalid region")
	}
	return nil
}
