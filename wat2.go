package main 
import (
	"github.com/murphy214/tileclip"
	"fmt"
	"strconv"
	"time"
	m "github.com/murphy214/mercantile"
	"math"
	"math/rand"
	"github.com/paulmach/go.geo"
	"github.com/paulmach/go.geojson"
	"github.com/mmcloughlin/geohash"

)

// gets the extrema object of  a given geohash
func GetExtrema(ghash string) m.Extrema {
	box := geohash.BoundingBox(ghash)
	return m.Extrema{S: box.MinLat, W: box.MinLng, N: box.MaxLat, E: box.MaxLng}
}


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

const base32 = "0123456789bcdefghjkmnpqrstuvwxyz"
var stringmap = map[string]int{"j":17, "s":24, "v":27, "2":2, "4":4, "6":6, "t":25, "1":1, "f":14, "g":15, "n":20, "x":29, "8":8, "b":10, "e":13, "w":28, "0":0, "q":22, "r":23, "k":18, "p":21, "z":31, "5":5, "c":11, "h":16, "3":3, "7":7, "y":30, "9":9, "d":12, "m":19, "u":26}


// GeoHash returns the geohash string of a point representing a lng/lat location.
// The resulting hash will be `GeoHashPrecision` characters long, default is 12.
// Optionally one can include their required number of chars precision.
func (p *Point) GeoHash(chars ...int) (int,int) {
	pp := geo.Point{p[0],p[1]}
	hash := pp.GeoHash(chars[0])
	return GetGeohashXY(hash)
}

func GetIntFromGeohash(a string) int {
	total := 0
	for _,i := range a {
		total <<= 5
		total |= stringmap[string(i)]
	}
	return total
}

// gets xs,ys from hash
func GetGeohashXY(a string) (int,int) {
	total := GetIntFromGeohash(a)
	return GetGeohashXYInt(total)
}

func GetGeohashString(hash int,precision int) string {
	var result [15]byte

	for i := 1; i <= precision; i++ {
		result[precision-i] = byte(base32[hash&0x1F])
		hash >>= 5
	}

	return string(result[:precision])
}


// gets the geohash x from a given geohash
func GetGeohashXYInt(total int) (int, int) {

	var xs,ys string
	eh := fmt.Sprintf("%b",total)
	//eh = bytestring[:int(total%5)] + eh
	//fmt.Println(eh,len(eh),"eh")
	for pos,i := range eh {
		if pos%2 == 0 {
			xs += string(i)
		} else {
			ys += string(i)
		}
	}

	xval, _ := strconv.ParseInt(xs, 2,64)
	yval, _ := strconv.ParseInt(ys, 2,64)

	return int(xval),int(yval)
}

func SpinUp(xs,ys,precision int) int {
	var myval int64
	if precision%2!=0 {
		xss := fmt.Sprintf("%022b",xs)
		yss := fmt.Sprintf("%023b",ys)
		total := string(yss[0])
		yss = yss[1:]
		for i :=  range yss {
			total+=string(xss[i])+string(yss[i])
		}
		//total+= string(xss[len(xss)-1])
		myval,_ = strconv.ParseInt(total,2,64)
	} else {
		xss := fmt.Sprintf("%023b",xs)
		yss := fmt.Sprintf("%022b",ys)
		total := string(xss[0])
		xss = xss[1:]
		for i :=  range yss {
			total+=string(yss[i])+string(xss[i])
		}
		//total+= string(xss[len(xss)-1])
		myval,_ = strconv.ParseInt(total,2,64)
	}
	return int(myval)
} 



// the mapping structure
type Mapping struct {
	XDim int // the y dim size
	YDim int // the x dim size 
	BottomLeft [2]int // the integer positoni fo the bottom left geohash
	Bounds m.Extrema // bounds of the mapping
	Indexes []interface{}
	Size int
	Bytes []byte // how the mapping is structured
	IndexMap map[interface{}]int
}

func NewMapping(bds m.Extrema,size int) *Mapping {
	pt1 := Point{bds.W,bds.S}
	x,y := pt1.GeoHash(size)

	pt2 := Point{bds.E,bds.N}
	x2,y2 := pt2.GeoHash(size)
	xdim := int(math.Abs(float64(x2-x)))
	ydim := int(math.Abs(float64(y2-y)))

	
	newbs := make([]byte,xdim*ydim)
	return &Mapping{
		Size:size,
		Bounds:bds,
		BottomLeft:[2]int{x,y},
		XDim:xdim,
		YDim:ydim,
		Bytes:newbs,
		IndexMap:map[interface{}]int{},
	}

}

func (mapping *Mapping) RandomPt() []float64 {
	return []float64{
		rand.Float64() * (mapping.Bounds.E-mapping.Bounds.W) + mapping.Bounds.W,
		rand.Float64() * (mapping.Bounds.N-mapping.Bounds.S) + mapping.Bounds.S,
	}
}

func (mapping *Mapping) GetSetIndex(value interface{}) int {
	idx,boolval := mapping.IndexMap[value]
	if !boolval {
		idx = len(mapping.IndexMap)
		mapping.IndexMap[value] = idx
		mapping.Indexes = append(mapping.Indexes,value)
	}
	return idx+1
}


// sets an abaritary value to the mapping structure
func (mapping *Mapping) Set(geohash string,value interface{}) bool {
	x,y := GetGeohashXY(geohash) 
	//fmt.Println(mapping.BottomLeft[0],mapping.BottomLeft[1],mapping.BottomLeft[0]+mapping.XDim,mapping.BottomLeft[1]+mapping.YDim)
	xref,yref := x-mapping.BottomLeft[0],y-mapping.BottomLeft[1]
	// 
	//fmt.Println(x,y,xref,yref,"refs")
	//fmt.Printf("%v %v\n",mapping.XDim,mapping.YDim)
	idx := mapping.GetSetIndex(value)
	//fmt.Println((yref-1)*mapping.XDim+xref,x,y)
	//fmt.Println(xref,yref,len(mapping.Bytes),(yref-1)*mapping.XDim+xref)

	offset := (yref-1)*mapping.XDim+(xref-1)

	if offset < len(mapping.Bytes) && offset > 0 {
		mapping.Bytes[offset] = byte(idx)
	}
	return true
}

// gets a geohash from the given mapping 
func (mapping *Mapping) Get(geohash string) (interface{},bool) {
	x,y := GetGeohashXY(geohash) 
	//fmt.Println(mapping.BottomLeft[0],mapping.BottomLeft[1],mapping.BottomLeft[0]+mapping.XDim,mapping.BottomLeft[1]+mapping.YDim)
	xref,yref := x-mapping.BottomLeft[0],y-mapping.BottomLeft[1]

	offset := (yref-1)*mapping.XDim+(xref-1)
	if offset < len(mapping.Bytes) && offset > 0 {
		val := int(mapping.Bytes[offset])
		if val > 0 {
			return mapping.Indexes[val-1],true
		}
	}
	return "",false
}


// translates the entire geohash mapping to geohash features
func (mapping *Mapping) ToFeatures() []*geojson.Feature {
	feats := []*geojson.Feature{}
	for x := 1; x < mapping.XDim; x++ {
		for y := 1; y < mapping.YDim+1; y++ {
			
			offset := (y-1)*mapping.XDim+x


			val := mapping.Bytes[offset-1]
			if val > 0 {
				newval := mapping.Indexes[val-1]
				var mygeohash int

				mygeohash = SpinUp(mapping.BottomLeft[0]+x,mapping.BottomLeft[1]+y,mapping.Size)

				
				mygeohashstr := GetGeohashString(mygeohash,mapping.Size)

				bds := GetExtrema(mygeohashstr)
				newfeature := geojson.NewPolygonFeature(
					[][][]float64{{{bds.E, bds.N},{bds.E, bds.S},{bds.W, bds.S},{bds.W, bds.N},{bds.E, bds.N}}},
				)
				newfeature.Properties = map[string]interface{}{"X":x,"Y":y,"GEOHASH":mygeohashstr,"VALUE":newval}
				feats = append(feats,newfeature)


			}

		}
	}
	return feats
}





func main() {
	
	pt := &Point{90.0,-45}
	b,c := pt.GeoHash(4)
	ptt := &geo.Point{pt[0],pt[1]}
		
	a := ptt.GeoHash(4)
	bb,cc := GetGeohashXY(a)
	fmt.Println(b,c,bb,cc)
	
	tileid := m.TileID{35,49,7}
	
	mapping := NewMapping(m.Bounds(tileid), 9)
	ctr := m.Center(tileid)
	ptt = &geo.Point{ctr[0],ctr[1]}
	ghash := ptt.GeoHash(9)


	mapping.Set(ghash,"eh")
	s := time.Now()
	eh,_ := mapping.Get(ghash)
	fmt.Println(time.Now().Sub(s),eh)

	newmap := map[string]string{}
	
	for i := 0; i < 10000; i++ {
		pt := mapping.RandomPt()
		//fmt.Println(m.Tile(pt[0],pt[1],7))
		ptt := &geo.Point{pt[0],pt[1]}
		
		ghash := ptt.GeoHash(9)
		mapping.Set(ghash,"here")
		newmap[ghash] = "here"
	}
	feats := mapping.ToFeatures()

	for _,f := range feats {
		mystr := f.Properties["GEOHASH"].(string)
		_,boolval := newmap[mystr]
		if !boolval {
			
			fmt.Println(mystr,f.Properties)
			ghash := f.Properties["GEOHASH"].(string)
			xx,yy := GetGeohashXY(mystr)
			fmt.Printf("%b %b %b %d %d %d\n",GetIntFromGeohash(mystr),xx,yy,GetIntFromGeohash(mystr),xx,yy)
			ghash = f.Properties["GEOHASH"].(string)
			xx,yy = GetGeohashXY(ghash)
			fmt.Printf("%b %b %b %d %d %d new \n\n",GetIntFromGeohash(ghash),xx,yy,GetIntFromGeohash(ghash),xx,yy)

		
		} else {
			delete(newmap,mystr)
		}
	}
	for mystr := range newmap {
		xx,yy := GetGeohashXY(mystr)
		xx,yy = xx-mapping.BottomLeft[0],yy-mapping.BottomLeft[1]
		fmt.Printf("%b %b %b %d %d %d\n",GetIntFromGeohash(mystr),xx,yy,GetIntFromGeohash(mystr),xx,yy)

	}
	bds := m.Bounds(tileid)
	tfeat := geojson.NewPolygonFeature([][][]float64{{{bds.E, bds.N},{bds.E, bds.S},{bds.W, bds.S},{bds.W, bds.N},{bds.E, bds.N}}})
	tfeat.Properties = map[string]interface{}{"TOTAL BOUNDS":"HRE"}
	feats = append(feats,tfeat)
	tileclip.MakeFeatures(feats, "a.geojson")
	fmt.Println(len(mapping.Bytes)/1000/1000)
	s = time.Now()

	for i := 0; i < 10000; i++ {
		pt := mapping.RandomPt()
		//fmt.Println(m.Tile(pt[0],pt[1],7))
		mapping.Get(geohash.EncodeWithPrecision(pt[1],pt[0],uint(9)))

	}
	fmt.Println(time.Now().Sub(s).Seconds()/10000)

	//fmt.Println(mapping.XDim,mapping.YDim)
	//fmt.Println(len(mapping.Bytes))
	// 
}