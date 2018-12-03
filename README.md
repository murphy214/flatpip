# flatpip

# Abstract

This set of files / hack represents some experiemntation with a completely flat point-in-polygon data structure. Which could potentially be ridicoulously quick and pretty small as well. It works by utilzing sort of a combination of the two things I've used for pip algorithms in the past, a defined tile structure to anchor the structure, and heirachal geohashing for super quick quadkeys. 

### Intendded Implementation 

When thinking about it, this really doesn't do anything THAT novel a lot of the concepts here are outlined in things like MapBox's UTF-8 grid but with a much different reason for the underlying implementation.

The backbone of this relies on geohash being a quadkey implementationl. Meaning you can get a defined resolution in the x,y in sub microsecond times, so from a geohash basically were going from geohash to min,max global x & global Y. This will give you something much like a tileid (global x,y) but derived much faster. Take note of the global here as its pretty important. So we have a tile were laying a grid over, we have a way of getting a global x,y from points the next step is to get the bounds of our tile in global geohash dimmensions. These two minimums will effectively be are   offset we subtract from any lookup.

```
    x         x
mx> 0         10 < max x
    x....|....x
100 x   105   x
    x         x
    x         x
    x         x
```
In the diagram below 100 & 105 would be globalx.

Getting the diesnon of our gridx size and gridy size is just the delta between min and max.

#### The Cool Part

So under a few assumptions are entire point in polygon index can be described as follows.

##### Creating the grid

First we create the shell grid taht will store are pip structure.

This grid should probably only be manipulated using the Get or Set method

```golang
grid := make([]byte,gridxsize*gridysize)
```

##### Getting an Index Position From a Point,Geohash, or geohash integer 

So this function gets the related index position given a point, geohash integer, and geohash string.. probably only one of these methods will ultimately be implemented but currenty needs all 3 for debugging purposes.

The name of the game here is stated earlier tap into a geohash enocode and ensure you get teh global x and global y out of the process or replace it with geohashing entirely. From this retured x,y we can get our position within our grid were using currently by subtracting from the min x & y. 

From getting these positions releative within our grid, we can do a simple calculation to get our offset to the byte that represents are geohash.

```golang
offset := (yref-1)*mapping.XDim+(xref-1)
representitive_byte := grid[offset]
```


#### How Feature Properties Are Stored

As you might have guessed feature properties are stored within that byte and are currently super limited in how many can be within a tile, 127. While often times this may be enough for a single tile I may end up changing the underlying implementation to use 2 bytes for each geohash if the values array gets larger than 128 etc. etc. All indexes are offset by 1 so that first 0 byte value is ALWAYS a no value. 

##### Processing for setting an item against the Structure

A map will have to be drapped along as your doing sets against the tile. While doing the set for a given pt,k,geohash you are also given an abitary interface{} to store with it. So basically the psudeocode for mananging the indicies,values_list,values_map looks like this.

```golang
test_value := "my polygon value"

func get_value_idx(value) int {
    values_map := map[interface{}]int{}
    values := []interface{}{}
    idx ,boolval := values_map[test_value]
    if !boolval {
        values_map[test_value] = len(values_map)
        values = append(values,test_value)
        idx = values_map[test_value]
    }
    return idx+1 // 1 bit offset to ensure that we know if a geohash was atleast hit
}
```


##### Getting a value int from a Raw Bytes

The code below shows the process for getting an abitary interface{} associated with a given polygon from the raw byte indexed from large bytes array representing the grid. This process is dead simple but the real code might have a lot of other sanity checks .

```golang
offset := (yref-1)*mapping.XDim+(xref-1)
representitive_byte := int(grid[offset])
feature_value := values[representitive_byte-1]
```

### Analaysis of Prototype

Currently I think this implementation could end of having a lot of value but there are tow main hang ups. One can I tap up into the fastest geohash libary and get the [x,y] bits out of each base32 char. (the shitty part is on the zoom it can either have a different number of offset bits based on precision level) 

i.e. Odd geohash levels have something like 3 x dim ors and 2 y or
i.e. even geohash levels have something like 3 y dim ors and 2 x or

Each reprenstiive of base 32 char is like 5 ors / offsets  all together. I need to ensure I can do this close to geohash speed. No BS converting the current representation to another where its convient, (currently code spawls everywhere to convert between geohash uint64, geohash xy, and geohash string) no this will be hot, easily the hottest piece of code nothing should be spared for that part.


#### Notes on complexity 

You may be thinking but storing a grid a high zoom will be a super huge grid won't that be massive? Yes, but it will be raw bytes that never need to be intialized. Furthermore in previous implementations I represented these grids as sparse maps (higher geohash if all lower geohashs were present in a given feature) with each key being generally like a 9 byte string and stored with a specific non-sparse value. 

Lets just assume these geohash below are a,b,c, & d (I know this isn't possible)

```
0 1 
1 0
```
So the oldway would look something like this.

```go
oldmap := map[string]interface{}{
    "a":0,
    "b":1,
    "c":1,
    "d":0,
}
```

The new way of representing the same thing is this.

```go
grid := []byte{0,1,1,0}
```

**I think its pretty safe to say that if all the things outlined above are correct than this may be a pretty good pip implementation for some use cases**
