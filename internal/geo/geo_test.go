package geo

import (
	"testing"
)

func TestGeoMatching(t *testing.T) {
	if DistanceKM(Point{0, 0}, Point{0, 0}) != 0 {
		t.Fatal("distance")
	}
	regions := []Region{{Name: "city", Center: Point{0, 0}, RadiusKM: 10, Active: true}, {Name: "inactive", Center: Point{0, 0}, RadiusKM: 10, Active: false}}
	if len(Match(regions, Point{0.01, 0})) != 1 {
		t.Fatal("match")
	}
	if Validate(Region{Name: "city", Center: Point{0, 0}, RadiusKM: 10}) != nil {
		t.Fatal("valid")
	}
	if NormalizeName("  A  B ") != "a b" {
		t.Fatal("normalize")
	}
}
