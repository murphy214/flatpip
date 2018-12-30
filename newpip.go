package main


import (
	"github.com/paulmach/go.geojson"
	m "github.com/murphy214/mercantile"
	"github.com/mmcloughlin/geohash"
	"math/rand"
	"strconv"
	"math"
	"fmt"
	"time"
	"io/ioutil"
	
)

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



// Squash the even bitlevels of X into a 32-bit word. Odd bitlevels of X are
// ignored, and may take any value.
func squash(X uint64) uint32 {
	X &= 0x5555555555555555
	X = (X | (X >> 1)) & 0x3333333333333333
	X = (X | (X >> 2)) & 0x0f0f0f0f0f0f0f0f
	X = (X | (X >> 4)) & 0x00ff00ff00ff00ff
	X = (X | (X >> 8)) & 0x0000ffff0000ffff
	X = (X | (X >> 16)) & 0x00000000ffffffff
	return uint32(X)
}

// Deinterleave the bits of X into 32-bit words containing the even and odd
// bitlevels of X, respectively.
func deinterleave(X uint64,bits_right int) (int, int,int) {
	return int(squash(X)), int(squash(X >> 1)), int(X>>uint64(bits_right))
}

// Spread out the 32 bits of x into 64 bits, where the bits of x occupy even
// bit positions.
func spread(x uint32) uint64 {
	X := uint64(x)
	X = (X | (X << 16)) & 0x0000ffff0000ffff
	X = (X | (X << 8)) & 0x00ff00ff00ff00ff
	X = (X | (X << 4)) & 0x0f0f0f0f0f0f0f0f
	X = (X | (X << 2)) & 0x3333333333333333
	X = (X | (X << 1)) & 0x5555555555555555
	return X
}

// Interleave the bits of x and y. In the result, x and y occupy even and odd
// bitlevels, respectively.
func interleave(x, y uint32) uint64 {
	return spread(x) | (spread(y) << 1)
}

// geohash and returns the x and y values for a given data set
func GeohashXY(lat,lng float64,precison,precisiontile int) (int,int,int) {
	tval := geohash.EncodeIntWithPrecision(lat, lng, uint(precison*5))
	return deinterleave(tval,(precison-precisiontile)*5)
}


// the mapping structure
// fancy tiling structuer
type SubMapping struct {
	Bounds m.Extrema
	XDim int // the y dim size
	YDim int // the x dim size 
	BottomLeft [2]int // the integer positoni fo the bottom left geohash
	Indexes []interface{}
	Size int
	SubSize int
	Bytes []byte // how the mapping is structured
	IndexMap map[interface{}]int
}


// creating a new sub mapping in which this would be a tile
func NewSubMapping(subhash int,precison,precisiontile int) *SubMapping {
	box := geohash.BoundingBoxIntWithPrecision(uint64(subhash),uint(precisiontile*5))
	xmin,ymin,_ := GeohashXY(box.MinLat, box.MinLng, precison, precisiontile)
	xmax,ymax,_ := GeohashXY(box.MaxLat, box.MaxLng, precison, precisiontile)
	xdim := int(math.Abs(float64(xmax-xmin)))
	ydim := int(math.Abs(float64(ymax-ymin)))

	return &SubMapping{
		SubSize:precisiontile,
		Size:precison,
		BottomLeft:[2]int{xmin,ymin},
		XDim:xdim,
		YDim:ydim,
		Bytes:make([]byte,xdim*ydim),
		IndexMap:map[interface{}]int{},
		Bounds:m.Extrema{N:box.MaxLat,S:box.MinLat,E:box.MaxLng,W:box.MinLng},
	}
} 
const base32 = "0123456789bcdefghjkmnpqrstuvwxyz"

func GetGeohashString(hash int,precision int) string {
	var result [15]byte

	for i := 1; i <= precision; i++ {
		result[precision-i] = byte(base32[hash&0x1F])
		hash >>= 5
	}

	return string(result[:precision])
}

// generates a random point within the sub fiel mapping
func (mapping *SubMapping) RandomPt() []float64 {
	return []float64{
		rand.Float64() * (mapping.Bounds.E-mapping.Bounds.W) + mapping.Bounds.W,
		rand.Float64() * (mapping.Bounds.N-mapping.Bounds.S) + mapping.Bounds.S,
	}
}

// sets a value on the underlying subfile index
func (mapping *SubMapping) GetSetIndex(value interface{}) int {
	idx,boolval := mapping.IndexMap[value]
	if !boolval {
		idx = len(mapping.IndexMap)
		mapping.IndexMap[value] = idx
		mapping.Indexes = append(mapping.Indexes,value)
	}
	return idx+1
}

// sets an abaritary value to the mapping structure
func (mapping *SubMapping) Set(lat,lng float64,value interface{}) bool {
	x,y,_ := GeohashXY(lat,lng,mapping.Size,mapping.SubSize) 
	//fmt.Println(mapping.BottomLeft[0],mapping.BottomLeft[1],mapping.BottomLeft[0]+mapping.XDim,mapping.BottomLeft[1]+mapping.YDim)
	xref,yref := x-mapping.BottomLeft[0],y-mapping.BottomLeft[1]
	//fmt.Println(xref,yref)
	//fmt.Println(x,y,xref,yref,"refs")
	//fmt.Printf("%v %v\n",mapping.XDim,mapping.YDim)
	idx := mapping.GetSetIndex(value)
	//fmt.Println((yref-1)*mapping.XDim+xref,x,y)
	//fmt.Println(xref,yref,len(mapping.Bytes),(yref-1)*mapping.XDim+xref)
	//fmt.Println(idx)
	offset := (yref-1)*mapping.XDim+(xref-1)
	if offset < len(mapping.Bytes) && offset > 0 {
		mapping.Bytes[offset] = byte(idx)
	}
	return true
}

// gets a geohash from the given mapping 
func (mapping *SubMapping) Get(lat,lng float64) (interface{},bool) {
	x,y,_ := GeohashXY(lat,lng,mapping.Size,mapping.SubSize) 
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
func (mapping *SubMapping) ToFeatures() []*geojson.Feature {
	feats := []*geojson.Feature{}
	for x := 1; x <= mapping.XDim; x++ {
		for y := 1; y <= mapping.YDim; y++ {
			
			offset := (y-1)*mapping.XDim+(x-1)

			val := mapping.Bytes[offset]
			
			//fmt.Println(val,x,y)
			if val > 0 {
				newval := mapping.Indexes[val-1]

				xx,yy := mapping.BottomLeft[0]+x,mapping.BottomLeft[1]+y
				val := interleave(uint32(xx), uint32(yy))
				box := geohash.BoundingBoxIntWithPrecision(val,uint(mapping.Size*5))
				bds := m.Extrema{N:box.MaxLat,S:box.MinLat,E:box.MaxLng,W:box.MinLng}
				//fmt.Println(bds)
				newfeature := geojson.NewPolygonFeature(
					[][][]float64{{{bds.E, bds.N},{bds.E, bds.S},{bds.W, bds.S},{bds.W, bds.N},{bds.E, bds.N}}},
				)
				newfeature.Properties = map[string]interface{}{"X":x,"Y":y,"GEOHASH":GetGeohashString(int(val), mapping.Size),"VALUE":newval}
				feats = append(feats,newfeature)


			}

		}
	}
	return feats
}





func main() {
	lat,lng := 38.0,-100.0
	precisiontile := 3
	_,_,subhash := GeohashXY(lat, lng, 9, precisiontile)
	submap := NewSubMapping(subhash,9,precisiontile)
	
	newlist := []*geojson.Feature{}
	// 
	for i := 0; i < 10000000; i++ {
		
		pt := submap.RandomPt()
		
		submap.Set(pt[1],pt[0],"mykey")

	}


	pt := submap.RandomPt()
	submap.Set(pt[1],pt[0],"check")
	val,boolval := submap.Get(pt[1],pt[0])
	fmt.Println(val,boolval)


	ss := time.Now()
	num := 100000000
	for i := 0; i < num; i++ {
		pt := submap.RandomPt()
		submap.Get(pt[1],pt[0])
	}
	fmt.Printf("Took %v to query %d million points\n",time.Now().Sub(ss),num/1000000)


	//fmt.Println(submap.XDim,submap.YDim)
	//fmt.Println(float64(len(submap.Bytes))/float64(1000*1000*1000))
	//newlist = append(newlist,submap.ToFeatures()...)



	fc := &geojson.FeatureCollection{Features:newlist}
	s,_ := fc.MarshalJSON()
	ioutil.WriteFile("a.geojson",[]byte(s),0677)



	/*
	s := time.Now()
	GeohashXY(22.223423,0,12,3)
	fmt.Println(time.Now().Sub(s))

	s = time.Now()
	m.Tile(0,22.223423,12)
	fmt.Println(time.Now().Sub(s))
	*/
}