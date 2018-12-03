import pandas as pd
import mapkit as mk
geohash = "0123456789bcdefghjkmnpqrstuvwxyz"


newlist = []
for i in geohash:
    for ii in geohash:
        for iii in geohash:
            newlist.append([i+ii+iii,'red'])
        newlist.append([i+ii,'green'])
    newlist.append([i,'blue'])
newlist.append(['7zz','black'])
newlist.append(['7z','white'])

data = pd.DataFrame(newlist,columns=['GEOHASH','COLORKEY'])
data['A'] = 100
#data['COLORKEY'].iloc[-1] = 'black'

print data
mk.make_blocks(data,'')