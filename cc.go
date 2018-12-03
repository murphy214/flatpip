package main


import (
	"fmt"
)


type Point [2]float64

// GeoHashInt64 returns the integer version of the geohash
// down to the given number of bits.
// The main usecase for this function is to be able to do integer based ordering of points.
// In that case the number of bits should be the same for all encodings.
func (p *Point) GeoHashInt64(bits int) (hash int64) {
	// This code was inspired by https://github.com/broady/gogeohash
	
	latMin, latMax := -90.0, 90.0
	lngMin, lngMax := -180.0, 180.0

	for i := 0; i < bits; i++ {
		hash <<= 1

		// interleave bits
		if i%2 == 0 {
			mid := (lngMin + lngMax) / 2.0
			if p[0] > mid {
				lngMin = mid
				hash |= 1
			} else {
				lngMax = mid
			}
		} else {
			mid := (latMin + latMax) / 2.0
			if p[1] > mid {
				latMin = mid
				hash |= 1
			} else {
				latMax = mid
			}
		}
	}

	return 
}

func main() {

	pt := &Point{-0.0,0.0}
	pt.GeoHashInt64(2)
	fmt.Println(pt)
}