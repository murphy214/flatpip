import json


def read(filename):
    return open(filename).read()
base32 = "0123456789bcdefghjkmnpqrstuvwxyz"


xmap = {}
ymap = {}
mymap3 = {}
rows = read('second.txt').splitlines()
for y,row in enumerate(rows):
    for x,char in enumerate(row):
        print x << 2,x
        xmap[char] = x
        ymap[char] = y

print [xmap[i] for i in base32]
print [ymap[i] for i in base32]
print mymap